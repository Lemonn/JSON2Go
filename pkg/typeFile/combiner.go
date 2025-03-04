package typeFile

import (
	"errors"
	timestampError "github.com/Lemonn/JSON2Go/internal/error"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/codeGenerators"
	j2gErrors "github.com/Lemonn/JSON2Go/pkg/errors"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"maps"
	"sort"
	"time"
)

type Combiner struct {
	startTime     time.Time
	codeGenerator codeGenerators.CodeGenerator
}

type FileDetail struct {
	CreationDate time.Time
	FileData     fieldData.FileData
}

type FileDetails []*FileDetail

func (f FileDetails) Len() int {
	return len(f)
}

func (f FileDetails) Less(i, j int) bool {
	return f[i].CreationDate.Before(f[j].CreationDate)
}

func (f FileDetails) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (c *Combiner) InterfaceReplacement(p, p1 *fieldData.PathData) bool {
	return false
}

func (c *Combiner) processStackedTypeChecks(stackedTypeChecks map[string]struct{}, fileData fieldData.FileData) error {
	c.codeGenerator.SetActiveTypeFile(fileData)
	for path := range stackedTypeChecks {
		err := c.codeGenerator.CheckType(path)
		err = c.appendFileErrorsToTypeFile(err, fileData[path])
		if err != nil {
			return err
		}
	}
	return nil
}

func NewCombiner(codeGenerator codeGenerators.CodeGenerator, startTime time.Time) *Combiner {
	return &Combiner{
		startTime:     startTime,
		codeGenerator: codeGenerator,
	}
}

func (c *Combiner) CombineFileDetails(fileDetails FileDetails) (fieldData.FileData, error) {
	sort.Sort(fileDetails)
	stackedTypeChecks := make(map[string]struct{})
	for {
		if len(fileDetails) == 1 {
			break
		}
		paths := make(map[string][]*fieldData.PathData)
		combined := make(fieldData.FileData)
		for path, data := range fileDetails[0].FileData {
			if _, ok := paths[path]; !ok {
				paths[path] = []*fieldData.PathData{}
			}
			paths[path] = append(paths[path], data)
		}

		for path, data := range fileDetails[1].FileData {
			if _, ok := paths[path]; !ok {
				paths[path] = []*fieldData.PathData{}
			}
			paths[path] = append(paths[path], data)
		}

		for path, data := range paths {
			if len(data) == 1 {
				combined[path] = data[0]
			} else {
				var oldParentSeenCounter int
				if v, ok := fileDetails[0].FileData[utils.GetParentPath(path)]; ok {
					oldParentSeenCounter = v.SeenCounter
				}
				combinedPathData, st, err := c.combinePathData(data[0], data[1], path, oldParentSeenCounter)
				if err != nil {
					return nil, err
				}
				maps.Copy(stackedTypeChecks, st)
				combined[path] = combinedPathData
			}
		}
		creationDate := fileDetails[1].CreationDate
		fileDetails = fileDetails[2:]
		fileDetails = append(fileDetails, &FileDetail{
			CreationDate: creationDate,
			FileData:     combined,
		})
	}
	err := c.processStackedTypeChecks(stackedTypeChecks, fileDetails[0].FileData)
	if err != nil {
		return nil, err
	}
	return fileDetails[0].FileData, nil
}

// TODO only return on hard errors, write all soft errors to the file
func (c *Combiner) combinePathData(p, p1 *fieldData.PathData, path string, oldParentSeenCounter int) (*fieldData.PathData, map[string]struct{}, error) {
	stackedTypeChecks := make(map[string]struct{})
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
			hardError := c.appendFileErrorsToTypeFile(err, &newP)
			if hardError != nil {
				return nil, nil, hardError
			}
			newP.TypeAdjusterData = combine
			if err == nil {
				stackedTypeChecks[path] = struct{}{}
				//c.laterCheck = append(c.laterCheck, path)
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
		return nil, nil, &j2gErrors.ConflictingJsonFieldNameError{
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
			return nil, nil, &j2gErrors.ConflictingForceSourceTypeError{}
		} else {
			newP.ForceSourceType = p.ForceSourceType
		}
	}

	if p.DirectToForceSourceType != p1.DirectToForceSourceType {
		//TODO respect conflict resolving strategic
		return nil, nil, &j2gErrors.ConflictingForceSourceTypeError{}
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
		newP.SeenCounter = p1.SeenCounter + oldParentSeenCounter
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

	return &newP, stackedTypeChecks, nil
}

func (c *Combiner) appendFileErrorsToTypeFile(err error, pathData *fieldData.PathData) error {
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
		pathData.Error = errors.Join(pathData.Error, ue)
	}
	return nil
}
