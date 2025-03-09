package typeAdjustment

import (
	"encoding/json"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/codeGenerators"
	"github.com/Lemonn/JSON2Go/pkg/errors/typeChecker"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
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

func (ta *TypeAdjuster) SetActiveTypeFile(fileData fieldData.FileData) {
	ta.fileData = fileData
}

func (ta *TypeAdjuster) getUnmarshallScaffold(path fieldData.Path, replacementExpr ast.Expr) *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: &ast.Ident{
			Name: "UnMarshall" + path.GetFieldName(),
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
						Type: replacementExpr,
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

func (ta *TypeAdjuster) getMarshallScaffold(path fieldData.Path, replacementExpr ast.Expr) *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: &ast.Ident{
			Name: "Marshall" + path.GetFieldName(),
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
						Type: replacementExpr,
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

func (ta *TypeAdjuster) CheckActiveChecker(path fieldData.Path, checkOnly bool) error {
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
	err = checker.SetState(typeAdjusterData.TypeAdjusterData, path, ta.fileData, ta.codeGenerator)
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
			//TODO implement whit the new generation method
			/*
				err := ta.setFunctions(path, checker)
				if err != nil {
					return err
				}

			*/
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
		ta.fileData[path].ActiveType = nil
		ta.fileData[path].TypeAdjusterData.NameOfActiveTypeAdjuster = nil
		ta.fileData[path].TypeAdjusterData.CheckerVersion = nil
		ta.fileData[path].TypeAdjusterData.SetTimestamp = nil
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

func (ta *TypeAdjuster) AdjustType(path fieldData.Path) (map[fieldData.Path][]*fieldData.File, ast.Expr, *fieldData.Import, error) {
	files := make(map[fieldData.Path][]*fieldData.File)
	var typeImport *fieldData.Import
	var replacementExpr ast.Expr

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
		err := checker.SetState(typeAdjusterData.TypeAdjusterData, path, ta.fileData, ta.codeGenerator)
		// Incompatible TypeAdjusterData should never happen here, as we only call this from the generator, which only
		// works whit already combined files.
		if err != nil {
			return nil, nil, nil, err
		}

		state, err := checker.CouldTypeBeApplied()
		if err != nil {
			return nil, nil, nil, err
		}
		if state == StateApplicable && !runCheckersOnly && !ta.checkOnly {
			checkerName := checker.GetName()
			checkerState, err := checker.GetState()
			if err != nil {
				return nil, nil, nil, err
			}
			replacementExpr, typeImport = checker.GetType()

			activeTypeString, err := utils.ExprToString(replacementExpr)
			if err != nil {
				return nil, nil, nil, err
			}

			ta.fileData[path].ActiveType = &activeTypeString
			typeAdjusterData.NameOfActiveTypeAdjuster = &checkerName
			typeAdjusterData.TypeAdjusterData = []json.RawMessage{checkerState}
			typeAdjusterData.SetTimestamp = ta.startTimestampPointer()
			typeAdjusterData.LastCheckedTimestamp = ta.startTimestampPointer()
			typeAdjusterData.CheckerVersion = checker.GetVersion()

			if checker.NeedsMarshaller() {
				//TODO we need to get, and handle the mod file contents
				marshall, i, err := checker.GenerateMarshall(ta.getMarshallScaffold(path, replacementExpr))
				if err != nil {
					return nil, nil, nil, err
				}
				if _, ok := files[""]; !ok {
					files[""] = []*fieldData.File{}
				}
				files[""] = append(files[""], fieldData.GetGoFile([]ast.Decl{marshall}, i, nil, fieldData.FileClassMarshallFunction))

				unmarshall, i, err := checker.GenerateUnmarshall(ta.getUnmarshallScaffold(path, replacementExpr))
				if err != nil {
					return nil, nil, nil, err
				}
				files[""] = append(files[""], fieldData.GetGoFile([]ast.Decl{unmarshall}, i, nil, fieldData.FileClassUnMarshallFunction))
			}
		} else if state == StateFailed {
			if ta.fileData[path].TypeAdjusterData == nil {
				(*ta.fileData[path]).TypeAdjusterData = &fieldData.TypeAdjusterData{}
			}
			if ta.fileData[path].TypeAdjusterData.CheckedNonMatchingTypes == nil {
				(*(*ta.fileData[path]).TypeAdjusterData).CheckedNonMatchingTypes = map[string]*fieldData.NonMatchingCheckerInfos{}
			}
			(*(*ta.fileData[path]).TypeAdjusterData).CheckedNonMatchingTypes[checker.GetName()] = &fieldData.NonMatchingCheckerInfos{
				Timestamp: ta.startTime.Unix(),
				Version:   checker.GetVersion(),
			}
		}

		if state == StateApplicable || state == StateUndecided {
			runCheckersOnly = true
		}
	}
	return files, replacementExpr, typeImport, nil
}

func (ta *TypeAdjuster) checkerExcluded(path fieldData.Path, checker TypeDeterminationFunction) bool {
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
