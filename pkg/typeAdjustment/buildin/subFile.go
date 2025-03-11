package buildin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/codeGenerators"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment"
	"github.com/Lemonn/JSON2Go/pkg/typeFile"
	"go/ast"
	"go/parser"
	"go/token"
	"time"
)

type SubFile struct {
	generator     func(input string) (interface{}, error)
	codeGenerator codeGenerators.CodeGenerator
	activePath    fieldData.Path
	fileData      fieldData.FileData
	state         *KvState
	files         map[fieldData.Path][]*fieldData.File
	generatorCode *ast.FuncLit
}

type KvState struct {
	fileData          fieldData.FileData
	creationTimestamp int64
}

type SubFileSettings struct {
	Generator     func(input string) (interface{}, error)
	GeneratorCode string
}

func (s *SubFile) ValidateFunction(generatorCode string) (*ast.FuncLit, error) {
	gc, err := parser.ParseExpr(generatorCode)
	if err != nil {
		return nil, err
	}
	if fl, ok := gc.(*ast.FuncLit); ok {
		if fl.Type.Params.List == nil {
			return nil, errors.New("empty function input, the input must be func(input string)")
		} else if len(fl.Type.Params.List) < 1 {
			return nil, errors.New("invalid function input, the input must be func(input string)")
		} else {
			if fl.Type.Params.List[0].Names[0].Name != "input" {
				return nil, errors.New(fmt.Sprintf("invalid function input name: %s, the input must be func(input string)", fl.Type.Params.List[0].Names[0].Name))
			} else if ident, ok := fl.Type.Params.List[0].Type.(*ast.Ident); !ok {
				return nil, errors.New(fmt.Sprintf("function input must be of type *ast.Ident, provided was %T", fl.Type.Params.List[0].Type))
			} else if ident.Name != "string" {
				return nil, errors.New(fmt.Sprintf("invalid function input type: %s, the input must be func(input string)", ident.Name))
			}
		}

		if fl.Type.Results.List == nil {
			return nil, errors.New("function returns nothing, should return (interface{}, error)")
		} else if len(fl.Type.Results.List) < 2 {
			return nil, errors.New("function returns to few values, should return (interface{}, error)")
		} else if interfaceType, ok := fl.Type.Results.List[0].Type.(*ast.InterfaceType); !ok {
			return nil, errors.New(fmt.Sprintf("retrun value should be of type *ast.InterfaceType, is %T", fl.Type.Results.List[0].Type))
		} else if interfaceType.Methods != nil && interfaceType.Methods.List != nil && len(interfaceType.Methods.List) > 0 {
			return nil, errors.New("first return value should be empty interface, but has methods")
		} else if ident, ok := fl.Type.Results.List[1].Type.(*ast.Ident); !ok {
			return nil, errors.New(fmt.Sprintf("second retrun value should be of type *ast.Ident, is %T", fl.Type.Results.List[1].Type))
		} else if ident.Name != "error" {
			return nil, errors.New(fmt.Sprintf("second retrun value should be of type string, is %s", ident.Name))
		}
		return fl, nil
	} else {
		return nil, errors.New(fmt.Sprintf("only *ast.FuncLit is supported, provided was %T", gc))
	}

}

func NewSubFile(generator func(input string) (interface{}, error), generatorCode string) (*SubFile, error) {
	s := &SubFile{
		generator: generator,
	}

	gc, err := s.ValidateFunction(generatorCode)
	if err != nil {
		return nil, err
	}
	s.generatorCode = gc
	return s, nil
}

func (s *SubFile) Clone() typeAdjustment.TypeDeterminationFunction {
	return &SubFile{
		generator: s.generator,
	}
}

func (s *SubFile) CouldTypeBeApplied() (typeAdjustment.State, error) {
	var rawSubFiles []interface{}

	if _, ok := s.fileData[s.activePath].Types[fieldData.String]; !ok {
		return typeAdjustment.StateFailed, nil
	}
	isBasicType, levelOfArrays, _ := s.codeGenerator.IsBasicTypeWhitDetails(s.activePath)
	if !isBasicType {
		return typeAdjustment.StateFailed, nil
	}
	for value, _ := range s.fileData[s.activePath].Types[fieldData.String][levelOfArrays] {
		sf, err := s.generator(value)
		if err != nil {
			return typeAdjustment.StateFailed, nil
		} else {
			rawSubFiles = append(rawSubFiles, sf)
		}
	}

	var files []*typeFile.FileDetail
	//Add state file
	if s.state != nil {
		files = append(files, &typeFile.FileDetail{
			FileData:     s.state.fileData,
			CreationDate: time.Unix(s.state.creationTimestamp, 0),
		})
	}

	for _, sf := range rawSubFiles {
		marshal, err := json.Marshal(sf)
		if err != nil {
			return typeAdjustment.StateFailed, err
		}
		file, err := typeFile.GenerateTypeFile(marshal, s.activePath.GetFieldName(), false)
		if err != nil {
			return typeAdjustment.StateFailed, err
		}
		files = append(files, &typeFile.FileDetail{
			CreationDate: time.Now(),
			FileData:     file,
		})
	}

	c := typeFile.NewCombiner(s.codeGenerator, time.Now())
	combined, err := c.CombineFileDetails(files)
	if err != nil {
		return typeAdjustment.StateFailed, err
	}

	//TODO analyze the generated file for file errors if some are found, return them

	subCodeGen := s.codeGenerator.Clone(combined)
	subCodeGen.SetActiveTypeFile(combined)
	subFiles, err := subCodeGen.Generate()
	if err != nil {
		return typeAdjustment.StateFailed, err
	}
	s.state.fileData = combined
	//TODO replace whit start timestamp
	s.state.creationTimestamp = time.Now().Unix()
	s.files = subFiles

	return typeAdjustment.StateApplicable, nil
}

