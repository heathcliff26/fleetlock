package mongodb

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/heathcliff26/fleetlock/pkg/lock-manager/types"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const DEFAULT_DATABASE = "fleetlock"

type MongoDBBackend struct {
	client   *mongo.Client
	database string
}

type MongoDBConfig struct {
	URL      string `yaml:"url,omitempty"`
	Database string `yaml:"database,omitempty"`
}

type MongoLock struct {
	ID      string    `bson:"_id,omitempty"`
	Created time.Time `bson:"created,omitempty"`
}

func NewMongoDBBackend(cfg MongoDBConfig) (*MongoDBBackend, error) {
	if cfg.Database == "" {
		cfg.Database = DEFAULT_DATABASE
	}

	timeout := time.Second * 5

	opts := options.Client()
	opts.ConnectTimeout = &timeout
	opts.Timeout = &timeout
	opts.ApplyURI(cfg.URL)

	c, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create mongodb client: %v", err)
	}

	err = c.Ping(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %v", err)
	}

	slog.Debug("Opened connection to mongodb", slog.String("database", cfg.Database))

	return &MongoDBBackend{
		client:   c,
		database: cfg.Database,
	}, nil
}

// Reserve a lock for the given group.
// Returns true if the lock is successfully reserved, even if the lock is already held by the specific id
func (m *MongoDBBackend) Reserve(group string, id string) error {
	coll := m.client.Database(m.database).Collection(group)

	newObj := MongoLock{
		ID:      id,
		Created: time.Now(),
	}

	_, err := coll.InsertOne(context.Background(), newObj)
	return err
}

// Returns the current number of locks for the given group
func (m *MongoDBBackend) GetLocks(group string) (int, error) {
	coll := m.client.Database(m.database).Collection(group)
	count, err := coll.CountDocuments(context.Background(), MongoLock{})
	return int(count), err
}

// Release the lock currently held by the id.
// Does not fail when no lock is held.
func (m *MongoDBBackend) Release(group string, id string) error {
	coll := m.client.Database(m.database).Collection(group)

	filter := MongoLock{
		ID: id,
	}
	_, err := coll.DeleteOne(context.Background(), filter)

	return err
}

// Return all locks older than x
func (m *MongoDBBackend) GetStaleLocks(ts time.Duration) ([]types.Lock, error) {
	panic("not implemented") // TODO: Implement
}

// Check if a given id already has a lock for this group
func (m *MongoDBBackend) HasLock(group string, id string) (bool, error) {
	coll := m.client.Database(m.database).Collection(group)

	filter := MongoLock{
		ID: id,
	}
	res := coll.FindOne(context.Background(), filter)

	switch res.Err() {
	case mongo.ErrNoDocuments:
		return false, nil
	case nil:
		return true, nil
	default:
		return false, res.Err()
	}
}

// Calls all necessary finalization if necessary
func (m *MongoDBBackend) Close() error {
	return m.client.Disconnect(context.Background())
}
