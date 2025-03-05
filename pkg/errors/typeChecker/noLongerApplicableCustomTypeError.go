package typeChecker

import (
	"time"
)

type NoLongerApplicableCustomTypeError struct {
	Timestamp         int64  `json:"timestamp"`
	ActiveTypeChecker string `json:"activeTypeChecker"`
}

/*
func (e *NoLongerApplicableCustomTypeError) UnmarshalJSON(bytes []byte) error {
	localType := struct {
		Timestamp         int64  `json:"timestamp"`
		ActiveTypeChecker string `json:"activeTypeChecker"`
		Type              string `json:"type"`
	}{}
	err := json.Unmarshal(bytes, &localType)
	if err != nil {
		return err
	}
	if localType.Type != reflect.TypeOf(e).String() {
		return &errors.ErrorTypeMismatchError{
			JsonType:  localType.Type,
			ErrorType: reflect.TypeOf(e).String(),
		}
	}
	return nil
}

func (e *NoLongerApplicableCustomTypeError) MarshalJSON() ([]byte, error) {
	localType := struct {
		Timestamp         int64  `json:"timestamp"`
		ActiveTypeChecker string `json:"activeTypeChecker"`
		Type              string `json:"type"`
	}{
		Timestamp:         e.Timestamp,
		ActiveTypeChecker: e.ActiveTypeChecker,
		Type:              reflect.TypeOf(e).String(),
	}
	return json.Marshal(localType)
}

*/

func (e *NoLongerApplicableCustomTypeError) SetTimestamp(t time.Time) {
	e.Timestamp = t.Unix()
}

func (e *NoLongerApplicableCustomTypeError) GetTimestamp() time.Time {
	return time.Unix(e.Timestamp, 0)
}

func (e *NoLongerApplicableCustomTypeError) IsFileError() {}

func (e *NoLongerApplicableCustomTypeError) Error() string {
	return ""
}