func (s *SubFile) GetType() (ast.Expr, fieldData.Imports) {
	return &ast.SelectorExpr{X: ast.NewIdent(s.activePath.GetFieldName()), Sel: ast.NewIdent(s.activePath.GetFieldName())},
		fieldData.Imports{
			&fieldData.Import{
				Path:            s.activePath,
				NeedsAdjustment: true,
			}}
}

func (s *SubFile) GenerateMarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, fieldData.Imports, error) {
	return nil, nil, nil
}

func (s *SubFile) GenerateUnmarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, fieldData.Imports, error) {
	functionScaffold.Body.List = []ast.Stmt{
		&ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.Ident{
					Name: "result",
				},
				&ast.Ident{
					Name: "err",
				},
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: s.generatorCode,
					Args: []ast.Expr{
						&ast.Ident{
							Name: "baseValue",
						},
					},
				},
			},
		},
		&ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X: &ast.Ident{
					Name: "err",
				},
				Op: token.NEQ,
				Y: &ast.Ident{
					Name: "nil",
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.Ident{
								Name: "nil",
							},
							&ast.Ident{
								Name: "err",
							},
						},
					},
				},
			},
		},
		&ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.Ident{
					Name: "subJson",
				},
				&ast.Ident{
					Name: "err",
				},
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.Ident{
							Name: "json",
						},
						Sel: &ast.Ident{
							Name: "Marshal",
						},
					},
					Args: []ast.Expr{
						&ast.Ident{
							Name: "result",
						},
					},
				},
			},
		},
		&ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X: &ast.Ident{
					Name: "err",
				},
				Op: token.NEQ,
				Y: &ast.Ident{
					Name: "nil",
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.Ident{
								Name: "nil",
							},
							&ast.Ident{
								Name: "err",
							},
						},
					},
				},
			},
		},
		&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{
							&ast.Ident{
								Name: "j",
							},
						},
						Type: &ast.SelectorExpr{
							X: &ast.Ident{
								Name: "JourneyId",
							},
							Sel: &ast.Ident{
								Name: "JourneyId",
							},
						},
					},
				},
			},
		},
		&ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.Ident{
					Name: "err",
				},
			},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.Ident{
							Name: "json",
						},
						Sel: &ast.Ident{
							Name: "Unmarshal",
						},
					},
					Args: []ast.Expr{
						&ast.Ident{
							Name: "subJson",
						},
						&ast.UnaryExpr{
							Op: token.AND,
							X: &ast.Ident{
								Name: "j",
							},
						},
					},
				},
			},
		},
		&ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X: &ast.Ident{
					Name: "err",
				},
				Op: token.NEQ,
				Y: &ast.Ident{
					Name: "nil",
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.Ident{
								Name: "nil",
							},
							&ast.Ident{
								Name: "err",
							},
						},
					},
				},
			},
		},
		&ast.ReturnStmt{
			Results: []ast.Expr{
				&ast.UnaryExpr{
					Op: token.AND,
					X: &ast.Ident{
						Name: "j",
					},
				},
				&ast.Ident{
					Name: "nil",
				},
			},
		},
	}
	//TODO set imports
	imports := fieldData.Imports{
		&fieldData.Import{
			Path: "encoding/json",
		},
		&fieldData.Import{
			Path: "errors",
		},
		&fieldData.Import{
			Path: "fmt",
		},
		&fieldData.Import{
			Path: "strings",
		},
		&fieldData.Import{
			Path:            s.activePath,
			NeedsAdjustment: true,
		},
	}
	return functionScaffold, imports, nil
}

func (s *SubFile) SetState(states []json.RawMessage, currentPath fieldData.Path, fileData fieldData.FileData, codeGenerator codeGenerators.CodeGenerator) error {
	s.state = &KvState{}

	s.activePath = currentPath
	s.fileData = fileData
	s.codeGenerator = codeGenerator
	return nil
}

func (s *SubFile) NeedsMarshaller() bool {
	return true
}

func (s *SubFile) GetName() string {
	return "json2go.SubFile"
}

func (s *SubFile) GetState() (json.RawMessage, error) {
	return json.Marshal(s.state)
}

func (s *SubFile) GetSubFiles() (map[fieldData.Path][]*fieldData.File, error) {
	return s.files, nil
}

func (s *SubFile) TypeExpansion() bool {
	//TODO implement me
	return false
}

func (s *SubFile) GetVersion() *string {
	return utils.StringToPointer("v0.0.1")
}
