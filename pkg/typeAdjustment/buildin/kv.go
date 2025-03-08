package buildin

import (
	"encoding/json"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/codeGenerators"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment"
	"github.com/Lemonn/JSON2Go/pkg/typeFile"
	"go/ast"
	"time"
)

type KV struct {
	generator     func(input string) (map[string]string, error)
	codeGenerator codeGenerators.CodeGenerator
	activePath    string
	fileData      fieldData.FileData
	state         *KvState
	files         map[string]*fieldData.File
}

type KvState struct {
	fileData          fieldData.FileData
	creationTimestamp int64
}

func NewKV(generator func(input string) (map[string]string, error)) *KV {
	return &KV{
		generator: generator,
	}
}

func (k *KV) CouldTypeBeApplied() (typeAdjustment.State, error) {
	var kvs []map[string]string

	if _, ok := k.fileData[k.activePath].Types[fieldData.String]; !ok {
		return typeAdjustment.StateFailed, nil
	}
	isBasicType, levelOfArrays, _ := k.codeGenerator.IsBasicTypeWhitDetails(k.activePath)
	if !isBasicType {
		return typeAdjustment.StateFailed, nil
	}
	for value, _ := range k.fileData[k.activePath].Types[fieldData.String][levelOfArrays] {
		kv, err := k.generator(value)
		if err != nil {
			return typeAdjustment.StateFailed, nil
		} else {
			kvs = append(kvs, kv)
		}
	}

	var files []*typeFile.FileDetail
	//Add state file
	files = append(files, &typeFile.FileDetail{
		FileData:     k.state.fileData,
		CreationDate: time.Unix(k.state.creationTimestamp, 0),
	})
	for _, kv := range kvs {
		marshal, err := json.Marshal(kv)
		if err != nil {
			return typeAdjustment.StateFailed, err
		}
		file, err := typeFile.GenerateTypeFile(marshal, utils.GetFieldName(k.activePath), false)
		if err != nil {
			return typeAdjustment.StateFailed, err
		}
		files = append(files, &typeFile.FileDetail{
			CreationDate: time.Now(),
			FileData:     file,
		})
	}

	c := typeFile.NewCombiner(k.codeGenerator, time.Now())
	combined, err := c.CombineFileDetails(files)
	if err != nil {
		return typeAdjustment.StateFailed, err
	}

	//TODO analyze the generated file for file errors if some are found, return them

	k.codeGenerator.SetActiveTypeFile(combined)

	subFiles, err := k.codeGenerator.Generate()
	if err != nil {
		return typeAdjustment.StateFailed, err
	}

	k.state.fileData = combined
	k.files = subFiles

	return typeAdjustment.StateApplicable, nil
}

func (k *KV) GetType() ast.Expr {
	return nil
}

func (k *KV) GenerateMarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, []string, error) {
	return nil, nil, nil
}

func (k *KV) GenerateUnmarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, []string, error) {
	return nil, nil, nil
}

func (k *KV) GetName() string {
	return "json2go.KV"
}

func (k *KV) SetState(states []json.RawMessage, currentPath string, fileData fieldData.FileData, activeTypeCheckers typeAdjustment.TypeDeterminationFunctions, codeGenerator codeGenerators.CodeGenerator) error {
	//TODO implement me
	panic("implement me")
}

func (k *KV) GetState() (json.RawMessage, error) {
	//TODO implement me
	panic("implement me")
}

func (k *KV) GetExtraCode() ([]ast.Decl, []string, error) {
	//TODO implement me
	panic("implement me")
}

func (k *KV) TypeExpansion() bool {
	//TODO implement me
	panic("implement me")
}

func (k *KV) ForceSourceType() *string {
	//TODO implement me
	panic("implement me")
}

func (k *KV) GetModFileContents() []*fieldData.ModFileContent {
	//TODO implement me
	panic("implement me")
}

func (k *KV) GetVersion() *string {
	//TODO implement me
	panic("implement me")
}
