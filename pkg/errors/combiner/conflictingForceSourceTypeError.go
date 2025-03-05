package combiner

import (
	"fmt"
	"time"
)

type ConflictingForceSourceTypeError struct {
	Timestamp int64
	Path      string
	OldType   string
	NewType   string
}

func (e *ConflictingForceSourceTypeError) SetTimestamp(t time.Time) {
	e.Timestamp = t.UnixNano()
}

func (e *ConflictingForceSourceTypeError) GetTimestamp() time.Time {
	return time.Unix(e.Timestamp, 0)
}

func (e *ConflictingForceSourceTypeError) Error() string {
	return fmt.Sprintf("could not combine two fields at the path: %s whit different ForceSourceType's. "+
		"Old field has: %s, new field has %s.", e.Path, e.OldType, e.NewType)
}
