package typeFile

import (
	"errors"
	"github.com/Lemonn/JSON2Go/internal/utils"
	j2gErrors "github.com/Lemonn/JSON2Go/pkg/errors"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"time"
)

type Combiner struct {
	startTime   time.Time
	fileDetails []*FileDetails
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

func (c *Combiner) Combine(p, p1 *fieldData.PathData, path string) (*fieldData.PathData, error) {
	if p.FirstSeenTimestamp > p1.FirstSeenTimestamp {
		//TODO sort
	}

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
		if p.ActiveType == nil {
			newP.ActiveType = p1.ActiveType
		} else if p1.ActiveType == nil {
			newP.ActiveType = p.ActiveType
		} else if p.ActiveType == p1.ActiveType {
			newP.ActiveType = p.ActiveType
		} else {

			//TODO we need to run the TypeChecker on the new file as well, if one is active at the old field

			if p.TypeAdjusterData != nil && p.TypeAdjusterData.NameOfActiveTypeAdjuster != nil &&
				p1.TypeAdjusterData == nil || (p1.TypeAdjusterData != nil && p1.TypeAdjusterData.NameOfActiveTypeAdjuster == nil) {
				//TODO run checker on p and see if it matches the type of p1
			} else if p1.TypeAdjusterData != nil && p1.TypeAdjusterData.NameOfActiveTypeAdjuster != nil &&
				p.TypeAdjusterData == nil || (p.TypeAdjusterData != nil && p.TypeAdjusterData.NameOfActiveTypeAdjuster == nil) {
				//TODO run checker on p1 and see if it matches the type of p
			} else if p1.TypeAdjusterData == nil || p1.TypeAdjusterData != nil && p1.TypeAdjusterData.NameOfActiveTypeAdjuster == nil {
				//TODO reset all active checker status and types and set to default type based on the values
				// Or set no type at all, as this can be done by the generator
			}

			newP.ActiveType = nil
			newP.Error = errors.Join(newP.Error, &j2gErrors.TypeChangeError{
				Timestamp:                0,
				OldType:                  *p.ActiveType,
				NewType:                  *p1.ActiveType,
				InterfaceTypeReplacement: false,

				WasCustomType: false,
				IsCustomType:  false,
			})
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

	//Combine TypeAdjusterData
	if p.TypeAdjusterData != nil && p1.TypeAdjusterData == nil {
		newP.TypeAdjusterData = p.TypeAdjusterData
	} else if p.TypeAdjusterData == nil && p1.TypeAdjusterData != nil {
		newP.TypeAdjusterData = p1.TypeAdjusterData
	} else if p.TypeAdjusterData != nil && p1.TypeAdjusterData != nil {
		combine, err := p.TypeAdjusterData.Combine(p1.TypeAdjusterData)
		if err != nil {
			return nil, err
		}
		newP.TypeAdjusterData = combine
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
