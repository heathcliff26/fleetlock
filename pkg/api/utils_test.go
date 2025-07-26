package api

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRequest(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    FleetLockRequest
		wantErr bool
	}{
		{
			name: "ValidRequest",
			body: `{"client_params":{"id":"test-id","group":"test-group"}}`,
			want: FleetLockRequest{
				Client: FleetLockRequestClient{
					ID:    "test-id",
					Group: "test-group",
				},
			},
			wantErr: false,
		},
		{
			name:    "InvalidJSON",
			body:    `{"client_params":{"id":"test-id","group":}`,
			want:    FleetLockRequest{},
			wantErr: true,
		},
		{
			name: "EmptyBody",
			body: `{}`,
			want: FleetLockRequest{
				Client: FleetLockRequestClient{
					ID:    "",
					Group: "",
				},
			},
			wantErr: false,
		},
		{
			name: "PartialData",
			body: `{"client_params":{"id":"test-id"}}`,
			want: FleetLockRequest{
				Client: FleetLockRequestClient{
					ID:    "test-id",
					Group: "",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := io.NopCloser(strings.NewReader(tt.body))
			got, err := ParseRequest(body)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    FleetLockResponse
		wantErr bool
	}{
		{
			name: "ValidResponse",
			body: `{"kind":"success","value":"operation completed"}`,
			want: FleetLockResponse{
				Kind:  "success",
				Value: "operation completed",
			},
			wantErr: false,
		},
		{
			name:    "InvalidJSON",
			body:    `{"kind":"success","value":}`,
			want:    FleetLockResponse{},
			wantErr: true,
		},
		{
			name: "EmptyBody",
			body: `{}`,
			want: FleetLockResponse{
				Kind:  "",
				Value: "",
			},
			wantErr: false,
		},
		{
			name: "ErrorResponse",
			body: `{"kind":"error","value":"something went wrong"}`,
			want: FleetLockResponse{
				Kind:  "error",
				Value: "something went wrong",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := io.NopCloser(strings.NewReader(tt.body))
			got, err := ParseResponse(body)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestPrepareRequest(t *testing.T) {
	tests := []struct {
		name    string
		group   string
		id      string
		wantErr bool
	}{
		{
			name:    "ValidRequest",
			group:   "test-group",
			id:      "test-id",
			wantErr: false,
		},
		{
			name:    "EmptyGroup",
			group:   "",
			id:      "test-id",
			wantErr: false,
		},
		{
			name:    "EmptyID",
			group:   "test-group",
			id:      "",
			wantErr: false,
		},
		{
			name:    "BothEmpty",
			group:   "",
			id:      "",
			wantErr: false,
		},
		{
			name:    "SpecialCharacters",
			group:   "test-group@#$%",
			id:      "test-id!@#$%^&*()",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PrepareRequest(tt.group, tt.id)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, got)
				
				// Read the body and verify it can be parsed
				body, err := io.ReadAll(got)
				require.NoError(t, err)
				
				var req FleetLockRequest
				err = json.Unmarshal(body, &req)
				require.NoError(t, err)
				
				assert.Equal(t, tt.id, req.Client.ID)
				assert.Equal(t, tt.group, req.Client.Group)
			}
		})
	}
}

func TestPrepareRequestRoundTrip(t *testing.T) {
	// Test that PrepareRequest and ParseRequest work together
	group := "test-group"
	id := "test-id"
	
	// Create request
	reader, err := PrepareRequest(group, id)
	require.NoError(t, err)
	
	// Convert to ReadCloser for ParseRequest
	body, err := io.ReadAll(reader)
	require.NoError(t, err)
	readCloser := io.NopCloser(bytes.NewReader(body))
	
	// Parse the request
	parsed, err := ParseRequest(readCloser)
	require.NoError(t, err)
	
	// Verify the round trip worked
	assert.Equal(t, id, parsed.Client.ID)
	assert.Equal(t, group, parsed.Client.Group)
}