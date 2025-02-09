# JSON2Go

<!-- TOC -->
* [JSON2Go](#json2go)
  * [Overview](#overview)
  * [Tag Documentation](#tag-documentation)
  * [Type Adjustment](#type-adjustment)
    * [Build-in TypeCheckers](#build-in-typecheckers)
  * [JSON-Marshaller generation](#json-marshaller-generation)
    * [UnMarshal](#unmarshal)
      * [Example](#example)
    * [Marshal](#marshal)
  * [Combine files](#combine-files)
  * [Example](#example-1)
<!-- TOC -->

## Overview

Json2Go could generate Golang structs from any given valid JSON document. But this is something that could be done by many tools out there. What sets Json2Go apart is the fact that it has been developed with undocumented or poorly documented JSON-APIs in mind.

The fact that such APIs change without notice and often omit empty fields. Makes it hard to generate a good matching Golang data type.
Here comes Json2Go into play. It provides features tailored to overcome the hurdles of such APIs.

- The core feature to overcome this obstacle is the ability to not only look at one but at all elements of an array. To determine the resulting datatype. But also the ability to combine a set of JSON responses into one data structure.

- This is paired with the ability to determine the resulting type of each field, based on all seen values. Either via built-in functions or custom ones.

- To counter the fact that sometimes a field is added or rarely seen. A soft error, which does not interrupt the unmarshalling process, is returned upon entering an unknown field.


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
[Build-in TypeCheckers](#build-in-typecheckers) and could also be extended with custom ones.

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
`CouldTypeBeApplied(seenValues []string) bool` Receives an array of strings, and returns true if the type could be applied based on the given
set of values.

`GetType()` Returns the type as a string, for example, time.Time for the time type.

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

### UnMarshal

`JSON2Go.GenerateJsonUnmarshall(file)`

The Unmarshal generator honors the custom types generated by the Typeadjusters. On top of that, it comes with logic,
which soft errors whenever it encounters an unknown field.
In this case a AdditionalElementErrors is returned. This could be used, to update the structure on the fly. 
The error contains the name of the parsed object and the unknown elements as a `map[element's name]json.RawMessage`

```golang
type AdditionalElementsError struct {
	ParsedObj string
	Elements  map[string]json.RawMessage
}
```
The soft error is realized by joining all AdditionalElementErrors together. Only if an error not of the  
AdditionalElementError type is observed, is the processing terminated, and the combined error is returned.

To check for an error on the caller side, use `CheckForFirstErrorNotOfTypeT`. To collect all
AdditionalElementsErrors use `GetAllErrorsOfType`.

#### Example

Check for non AdditionalElementError
```golang
var rawJson []byte
var generatedStruct GeneratedStruct
err := json.Unmarshal(rawJson, &generatedStruct)
e := GeneratePackage.CheckForFirstErrorNotOfTypeT(GeneratePackage.AdditionalElementError{}, err)
if e != nil {
	//TODO handle error
}
```

Get slice of AdditionalElementErrors
```golang
elementErrors, err := GeneratePackage.GetAllErrorsOfType(GeneratePackage.AdditionalElementError{}, err)
if err != nil {
	return
} else {
	for _, elementError := range elementErrors {
		//TODO handle  AdditionalElementError
	}
}
```

### Marshal

`JSON2Go.GenerateJsonMarshall(file)`

A marshal function is generated for each struct, that contains a custom type. 

## Combine files

When working with undocumented JSON-APIs it's essential to generate the datatype not from a single response. 
But instead use a wide set of responses, to be able to get the most complete data structure. To archive this, 
Json2Go comes with the ability to combine all equally named structs of two files into one file.

Let's say we're received the following to JSON-Responses from an API:

```json
{
  "Name": "Nick",
  "Age": "99",
  "Food": "BBQ"
}
```

```json
{
  "Name": "Peter",
  "Age": "22",
  "Gender": "male",
  "Origin": "US"
}
```



```golang
type RRR struct {
	Food	string	`json:"Food" json2go:"eyJzZWVuVmFsdWVzIjpbIkJCUSJdfQ=="`
	Gender	string	`json:"Gender" json2go:"eyJzZWVuVmFsdWVzIjpbIm1hbGUiXX0="`
	Origin	string	`json:"Origin" json2go:"eyJzZWVuVmFsdWVzIjpbIlVTIl19"`
	Name	string	`json:"Name" json2go:"eyJzZWVuVmFsdWVzIjpbIk5pY2siLCJQZXRlciJdfQ==" `
	Age	string	`json:"Age" json2go:"eyJzZWVuVmFsdWVzIjpbIjk5IiwiMjIiXX0=" `
}
```

## Example

The follwing JSON-Document is 

```json
{
  "ListOfUUIDs": [
    "c46f9b5c-13b9-4db8-9272-fc549b26e90b",
    "9e4f11c2-d059-4172-a6b8-1fb3dcb77916"
  ],
  "DateString": "2025-02-20T17:25:00",
  "ArrayWhitStructs": [
    {
      "Name": "Anne",
      "Gender": "f",
      "FavoriteFood": "ButterChicken"
    },
    {
      "Name": "Carl",
      "Profession": "carpenter"
    },
    {
      "Name": "Carla",
      "LastName": "Miller"
    }
  ]
}
```


```golang

```

```golang
package main

import (
	"github.com/google/uuid"
	"time"
)

type Test struct {
	ListOfUuids		uuid.UUID		`json2go:"eyJzZWVuVmFsdWVzIjpbImM0NmY5YjVjLTEzYjktNGRiOC05MjcyLWZjNTQ5YjI2ZTkwYiIsIjllNGYxMWMyLWQwNTktNDE3Mi1hNmI4LTFmYjNkY2I3NzkxNiJdLCJwYXJzZUZ1bmN0aW9ucyI6eyJmcm9tVHlwZVBhcnNlRnVuY3Rpb24iOiJmcm9tTGlzdE9mVXVpZHNSUlIiLCJ0b1R5cGVQYXJzZUZ1bmN0aW9uIjoidG9MaXN0T2ZVdWlkc1JSUiJ9LCJiYXNlVHlwZSI6InN0cmluZyIsImxhc3RTZWVuVGltZXN0YW1wIjoxNzM4NTc3Mzg4fQ==" json:"ListOfUUIDs" `
	DateString		time.Time		`json2go:"eyJzZWVuVmFsdWVzIjpbIjIwMjUtMDItMjBUMTc6MjU6MDAiXSwicGFyc2VGdW5jdGlvbnMiOnsiZnJvbVR5cGVQYXJzZUZ1bmN0aW9uIjoiZnJvbURhdGVTdHJpbmdSUlIiLCJ0b1R5cGVQYXJzZUZ1bmN0aW9uIjoidG9EYXRlU3RyaW5nUlJSIn0sImJhc2VUeXBlIjoic3RyaW5nIiwibGFzdFNlZW5UaW1lc3RhbXAiOjE3Mzg1NzczODh9" json:"DateString" `
	ArrayWhitStructs	[]*ArrayWhitStructs	`json:"ArrayWhitStructs,omitempty"`
}
type ArrayWhitStructs struct {
	Name		string	`json2go:"eyJzZWVuVmFsdWVzIjpbIkFubmUiLCJDYXJsIiwiQ2FybGEiXSwibGFzdFNlZW5UaW1lc3RhbXAiOjE3Mzg1NzczODh9" json:"Name" `
	Gender		string	`json2go:"eyJzZWVuVmFsdWVzIjpbImYiXSwibGFzdFNlZW5UaW1lc3RhbXAiOjE3Mzg1NzczODh9"  json:"Gender"`
	FavoriteFood	string	`json2go:"eyJzZWVuVmFsdWVzIjpbIkJ1dHRlckNoaWNrZW4iXSwibGFzdFNlZW5UaW1lc3RhbXAiOjE3Mzg1NzczODh9"  json:"FavoriteFood"`
	Profession	string	`json2go:"eyJzZWVuVmFsdWVzIjpbImNhcnBlbnRlciJdLCJsYXN0U2VlblRpbWVzdGFtcCI6MTczODU3NzM4OH0="  json:"Profession"`
	LastName	string	`json2go:"eyJzZWVuVmFsdWVzIjpbIk1pbGxlciJdLCJsYXN0U2VlblRpbWVzdGFtcCI6MTczODU3NzM4OH0="  json:"LastName"`
}
```
