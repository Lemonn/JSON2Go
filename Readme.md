# JSON2Go

Json2Go can generate Golang structs from a single or set of JSON-File/s. It comes with features that make it especially useful when working with undocumented JSON responses.

## Feature List:
 - While other generators assume that the first element of an array represents the type, this one uses all observed elements.
 - Decide which type is appropriate via built-in type-determiners (time.Time and UUID) or custom ones.
 - Determines the final field type by looking at all values seen for a field. For example, if a response for a field contains ["1", "2", "2z"], the final type would be string, not int.
 - Generates a custom unmarshal function for each type. Which uses the functions provided by the type-determiners" to unmarshal special types.
 - The custom unmarshal function returns an error, which does not interrupt the unmarshalling process, upon encountering an unknown field.
 - Use multiple JSON-Files to Determine the type. This could be used to adjust the resulting structure upon newly received data.




For Example this JSON 
```json
{
 "Array": [
  {
   "Test1": 1
  },
  {
   "Test2": 1,
   "Substructure": {"Hello":  "World"}
  },
  {
   "Substructure": {},
   "Subarray": [
    {
     "T": 1
    }
   ]
  },
  {
   "Substructure": {"Hi":  "Moon"},
   "Subarray": [
    {
     "T2": 1
    }
   ]
  }
 ]
}
```
is converted into the following Golang code

```golang
Array []struct {
	Test1		float64	`json:"Test1"`
	Test2		float64	`json:"Test2"`
	Substructure	struct {
		Hello	string	`json:"Hello"`
		Hi	string	`json:"Hi"`
	}	`json:"Substructure"`
	Subarray	[]struct {
		T2	float64	`json:"T2"`
		T	float64	`json:"T"`
	}	`json:"Subarray"`
}
```

## Tag Documentation

Uses the _**json2go**_ tag, to store relevant data. This includes all previously seen values for a field. Its original
type, should the type have been adjusted, and references to the to and from type functions.
The data is organized in a structure.

```golang
type Tag struct {
	SeenValues              []string        `json:"seenValues"`
	CheckedNonMatchingTypes []string        `json:"checkedNonMatchingTypes"`
	ParseFunctions          *ParseFunctions `json:"parseFunctions,omitempty"`
	BaseType                *string         `json:"baseType,omitempty"`
}
```
To store the struct as a tag, it's JSON marshalled and base64 encoded.

The data is used by Json2Go in different steps. It's also used to combine multiple generated structures into one.
Therefore, the data needs to be kept, as long as the type determination is not final. 
After the generation has been finalized, the data could be removed. Sometimes it's required to combine
a lot of data, to determine the struct type. Or even do it in production on live data. 
To prevent the tag from exploding. It also stores a set of strings, that contain all non-matching types.
The Typeadjuster could be set, to ignore the given adjusters.

## Type Adjustment

**Important, it's required to unnest your generated structs first, before using the type adjustment!**

Json2Go can determine the best type of set of input values using Typeadjusters. It comes with some built-in ones 
(time.Time and UUID) and could also be extended with custom ones.

To use the type adjustment, first, create a slice of Typeadjusters. The order determines the priority, 
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

To use a custom adjuster, implement the following interface

```golang
type TypeDeterminationFunction interface {
	CouldTypeBeApplied(seenValues []string) bool
	GetType() string
	GenerateFromTypeFunction(functionScaffold *ast.FuncDecl) *ast.FuncDecl
	GenerateToTypeFunction(functionScaffold *ast.FuncDecl) *ast.FuncDecl
}
```
**_CouldTypeBeApplied_** Receives an array of strings, and returns true if the type could be applied based on the given
set of values.

**_GetType_** Returns the type as a string, for example, time.Time for the time type.

_**GenerateFromTypeFunction**_ and _**GenerateToTypeFunction**_ receive a function with its name and header and 
return values set. The function only needs to generate the function body. To generate the ast-code it's 
advised to write the code first into a dummy function and use this amazing tool https://astextract.lu4p.xyz/ 
to convert it to ast-code.

The build-in time adjuster is a good starting point for your own Typeadjuster, and could be found here: 
[typeDeterminers.go](https://github.com/Lemonn/JSON2Go/blob/ce85a6cc8abf255c8c8733ddbcb10d3dc40fa7a1/typeDeterminers.go#L15)

