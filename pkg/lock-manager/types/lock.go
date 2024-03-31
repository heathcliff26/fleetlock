package types

import "time"

type Lock struct {
	Group, ID string
	Created   time.Time
}
