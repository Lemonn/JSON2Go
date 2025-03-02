package typeAdjustment

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	j2gError "github.com/Lemonn/JSON2Go/pkg/typeAdjustment/errors"
	"go/ast"
	"go/printer"
	"go/token"
	"strings"
	"time"
)

type TypeAdjuster struct {
	seenTypes              map[string]*fieldData.PathData
	registeredTypeCheckers []TypeDeterminationFunction
	skipPreviouslyFailed   bool
	checkOnly              bool
	startTime              time.Time
	*utils.SeenTypeUtils
}

func NewTypeAdjuster(seenTypes map[string]*fieldData.PathData, checkers []TypeDeterminationFunction, startTime time.Time) *TypeAdjuster {
	return &TypeAdjuster{
		seenTypes:              seenTypes,
		registeredTypeCheckers: checkers,
		skipPreviouslyFailed:   false,
		checkOnly:              false,
		startTime:              startTime,
		SeenTypeUtils:          utils.NewSeenTypeUtils(seenTypes),
	}
}

func (ta *TypeAdjuster) searchTypeDeterminationFunctionByName(name string) (TypeDeterminationFunction, error) {
	for i, checker := range ta.registeredTypeCheckers {
		if checker.GetName() == name {
			return ta.registeredTypeCheckers[i], nil
		}
	}
	return nil, errors.New("type adjustment function not found")
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
						Type: ta.GetFieldType(path, true),
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
						Type: ta.GetFieldType(path, true),
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

	ta.seenTypes[path].TypeAdjusterData.ParseFunctions = &fieldData.ParseFunctions{
		Unmarshall:        unmarshallFunctionOutput.String(),
		UnmarshallImports: unmarshallImports,
		Marshall:          marshallFunctionOutput.String(),
		MarshallImports:   marshallImports,
	}
	return nil
}

func (ta *TypeAdjuster) getTypeString(checker TypeDeterminationFunction) (string, error) {
	marshallFunctionOutput := bytes.NewBuffer([]byte{})
	if err := printer.Fprint(marshallFunctionOutput, token.NewFileSet(), checker.GetType()); err != nil {
		return "", err
	}
	return marshallFunctionOutput.String(), nil
}

func (ta *TypeAdjuster) AdjustTypesNew(path string) error {
	var runCheckersOnly bool
	//TODO set the required imports
	//Handle case were there is an active type replacement
	if ta.seenTypes[path].TypeAdjusterData != nil && ta.seenTypes[path].TypeAdjusterData.NameOfActiveTypeAdjuster != nil {
		//TODO set checker state
		checker, err := ta.searchTypeDeterminationFunctionByName(*ta.seenTypes[path].TypeAdjusterData.NameOfActiveTypeAdjuster)
		if err != nil {
			return j2gError.ActiveAdjusterNotFoundError{}
		}
		state, err := checker.CouldTypeBeApplied(path)
		if err != nil {
			//TODO check for complex type change error
			//TODO we could potentially avoid this for the time type, if we start to support both types.
			return err
		}
		if state == StateApplicable {
			if checker.TypeExpansion() {
				err := ta.setFunctions(path, checker)
				if err != nil {
					return err
				}
				ta.seenTypes[path].Error = errors.Join(ta.seenTypes[path].Error, &j2gError.TypeExpansionError{
					Path:      path,
					Timestamp: ta.startTime.Unix(),
				})
			}
			checkerState, err := checker.GetState()
			if err != nil {
				return err
			}
			ta.seenTypes[path].TypeAdjusterData.TypeAdjusterData = []json.RawMessage{checkerState}
			ta.seenTypes[path].TypeAdjusterData.LastCheckedTimestamp = ta.startTime.Unix()
			runCheckersOnly = true
		} else {
			// TODO if u think that the error could be set here, yes, but the new type for the error is not yet determined
			//TODO set error for type change -> This needs to steps, run all other checkers and see if one is
			// applicable, if not the new type is the base type
		}
	}

	for _, checker := range ta.registeredTypeCheckers {
		if ta.checkerExcluded(path, checker) {
			continue
		}

		var TypeAdjusterData []json.RawMessage
		if ta.seenTypes[path].TypeAdjusterData != nil {
			TypeAdjusterData = ta.seenTypes[path].TypeAdjusterData.TypeAdjusterData
		}

		err := checker.SetState(TypeAdjusterData, path, ta.registeredTypeCheckers, ta.seenTypes)
		if err != nil {
			//TODO handle incompatible state error
			return err
		}
		state, err := checker.CouldTypeBeApplied(path)
		if err != nil {
			return err
		}
		if state == StateApplicable && !runCheckersOnly && !ta.checkOnly {
			checkerName := checker.GetName()
			checkerState, err := checker.GetState()
			if err != nil {
				return err
			}
			typeString, err := ta.getTypeString(checker)
			if err != nil {
				return err
			}

			if ta.seenTypes[path].TypeAdjusterData == nil {
				ta.seenTypes[path].TypeAdjusterData = &fieldData.TypeAdjusterData{}
			}
			ta.seenTypes[path].TypeAdjusterData.NameOfActiveTypeAdjuster = &checkerName
			ta.seenTypes[path].TypeAdjusterData.TypeAdjusterData = []json.RawMessage{checkerState}
			ta.seenTypes[path].ActiveType = &typeString
			ta.seenTypes[path].TypeAdjusterData.SetTimestamp = ta.startTime.Unix()
			ta.seenTypes[path].TypeAdjusterData.LastCheckedTimestamp = ta.startTime.Unix()
			ta.seenTypes[path].ForceSourceType = checker.ForceSourceType()
			ta.seenTypes[path].TypeAdjusterData.CheckerVersion = checker.GetVersion()
			ta.seenTypes[path].TypeAdjusterData.ModFileContents = checker.GetModFileContents()

			err = ta.setFunctions(path, checker)
			if err != nil {
				return err
			}

		} else if state == StateFailed {
			if ta.seenTypes[path].TypeAdjusterData == nil {
				(*ta.seenTypes[path]).TypeAdjusterData = &fieldData.TypeAdjusterData{}
			}
			if ta.seenTypes[path].TypeAdjusterData.CheckedNonMatchingTypes == nil {
				(*(*ta.seenTypes[path]).TypeAdjusterData).CheckedNonMatchingTypes = map[string]int64{}
			}
			(*(*ta.seenTypes[path]).TypeAdjusterData).CheckedNonMatchingTypes[checker.GetName()] = ta.startTime.Unix()
		}

		if state == StateApplicable || state == StateUndecided {
			runCheckersOnly = true
		}
	}

	return nil
}

func (ta *TypeAdjuster) checkerExcluded(path string, checker TypeDeterminationFunction) bool {
	if ta.seenTypes[path].TypeAdjusterData != nil {
		//Check if excluded by user
		if ta.seenTypes[path].TypeAdjusterData.ExcludedTypeCheckers != nil {
			for _, typeChecker := range ta.seenTypes[path].TypeAdjusterData.ExcludedTypeCheckers {
				if checker.GetName() == typeChecker {
					return true
				}
			}
		}

		//Check if excluded because of previous failure
		if ta.seenTypes[path].TypeAdjusterData.CheckedNonMatchingTypes != nil {
			if _, ok := ta.seenTypes[path].TypeAdjusterData.CheckedNonMatchingTypes[checker.GetName()]; ok {
				return true
			}
		}
	}
	return false
}

// TODO
func (ta *TypeAdjuster) errorHandler(path string, err error) error {

	return nil
}

// TODO
func (ta *TypeAdjuster) appendErrorToPath(path string, err error) error {

	return nil
}
