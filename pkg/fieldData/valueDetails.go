package fieldData

type ValueDetails struct {
	Count              int   `json:"count,omitempty"`
	FirstSeenTimestamp int64 `json:"firstSeenTimestamp,omitempty"`
	LastSeenTimestamp  int64 `json:"lastSeenTimestamp,omitempty"`
}

func (v *ValueDetails) Combine(v1 *ValueDetails) *ValueDetails {
	var combinedValueDetails ValueDetails
	//Combine Count
	combinedValueDetails.Count = v.Count + v1.Count
	//Combine FirstSeenTimestamp
	if v.FirstSeenTimestamp < v1.FirstSeenTimestamp {
		combinedValueDetails.FirstSeenTimestamp = v.FirstSeenTimestamp
	} else {
		combinedValueDetails.FirstSeenTimestamp = v1.FirstSeenTimestamp
	}
	// Combine LastSeenTimestamp
	if v.LastSeenTimestamp > v1.LastSeenTimestamp {
		combinedValueDetails.LastSeenTimestamp = v.LastSeenTimestamp
	} else {
		combinedValueDetails.LastSeenTimestamp = v1.LastSeenTimestamp
	}
	return &combinedValueDetails
}
