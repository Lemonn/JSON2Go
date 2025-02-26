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
	seenTypes     map[string]*fieldData.PathData
	startTime     time.Time
	eaa           bool
	seenTypeUtils *utils.SeenTypeUtils
}

func GenerateTypeFile(jsonData []byte, structName string, externalizeAnonymousArray bool) (map[string]*fieldData.PathData, error) {
	seenTypes := make(map[string]*fieldData.PathData)
	p := Parser{
		seenTypes:     seenTypes,
		startTime:     time.Now(),
		eaa:           externalizeAnonymousArray,
		seenTypeUtils: utils.NewSeenTypeUtils(seenTypes),
	}

	var JsonData interface{}
	var err error
	err = json.Unmarshal(jsonData, &JsonData)
	if err != nil {
		return nil, err
	}
	err = p.codeGen(JsonData, structName, 0)
	if err != nil {
		return nil, err
	}
	if p.eaa {
		p.externalizeAnonymousArray(structName)
	}
	return p.seenTypes, err
}

func (p *Parser) externalizeAnonymousArray(structName string) {
	if p.seenTypeUtils.IsStruct(structName) {
		var levelOfArrays int
		for levelOfArrays, _ = range p.seenTypes[structName].Types["field"] {
			break
		}
		if levelOfArrays > 0 {
			p.seenTypes[structName+".AnonymousArray"] = p.seenTypes[structName]

			p.seenTypes[structName] = &fieldData.PathData{
				Types:                   map[fieldData.Type]map[int]map[string]*fieldData.ValueDetails{},
				JsonFieldName:           "",
				Package:                 nil,
				Omitempty:               false,
				TypeAdjusterData:        nil,
				Error:                   nil,
				RequiredField:           false,
				ForceSourceType:         nil,
				DirectToForceSourceType: false,
			}
			p.seenTypes[structName].Types[fieldData.Field] = map[int]map[string]*fieldData.ValueDetails{}
			p.seenTypes[structName].Types[fieldData.Field][levelOfArrays] = make(map[string]*fieldData.ValueDetails)
			p.seenTypes[structName].Types[fieldData.Field][levelOfArrays][structName+".AnonymousArray"] = &fieldData.ValueDetails{
				Count:              0,
				FirstSeenTimestamp: 0,
				LastSeenTimestamp:  0,
			}
		}

	}
}

func (p *Parser) codeGen(jsonData interface{}, path string, depth int) error {
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
func (p *Parser) processStruct(structData map[string]interface{}, path string, depth int) error {
	for fieldName, field := range structData {
		p.setTypeAtLevel(path, fieldData.Field, depth, path+"."+utils.JsonNameToGoName(fieldName))
		if _, ok := p.seenTypes[path+"."+utils.JsonNameToGoName(fieldName)]; !ok {
			p.seenTypes[path+"."+utils.JsonNameToGoName(fieldName)] = &fieldData.PathData{JsonFieldName: fieldName}
		} else {
			p.seenTypes[path+"."+utils.JsonNameToGoName(fieldName)].JsonFieldName = fieldName
		}
		err := p.codeGen(field, path+"."+utils.JsonNameToGoName(fieldName), 0)
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
func (p *Parser) processSlice(sliceData []interface{}, path string, depth int) error {
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
		}
	}
	if len(sliceData) == 0 {
		p.setTypeAtLevel(path, fieldData.EmptyArray, depth, "[]")
	}
	return nil
}

func (p *Parser) processField(field interface{}, path string, depth int) error {
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

func (p *Parser) setTypeAtLevel(path string, Type fieldData.Type, Depth int, value string) {
	valueDetails := &fieldData.ValueDetails{
		Count:              1,
		FirstSeenTimestamp: p.startTime.Unix(),
		LastSeenTimestamp:  p.startTime.Unix(),
	}
	if p.seenTypes == nil {
		p.seenTypes = map[string]*fieldData.PathData{}
	}
	if v, ok := p.seenTypes[path]; !ok {
		p.seenTypes[path] = &fieldData.PathData{
			Types: map[fieldData.Type]map[int]map[string]*fieldData.ValueDetails{},
		}
	} else if v.Types == nil {
		p.seenTypes[path].Types = map[fieldData.Type]map[int]map[string]*fieldData.ValueDetails{}
	}
	if _, ok := p.seenTypes[path].Types[Type]; !ok {
		p.seenTypes[path].Types[Type] = map[int]map[string]*fieldData.ValueDetails{}
	}
	if _, ok := p.seenTypes[path].Types[Type][Depth]; !ok {
		p.seenTypes[path].Types[Type][Depth] = map[string]*fieldData.ValueDetails{}
	}
	if _, ok := p.seenTypes[path].Types[Type][Depth][value]; !ok {
		p.seenTypes[path].Types[Type][Depth][value] = valueDetails
	} else {
		p.seenTypes[path].Types[Type][Depth][value] = p.seenTypes[path].Types[Type][Depth][value].Combine(valueDetails)
	}
}
