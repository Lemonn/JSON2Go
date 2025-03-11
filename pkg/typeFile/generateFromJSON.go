package typeFile

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"strconv"
	"time"
)

type Parser struct {
	fileData  fieldData.FileData
	startTime time.Time
	eaa       bool
}

func GenerateTypeFile(jsonData []byte, structName string, externalizeAnonymousArray bool) (fieldData.FileData, error) {
	p := Parser{
		fileData:  make(fieldData.FileData),
		startTime: time.Now(),
		eaa:       externalizeAnonymousArray,
	}

	var JsonData interface{}
	var err error
	err = json.Unmarshal(jsonData, &JsonData)
	if err != nil {
		return nil, err
	}
	//TODO create proper path whit new
	err = p.codeGen(JsonData, fieldData.Path(structName), 0)
	if err != nil {
		return nil, err
	}
	/*
		if p.eaa {
			p.externalizeAnonymousArray(structName)
		}

	*/
	return p.fileData, err
}

//TODO this should be moved into the codeGenerator, as it'S only up to it how it handles such a case
/*
func (p *Parser) externalizeAnonymousArray(structName string) {
	if p.codeGenerator.IsStruct(structName) {
		var levelOfArrays int
		for levelOfArrays, _ = range p.fileData[structName].Types["field"] {
			break
		}
		if levelOfArrays > 0 {
			p.fileData[structName+".AnonymousArray"] = p.fileData[structName]

			p.fileData[structName] = &fieldData.PathData{
				Types:                   map[fieldData.Type]map[int]map[string]*fieldData.ValueDetails{},
				JsonFieldName:           "",
				TypeAdjusterData:        nil,
				Error:                   nil,
				RequiredField:           false,
				ForceSourceType:         nil,
				DirectToForceSourceType: false,
			}
			p.fileData[structName].Types[fieldData.Field] = map[int]map[string]*fieldData.ValueDetails{}
			p.fileData[structName].Types[fieldData.Field][levelOfArrays] = make(map[string]*fieldData.ValueDetails)
			p.fileData[structName].Types[fieldData.Field][levelOfArrays][structName+".AnonymousArray"] = &fieldData.ValueDetails{
				Count:              0,
				FirstSeenTimestamp: 0,
				LastSeenTimestamp:  0,
			}
		}
	}
}

*/

func (p *Parser) codeGen(jsonData interface{}, path fieldData.Path, depth int) error {
	switch result := jsonData.(type) {
	case map[string]interface{}:
		err := p.processStruct(result, path, depth)
		if err != nil {
			return err
		}
	case []interface{}:
		err := p.processSlice(result, path, depth)
		if err != nil {
			return err
		}
	default:
		err := p.processField(result, path, depth)
		if err != nil {
			return err
		}
	}
	return nil
}

// Processes JSON-Struct elements
func (p *Parser) processStruct(structData map[string]interface{}, path fieldData.Path, depth int) error {
	if _, ok := p.fileData[path]; !ok {
		p.fileData[path] = &fieldData.PathData{}
	}
	p.fileData[path].SeenCounter++
	for fieldName, field := range structData {
		localPath := fieldData.Path(path.String() + "." + utils.JsonNameToGoName(fieldName))
		p.setTypeAtLevel(path, fieldData.Field, depth, localPath.String())
		p.fileData[path].IntroductionCount = p.fileData[path].SeenCounter - 1
		p.fileData[path].LastSeenTimestamp = p.startTime.Unix()
		if p.fileData[path].FirstSeenTimestamp == 0 {
			p.fileData[path].FirstSeenTimestamp = p.startTime.Unix()
		}
		if _, ok := p.fileData[localPath]; !ok {
			p.fileData[fieldData.Path(string(path)+"."+utils.JsonNameToGoName(fieldName))] = &fieldData.PathData{JsonFieldName: fieldName}
		} else {
			p.fileData[localPath].JsonFieldName = fieldName
		}
		err := p.codeGen(field, localPath, 0)
		if err != nil {
			return err
		}
	}
	if len(structData) == 0 {
		p.setTypeAtLevel(path, fieldData.EmptyStruct, depth, "{}")
	}
	return nil
}

// Processes JSON-Array elements
func (p *Parser) processSlice(sliceData []interface{}, path fieldData.Path, depth int) error {
	var err error
	depth++
	for _, i := range sliceData {
		switch v := i.(type) {
		case []interface{}:
			err = p.processSlice(v, path, depth)
			if err != nil {
				return err
			}
		case map[string]interface{}:
			err = p.processStruct(v, path, depth)
			if err != nil {
				return err
			}
		case interface{}:
			err = p.processField(v, path, depth)
			if err != nil {
				return err
			}
		case nil:
			err = p.processField(v, path, depth)
			if err != nil {
				return err
			}
		}
	}
	if len(sliceData) == 0 {
		p.setTypeAtLevel(path, fieldData.EmptyArray, depth, "[]")
	}
	return nil
}

// Processes JSON-Field elements
func (p *Parser) processField(field interface{}, path fieldData.Path, depth int) error {
	if _, ok := p.fileData[path]; !ok {
		p.fileData[path] = &fieldData.PathData{}
	}
	p.fileData[path].SeenCounter++
	var fieldValue string
	switch t := field.(type) {
	case float64:
		fieldValue = strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		if t {
			fieldValue = "true"
		}
		fieldValue = "false"
	case string:
		fieldValue = field.(string)
	case nil:
		fieldValue = "null"
	default:
		return errors.New(fmt.Sprintf("unsupported type of field data: %T", field))
	}
	Type, err := fieldData.TypeFromAny(field)
	if err != nil {
		return err
	}
	p.setTypeAtLevel(path, Type, depth, fieldValue)
	return nil
}

func (p *Parser) setTypeAtLevel(path fieldData.Path, Type fieldData.Type, Depth int, value string) {
	valueDetails := &fieldData.ValueDetails{
		Count:              1,
		FirstSeenTimestamp: p.startTime.Unix(),
		LastSeenTimestamp:  p.startTime.Unix(),
	}
	if p.fileData == nil {
		p.fileData = make(fieldData.FileData)
	}
	if v, ok := p.fileData[path]; !ok {
		p.fileData[path] = &fieldData.PathData{
			Types: map[fieldData.Type]map[int]map[string]*fieldData.ValueDetails{},
		}
	} else if v.Types == nil {
		p.fileData[path].Types = map[fieldData.Type]map[int]map[string]*fieldData.ValueDetails{}
	}
	if _, ok := p.fileData[path].Types[Type]; !ok {
		p.fileData[path].Types[Type] = map[int]map[string]*fieldData.ValueDetails{}
	}
	if _, ok := p.fileData[path].Types[Type][Depth]; !ok {
		p.fileData[path].Types[Type][Depth] = map[string]*fieldData.ValueDetails{}
	}
	if _, ok := p.fileData[path].Types[Type][Depth][value]; !ok {
		p.fileData[path].Types[Type][Depth][value] = valueDetails
	} else {
		p.fileData[path].Types[Type][Depth][value] = p.fileData[path].Types[Type][Depth][value].Combine(valueDetails)
	}
}
