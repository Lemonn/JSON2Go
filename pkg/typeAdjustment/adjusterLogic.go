package typeAdjustment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/codeGenerators"
	"github.com/Lemonn/JSON2Go/pkg/errors/typeChecker"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"go/printer"
	"go/token"
	"strings"
	"time"
)

type TypeAdjuster struct {
	fileData               fieldData.FileData
	registeredTypeCheckers TypeDeterminationFunctions
	skipPreviouslyFailed   bool
	checkOnly              bool
	startTime              time.Time
	codeGenerator          codeGenerators.CodeGenerator
}

func NewTypeAdjuster(fileData fieldData.FileData, codeGenerator codeGenerators.CodeGenerator, checkers []TypeDeterminationFunction, startTime time.Time) *TypeAdjuster {
	return &TypeAdjuster{
		fileData:               fileData,
		registeredTypeCheckers: checkers,
		skipPreviouslyFailed:   false,
		checkOnly:              false,
		startTime:              startTime,
		codeGenerator:          codeGenerator,
	}
}

func (ta *TypeAdjuster) GetFileHeader() []string {
	if ta.registeredTypeCheckers == nil {
		return []string{}
	}
	return ta.registeredTypeCheckers.GenerateHeader()
}

func (ta *TypeAdjuster) SetActiveTypeFile(fileData map[string]*fieldData.PathData) {
	ta.fileData = fileData
}

func (ta *TypeAdjuster) getUnmarshallScaffold(path string, checker TypeDeterminationFunction) *ast.FuncDecl {
	pathElements := strings.Split(path, ".")
	return &ast.FuncDecl{
		Name: &ast.Ident{
			Name: "UnMarshall" + pathElements[len(pathElements)-1],
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							{
								Name: "baseValue",
							},
						},
						Type: ta.codeGenerator.GetFieldType(path, true),
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: checker.GetType(),
					},
					{
						Type: &ast.Ident{
							Name: "error",
						},
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{},
		},
	}
}

func (ta *TypeAdjuster) getMarshallScaffold(path string, checker TypeDeterminationFunction) *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: &ast.Ident{
			Name: "Marshall" + utils.GetFieldName(path),
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							{
								Name: "baseValue",
							},
						},
						Type: checker.GetType(),
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: ta.codeGenerator.GetFieldType(path, true),
					},
					{
						Type: &ast.Ident{
							Name: "error",
						},
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{},
		},
	}
}

func (ta *TypeAdjuster) setFunctions(path string, checker TypeDeterminationFunction) error {
	unmarshallFunction, unmarshallImports, err := checker.GenerateUnmarshall(ta.getUnmarshallScaffold(path, checker))
	if err != nil {
		return err
	}

	unmarshallFunctionOutput := bytes.NewBuffer([]byte{})
	if err := printer.Fprint(unmarshallFunctionOutput, token.NewFileSet(), unmarshallFunction); err != nil {
		return err
	}

	marshallFunction, marshallImports, err := checker.GenerateMarshall(ta.getMarshallScaffold(path, checker))
	if err != nil {
		return err
	}

	marshallFunctionOutput := bytes.NewBuffer([]byte{})
	if err := printer.Fprint(marshallFunctionOutput, token.NewFileSet(), marshallFunction); err != nil {
		return err
	}

	extraCode, extraImports, err := checker.GetExtraCode()
	if err != nil {
		return err
	}
	extraCodeOutput := bytes.NewBuffer([]byte{})
	if err := printer.Fprint(extraCodeOutput, token.NewFileSet(), extraCode); err != nil {
		return err
	}

	ta.fileData[path].TypeAdjusterData.ParseFunctions = &fieldData.ParseFunctions{
		Unmarshall:        unmarshallFunctionOutput.String(),
		UnmarshallImports: unmarshallImports,
		Marshall:          marshallFunctionOutput.String(),
		MarshallImports:   marshallImports,
		ExtraCode:         extraCodeOutput.String(),
		ExtraImports:      extraImports,
	}
	return nil
}

