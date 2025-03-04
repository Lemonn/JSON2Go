package typeFile

import (
	"errors"
	timestampError "github.com/Lemonn/JSON2Go/internal/error"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/codeGenerators"
	j2gErrors "github.com/Lemonn/JSON2Go/pkg/errors"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"time"
)

type Combiner struct {
	startTime         time.Time
	fileDetails       []*FileDetails
	codeGenerator     codeGenerators.CodeGenerator
	laterCheck        []string
	combinedSeenTypes map[string]*fieldData.PathData
}

type FileDetails struct {
	creationDate  time.Time
	seenTypes     map[string]*fieldData.PathData
	seenTypeUtils *utils.SeenTypeUtils
}

func (c *Combiner) getOldDetails() *FileDetails {
	return nil
}

func (c *Combiner) getNewDetails() *FileDetails {
	return nil
}

func (c *Combiner) InterfaceReplacement(p, p1 *fieldData.PathData) bool {
	return false
}

func (c *Combiner) processStackedTypeChecks() error {
	c.codeGenerator.SetActiveTypeFile(c.combinedSeenTypes)
	for _, path := range c.laterCheck {
		err := c.codeGenerator.CheckType(path)
		err = c.appendFileErrorsToTypeFile(err, path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Combiner) Combine(p, p1 *fieldData.PathData, path string) (*fieldData.PathData, error) {
	newP := fieldData.PathData{Types: make(map[fieldData.Type]map[int]map[string]*fieldData.ValueDetails)}

	//Combine Types
	for t, arrayLevels := range p.Types {
		if _, ok := newP.Types[t]; !ok {
			newP.Types[t] = make(map[int]map[string]*fieldData.ValueDetails)
		}
		for level, valueDetails := range arrayLevels {
			if _, ok := newP.Types[t][level]; !ok {
				newP.Types[t][level] = map[string]*fieldData.ValueDetails{}
			}
			for value, valueDetail := range valueDetails {
				if _, ok := newP.Types[t][level][value]; !ok {
					newP.Types[t][level][value] = valueDetail
				} else {
					newP.Types[t][level][value].Combine(valueDetail)
				}
			}

		}
	}
	for t, arrayLevels := range p1.Types {
		if _, ok := newP.Types[t]; !ok {
			newP.Types[t] = make(map[int]map[string]*fieldData.ValueDetails)
		}
		for level, valueDetails := range arrayLevels {
			if _, ok := newP.Types[t][level]; !ok {
				newP.Types[t][level] = map[string]*fieldData.ValueDetails{}
			}
			for value, valueDetail := range valueDetails {
				if _, ok := newP.Types[t][level][value]; !ok {
					newP.Types[t][level][value] = valueDetail
				} else {
					newP.Types[t][level][value].Combine(valueDetail)
				}
			}
		}
	}

	// Only care about type changes, if the type is active.
	if p.Active || p1.Active {
		// Combine the TypeAdjusterData
		var combine *fieldData.TypeAdjusterData
		var err error
		if p.TypeAdjusterData != nil {
			combine, err = p.TypeAdjusterData.Combine(p1.TypeAdjusterData)

		} else if p1.TypeAdjusterData != nil {
			combine, err = p1.TypeAdjusterData.Combine(p.TypeAdjusterData)
		}

		if combine != nil || err != nil {
			hardError := c.appendFileErrorsToTypeFile(err, path)
			if hardError != nil {
				return nil, hardError
			}
			newP.TypeAdjusterData = combine
			if err == nil {
				c.laterCheck = append(c.laterCheck, path)
			}
		}

		// Combine the ActiveType
		if p.ActiveType == nil {
			newP.ActiveType = p1.ActiveType
		} else if p1.ActiveType == nil {
			newP.ActiveType = p.ActiveType
		} else if p.ActiveType == p1.ActiveType {
			newP.ActiveType = p.ActiveType
		} else if !(p.TypeAdjusterData != nil && p.TypeAdjusterData.NameOfActiveTypeAdjuster != nil ||
			p1.TypeAdjusterData != nil && p1.TypeAdjusterData.NameOfActiveTypeAdjuster != nil) {
			//TODO we need to set an type change error
			newP.ActiveType = nil
		}
	} else {
		newP.ActiveType = nil
	}

	//Combine JsonFieldName
	if p.JsonFieldName != p1.JsonFieldName {
		// TODO respect the conflict handler
		return nil, &j2gErrors.ConflictingJsonFieldNameError{
			Timestamp:    c.startTime.Unix(),
			OldFieldName: p.JsonFieldName,
			NewFieldName: p1.JsonFieldName,
		}
	} else {
		newP.JsonFieldName = p.JsonFieldName
	}

	//Combine Error
	newP.Error = errors.Join(p.Error, p1.Error)

	//Combine RequiredField
	if p.RequiredField || p1.RequiredField {
		newP.RequiredField = true
	}

	//Combine ForceSourceType
	if p.ForceSourceType != nil && p1.ForceSourceType == nil {
		newP.ForceSourceType = p.ForceSourceType
	} else if p.ForceSourceType == nil && p1.ForceSourceType != nil {
		newP.ForceSourceType = p1.ForceSourceType
	} else if p.ForceSourceType != nil && p1.ForceSourceType != nil {
		//TODO respect conflict resolving strategic
		if *p1.ForceSourceType != *p.ForceSourceType {
			return nil, &j2gErrors.ConflictingForceSourceTypeError{}
		} else {
			newP.ForceSourceType = p.ForceSourceType
		}
	}

	if p.DirectToForceSourceType != p1.DirectToForceSourceType {
		//TODO respect conflict resolving strategic
		return nil, &j2gErrors.ConflictingForceSourceTypeError{}
	} else {
		newP.DirectToForceSourceType = p.DirectToForceSourceType
	}

	// Combine IntroductionCount/SeenCounter
	// We need to make some assumptions here.
	// 1. If the newer file does not contain the value, we assume its omitted or does no longer exists.
	// Therefore, we use the values of the old file
	// 2. If the older file does not contain the value, we assume it has been added later. Therefore, the old parent
	// SeenCounter is used as additional offset, to the IntroductionCount of the newer file.
	// 3. If the value is present in both files, we simply combine the IntroductionCount and SeenCounter with
	// their counterparts.
	if p1.IntroductionCount == 0 && p1.SeenCounter == 0 {
		newP.IntroductionCount = p.IntroductionCount
		newP.SeenCounter = p.SeenCounter
	} else if p1.IntroductionCount == 0 && p1.SeenCounter == 0 {
		oldDetails := c.getOldDetails()
		var seenCounterOffset int
		//Check if old file contains the parent type, if so add the value as offset.
		if _, ok := oldDetails.seenTypes[oldDetails.seenTypeUtils.GetParentPath(path)]; ok {
			seenCounterOffset = oldDetails.seenTypes[oldDetails.seenTypeUtils.GetParentPath(path)].SeenCounter
		}
		newP.SeenCounter = p1.SeenCounter + seenCounterOffset
	} else {
		newP.IntroductionCount = p.IntroductionCount + p1.IntroductionCount
		newP.SeenCounter = p.SeenCounter + p1.SeenCounter
	}

	//Combine FirstSeenTimestamp
	if p.FirstSeenTimestamp > p1.FirstSeenTimestamp {
		newP.FirstSeenTimestamp = p1.FirstSeenTimestamp
	} else {
		newP.FirstSeenTimestamp = p.FirstSeenTimestamp
	}

	//Combine LastSeenTimestamp
	if p.FirstSeenTimestamp < p1.FirstSeenTimestamp {
		newP.FirstSeenTimestamp = p1.FirstSeenTimestamp
	} else {
		newP.FirstSeenTimestamp = p.FirstSeenTimestamp
	}

	//Combine Active
	if p.Active || p1.Active {
		newP.Active = true
	}

	//TODO combine user settings

	return &newP, nil
}

func (c *Combiner) appendFileErrorsToTypeFile(err error, path string) error {
	if err == nil {
		return nil
	}
	ea := utils.GetAllWrappedErrors(err)
	for _, ue := range ea {
		if _, ok := ue.(timestampError.FileError); !ok {
			return ue
		}
		if v, ok := ue.(timestampError.TimestampInterface); ok {
			v.SetTimestamp(c.startTime)
		}
	}
	for _, ue := range ea {
		c.combinedSeenTypes[path].Error = errors.Join(c.combinedSeenTypes[path].Error, ue)
	}
	return nil
}
