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
return values set. The function only needs to generate the function body.

The build-in time adjuster is a good starting point for your own Typeadjuster, and could be found here: