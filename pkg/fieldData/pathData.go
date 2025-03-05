package fieldData

import (
	"encoding/json"
	"errors"
	j2gErrors "github.com/Lemonn/JSON2Go/pkg/errors"
)

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

func (p *PathData) UnmarshalJSON(bytes []byte) error {
	localType := struct {
		Types                   map[Type]map[int]map[string]*ValueDetails `json:"types,omitempty"`
		JsonFieldName           string                                    `json:"jsonFieldName,omitempty"`
		ActiveType              *string                                   `json:"activeType,omitempty"`
		TypeAdjusterData        *TypeAdjusterData                         `json:"typeAdjusterData,omitempty"`
		Error                   []*j2gErrors.StoreType                    `json:"error,omitempty"`
		ForceSourceType         *string                                   `json:"forceSourceType,omitempty"`
		DirectToForceSourceType bool                                      `json:"directToForceSourceType,omitempty"`
		IntroductionCount       int                                       `json:"introductionCount,omitempty"`
		FirstSeenTimestamp      int64                                     `json:"firstSeenTimestamp,omitempty"`
		LastSeenTimestamp       int64                                     `json:"lastSeenTimestamp,omitempty"`
		SeenCounter             int                                       `json:"seenCounter,omitempty"`
		Active                  bool                                      `json:"active,omitempty"`
		FieldComment            string                                    `json:"fieldComment,omitempty"`
		ForcePointerType        bool                                      `json:"forcePointerType,omitempty"`
		ForcedImports           []string                                  `json:"forcedImports,omitempty"`
		RequiredField           bool                                      `json:"requiredField,omitempty"`
		ForceOmitempty          *bool                                     `json:"forceOmitempty,omitempty"`
	}{}
	err := json.Unmarshal(bytes, &localType)
	if err != nil {
		return err
	}
	for _, storeType := range localType.Error {
		err = errors.Join(err, storeType.Err)
	}
	*p = PathData{
		Types:                   localType.Types,
		JsonFieldName:           localType.JsonFieldName,
		ActiveType:              localType.ActiveType,
		TypeAdjusterData:        localType.TypeAdjusterData,
		Error:                   err,
		ForceSourceType:         localType.ForceSourceType,
		DirectToForceSourceType: localType.DirectToForceSourceType,
		IntroductionCount:       localType.IntroductionCount,
		FirstSeenTimestamp:      localType.FirstSeenTimestamp,
		LastSeenTimestamp:       localType.LastSeenTimestamp,
		SeenCounter:             localType.SeenCounter,
		Active:                  localType.Active,
		FieldComment:            localType.FieldComment,
		ForcePointerType:        localType.ForcePointerType,
		ForcedImports:           localType.ForcedImports,
		RequiredField:           localType.RequiredField,
		ForceOmitempty:          localType.ForceOmitempty,
	}
	return nil
}

func (p *PathData) MarshalJSON() ([]byte, error) {
	localType := struct {
		Types                   map[Type]map[int]map[string]*ValueDetails `json:"types,omitempty"`
		JsonFieldName           string                                    `json:"jsonFieldName,omitempty"`
		ActiveType              *string                                   `json:"activeType,omitempty"`
		TypeAdjusterData        *TypeAdjusterData                         `json:"typeAdjusterData,omitempty"`
		Error                   []*j2gErrors.StoreType                    `json:"error,omitempty"`
		ForceSourceType         *string                                   `json:"forceSourceType,omitempty"`
		DirectToForceSourceType bool                                      `json:"directToForceSourceType,omitempty"`
		IntroductionCount       int                                       `json:"introductionCount,omitempty"`
		FirstSeenTimestamp      int64                                     `json:"firstSeenTimestamp,omitempty"`
		LastSeenTimestamp       int64                                     `json:"lastSeenTimestamp,omitempty"`
		SeenCounter             int                                       `json:"seenCounter,omitempty"`
		Active                  bool                                      `json:"active,omitempty"`
		FieldComment            string                                    `json:"fieldComment,omitempty"`
		ForcePointerType        bool                                      `json:"forcePointerType,omitempty"`
		ForcedImports           []string                                  `json:"forcedImports,omitempty"`
		RequiredField           bool                                      `json:"requiredField,omitempty"`
		ForceOmitempty          *bool                                     `json:"forceOmitempty,omitempty"`
	}{Types: p.Types, JsonFieldName: p.JsonFieldName, ActiveType: p.ActiveType, TypeAdjusterData: p.TypeAdjusterData,
		Error: j2gErrors.NewStoreTypeArray(p.Error), ForceSourceType: p.ForceSourceType, DirectToForceSourceType: p.DirectToForceSourceType,
		IntroductionCount: p.IntroductionCount, FirstSeenTimestamp: p.FirstSeenTimestamp,
		LastSeenTimestamp: p.LastSeenTimestamp, SeenCounter: p.SeenCounter, Active: p.Active,
		FieldComment: p.FieldComment, ForcePointerType: p.ForcePointerType, ForcedImports: p.ForcedImports,
		RequiredField: p.RequiredField, ForceOmitempty: p.ForceOmitempty}
	return json.Marshal(localType)
}

type FileData map[string]*PathData
