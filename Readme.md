# JSON2Go

Json2Go is an advanced JSON to Golang-datatype generator. That comes with a lot of features besides the pure data type generation. It has been written with undocumented JSON-APIs in mind. The goal was to give a similar experience to a documented API.

When using documented APIs, one creates a struct and file type parser from a given specification such as OpenAPI. Json2Go provides a similar experience. But instead of taking in a given specification, it uses a set of API responses to determine the struct and corresponding types.


## Features

- Obviously, generate Go structs from a JSON file.

- In the case of arrays, all elements are considered to determine the proper type.

- Combine JSON responses from an API into one struct.

- Determine the field type, either via built-in detectors or custom ones.

- Generate un/-marshall functions for the determined types.

- Get notified when an unknown field is encountered during unmarshalling.

## Type Adjustment

**Important, it's required to unnest your generated structs first, before using the type adjustment!**

Json2Go can determine the best type of set of input values using TypeAdjusters. It comes with some built-in ones
[Build-in TypeCheckers](#build-in-typecheckers) and could also be extended with custom ones.


### Build-in

To use the type adjustment, first, create a slice of TypeAdjusters. The order determines the priority, 
should more than one create a match.

```golang
TypeCheckers := []JSON2Go.TypeDeterminationFunction{&JSON2Go.TimeTypeChecker{}}
```

Then call the adjustment function

```golang
err := JSON2Go.AdjustTypes(file, TypeCheckers)
if err != nil {
	fmt.Println(err)
	return
}
```

### Custom ones

To create a custom adjuster, implement the following interface

```golang
type TypeDeterminationFunction interface {
	CouldTypeBeApplied(seenValues map[string]string) bool
	GetType() ast.Expr
	GenerateFromTypeFunction(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error)
	GenerateToTypeFunction(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error)
	GetRequiredImports() []string
	SetFile(file *ast.File)
	GetName() string
	SetState(state json.RawMessage) error
	GetState() (json.RawMessage, error)
}
```

`GenerateFromTypeFunction(functionScaffold *ast.FuncDecl) *ast.FuncDecl` and 
`GenerateToTypeFunction(functionScaffold *ast.FuncDecl) *ast.FuncDecl` receive a function with its name and header and 
return values set. The function only needs to generate the function body. To generate the ast-code it's 
advised to write the code first into a dummy function and use this amazing tool https://astextract.lu4p.xyz/ 
to convert it to ast-code.

### Build-in TypeCheckers


| Name                    | Type      | Settings                                                        | Description                                                                                                                                                                                                                                                       |
|-------------------------|-----------|-----------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| json2go.UUIDTypeChecker | uuid.UUID |                                                                 | Checks if values could be represented as `uuid.UUID` using `github.com/google/uuid` as a dependency                                                                                                                                                               |
| json2go.TimeTypeChecker | time.Time | `IgnoreYearOnlyStrings` Set to true, to ignore year only values | Checks if values can be represented as `time.Time` uses `github.com/araddon/dateparse` to check for valid time types.                                                                                                                                             |
| json2go.IntTypeChecker  | int       |                                                                 | Checks if the given values can be represented as integers. Be careful with APIs that use the dot notation, to signal a float response, despite the values being valid integer ones. The information is lost during the process. Only the pure value plays a role. |

The build-in time adjuster is a good starting point for your own Typeadjuster, and could be found here: 
[typeDeterminers.go](https://github.com/Lemonn/JSON2Go/blob/ce85a6cc8abf255c8c8733ddbcb10d3dc40fa7a1/typeDeterminers.go#L15)

## JSON-Marshaller generation

If desired, custom marshall functions could be generated. If Typeadjusters are used, this is required!

### Unmarshall handling of unknown fields

Sometimes a field is added, or a rarely seen field appears for the first time. To get notified in such cases,
the generated unmarshal functions return a soft error whenever an unknown field is encountered. This error
is of type `AdditionalElementsError`. It's a soft error because it does not interrupt the unmarshalling process.
Instead, it gets joined and is returned to the caller.

To assist in the handling of such errors. Two helper functions are provided. The first one checks for hard errors.

```golang
var rawJson []byte
var generatedStruct GeneratedStruct
err := json.Unmarshal(rawJson, &generatedStruct)
e := GeneratePackage.CheckForFirstErrorNotOfTypeT(GeneratePackage.AdditionalElementError{}, err)
if e != nil {
	//TODO handle error
}
```

The second one returns a list of all `AdditionalElementsError` errors collected during the unmarshalling process.

```golang
elementErrors, err := GeneratePackage.GetAllErrorsOfType(GeneratePackage.AdditionalElementError{}, err)
if err != nil {
	return
} else {
	for _, elementError := range elementErrors {
		//TODO handle  AdditionalElementErrors
	}
}
```