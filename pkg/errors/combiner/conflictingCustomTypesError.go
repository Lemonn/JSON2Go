package combiner

import (
	"fmt"
	"time"
)

// ConflictingCustomTypesError is emitted whenever two files are combined, which have conflicting TypeAdjuster's
type ConflictingCustomTypesError struct {
	Timestamp int64 `json:"timestamp"`
	// OldTypeAdjuster Is the TypeAdjuster which has been set first, according to the set timestamp.
	// If both timestamps are equal, the order is nondeterministic
	OldTypeAdjuster string `json:"oldTypeAdjuster"`
	// OldTypeAdjuster Is the TypeAdjuster which has been set last, according to the set timestamp.
	// If both timestamps are equal, the order is nondeterministic
	NewTypeAdjuster string `json:"newTypeAdjuster"`
}

func (e *ConflictingCustomTypesError) SetTimestamp(t time.Time) {
	e.Timestamp = t.UnixNano()
}

func (e *ConflictingCustomTypesError) GetTimestamp() time.Time {
	return time.Unix(e.Timestamp, 0)
}

func (e *ConflictingCustomTypesError) IsFileError() {}

func (e *ConflictingCustomTypesError) Error() string {
	return fmt.Sprintf("conflicting custom types, old file uses %s and new file uses %s", e.OldTypeAdjuster, e.NewTypeAdjuster)
}
