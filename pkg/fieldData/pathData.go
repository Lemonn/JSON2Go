package fieldData

type Metadata struct {
	TotalSampleCount int              `json:"totalSampleCount"`
	LastRunTimestamp int64            `json:"lastRunTimestamp"`
	GeneratorData    []*GeneratorData `json:"generatorData"`
	Data             FileData         `json:"fileData"`
}

type GeneratorData struct {
	ActiveTypeCheckers []TypeCheckerDetails `json:"activeTypeCheckers"`
	GeneratorRuntime   int64                `json:"generatorRuntime"`
	Name               string               `json:"name"`
	Version            string               `json:"version"`
}

type TypeCheckerDetails struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type PathData struct {
	Types map[Type]map[int]map[string]*ValueDetails `json:"types,omitempty"`
	// Contains the name, as found in the JSON-File
	JsonFieldName string `json:"jsonFieldName,omitempty"`
	// ActiveType represents the goland code string of the currently active type
	ActiveType *string `json:"activeType,omitempty"`

	TypeAdjusterData *TypeAdjusterData `json:"typeAdjusterData,omitempty"`
	// Error Holds an error of type FieldPresenceChangeError, errors.IncompatibleCustomTypeError or errors.TypeChangeError.
	// It's up to the caller, what to do with this. errors.IncompatibleCustomTypeError or TypeChangeError
	// indicate a major version change. Should be set to nil, before next run!
	Error error `json:"error,omitempty"`

	// ForceSourceType is set, a generator ignores the type determined by the values and uses this type instead
	ForceSourceType *string `json:"forceSourceType,omitempty"`
	// DirectToForceSourceType Set if the ForceSourceType implements its own unmarshall function and therefore
	// does not need an intermediate step
	DirectToForceSourceType bool  `json:"directToForceSourceType,omitempty"`
	IntroductionCount       int   `json:"introductionCount,omitempty"`
	FirstSeenTimestamp      int64 `json:"firstSeenTimestamp,omitempty"`
	// LastSeenTimestamp Unix timestamp of the last time a value was seen. Is used to remove stale values and or types
	LastSeenTimestamp int64 `json:"lastSeenTimestamp,omitempty"`
	// SeenCounter Represents the total amount a type has been seen.
	SeenCounter int `json:"seenCounter,omitempty"`
	// Active is set, when the type is part of the last generated version.
	Active bool `json:"active,omitempty"`

	//User Settings. Changing them. whilst the type is active could result in a breaking change, which in turn results
	// in a new version.

	// FieldComment Comment that should be added to the field. If automatic comments are present, this
	// one is attached at the end
	FieldComment     string   `json:"fieldComment,omitempty"`
	ForcePointerType bool     `json:"forcePointerType,omitempty"`
	ForcedImports    []string `json:"forcedImports,omitempty"`
	// RequiredField set to true, if the unmarshall generator should make this a required field.
	RequiredField bool `json:"requiredField,omitempty"`
	// ForceOmitempty Set whenever a filed should be forced as omitempty, regardless of the automatic determination.
	ForceOmitempty *bool `json:"forceOmitempty,omitempty"`
}

type FileData map[string]*PathData
