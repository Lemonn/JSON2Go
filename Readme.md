# JSON2Go

 Generates Golang code from a given JSON-Document. The code is returned as an Ast-Tree. Whilst most other
 converters make the assumption, that the first seen element of a type, represents the type. 
 This converter does it differently. It combines types with the same name at the same
 level into one type that contains all unique fields.
 This is useful, if a JSON-Array which omits empty elements should be converted
 into a Golang struct. In case a simple field occurs more than once inside to equally named and leveled structures,
 the last one seen is used


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