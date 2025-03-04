package error

import "time"

type TimestampInterface interface {
	GetTimestamp() time.Time
	SetTimestamp(time.Time)
}