func (ta *TypeAdjuster) CheckActiveChecker(path string, checkOnly bool) error {
	typeAdjusterData := ta.fileData[path].TypeAdjusterData

	if !(ta.fileData[path].TypeAdjusterData != nil && ta.fileData[path].TypeAdjusterData.NameOfActiveTypeAdjuster != nil) {
		return nil
	}

	checker, err := ta.registeredTypeCheckers.GetByName(*ta.fileData[path].TypeAdjusterData.NameOfActiveTypeAdjuster)
	if err != nil {
		return &typeChecker.ActiveAdjusterNotFoundError{
			NotFoundCheckerName: *typeAdjusterData.NameOfActiveTypeAdjuster,
			ActiveCheckerNames:  ta.registeredTypeCheckers.GetNames(),
		}
	}
	err = checker.SetState(typeAdjusterData.TypeAdjusterData, path, ta.fileData, ta.registeredTypeCheckers, ta.codeGenerator)
	if err != nil {
		return err
	}

	state, err := checker.CouldTypeBeApplied()
	if err != nil {
		//TODO check for complex type change error
		//TODO we could potentially avoid this for the time type, if we start to support both types.
		return err
	}
	if state == StateApplicable {
		if checkOnly {
			return nil
		}
		if checker.TypeExpansion() {
			err := ta.setFunctions(path, checker)
			if err != nil {
				return err
			}
			//TODO return this error instead of setting it directly
			/*
				ta.fileData[path].Error = errors.Join(ta.fileData[path].Error, &j2gError.TypeExpansionError{
					Path: path,
				})

			*/
		}
		checkerState, err := checker.GetState()
		if err != nil {
			return err
		}
		typeAdjusterData.TypeAdjusterData = []json.RawMessage{checkerState}
		typeAdjusterData.LastCheckedTimestamp = ta.startTimestampPointer()
	} else {
		//TODO put the reset into an extra method
		ta.fileData[path].TypeAdjusterData.ParseFunctions = nil
		ta.fileData[path].ActiveType = nil
		ta.fileData[path].TypeAdjusterData.NameOfActiveTypeAdjuster = nil
		ta.fileData[path].TypeAdjusterData.CheckerVersion = nil
		ta.fileData[path].TypeAdjusterData.SetTimestamp = nil
		ta.fileData[path].TypeAdjusterData.ModFileContents = nil
		return &typeChecker.NoLongerApplicableCustomTypeError{
			ActiveTypeChecker: checker.GetName(),
		}
	}
	return nil
}

func (ta *TypeAdjuster) startTimestampPointer() *int64 {
	i := ta.startTime.Unix()
	return &i
}

func (ta *TypeAdjuster) AdjustTypes(path string) error {
	var runCheckersOnly bool
	if ta.fileData[path].TypeAdjusterData == nil {
		ta.fileData[path].TypeAdjusterData = &fieldData.TypeAdjusterData{}
	}
	typeAdjusterData := ta.fileData[path].TypeAdjusterData

	//TODO set the required imports
	for _, checker := range ta.registeredTypeCheckers {
		if ta.checkerExcluded(path, checker) {
			continue
		}
		fmt.Println("#" + path)
		err := checker.SetState(typeAdjusterData.TypeAdjusterData, path, ta.fileData, ta.registeredTypeCheckers, ta.codeGenerator)
		// Incompatible TypeAdjusterData should never happen here, as we only call this from the generator, which only
		// works whit already combined files.
		if err != nil {
			return err
		}

		state, err := checker.CouldTypeBeApplied()
		if err != nil {
			return err
		}
		if state == StateApplicable && !runCheckersOnly && !ta.checkOnly {
			checkerName := checker.GetName()
			checkerState, err := checker.GetState()
			if err != nil {
				return err
			}
			activeType, err := utils.ExprToString(checker.GetType())
			if err != nil {
				return err
			}

			ta.fileData[path].ActiveType = &activeType
			ta.fileData[path].ForceSourceType = checker.ForceSourceType()

			typeAdjusterData.NameOfActiveTypeAdjuster = &checkerName
			typeAdjusterData.TypeAdjusterData = []json.RawMessage{checkerState}
			typeAdjusterData.SetTimestamp = ta.startTimestampPointer()
			typeAdjusterData.LastCheckedTimestamp = ta.startTimestampPointer()
			typeAdjusterData.CheckerVersion = checker.GetVersion()
			typeAdjusterData.ModFileContents = checker.GetModFileContents()

			err = ta.setFunctions(path, checker)
			if err != nil {
				return err
			}

		} else if state == StateFailed {
			if ta.fileData[path].TypeAdjusterData == nil {
				(*ta.fileData[path]).TypeAdjusterData = &fieldData.TypeAdjusterData{}
			}
			if ta.fileData[path].TypeAdjusterData.CheckedNonMatchingTypes == nil {
				(*(*ta.fileData[path]).TypeAdjusterData).CheckedNonMatchingTypes = map[string]int64{}
			}
			(*(*ta.fileData[path]).TypeAdjusterData).CheckedNonMatchingTypes[checker.GetName()] = ta.startTime.Unix()
		}

		if state == StateApplicable || state == StateUndecided {
			runCheckersOnly = true
		}
	}

	return nil
}

func (ta *TypeAdjuster) checkerExcluded(path string, checker TypeDeterminationFunction) bool {
	if ta.fileData[path].TypeAdjusterData != nil {
		//Check if excluded by user
		if ta.fileData[path].TypeAdjusterData.ExcludedTypeCheckers != nil {
			for _, excludedChecker := range ta.fileData[path].TypeAdjusterData.ExcludedTypeCheckers {
				if checker.GetName() == excludedChecker {
					return true
				}
			}
		}

		//Check if excluded because of previous failure
		if ta.fileData[path].TypeAdjusterData.CheckedNonMatchingTypes != nil {
			if _, ok := ta.fileData[path].TypeAdjusterData.CheckedNonMatchingTypes[checker.GetName()]; ok {
				return true
			}
		}
	}
	return false
}
