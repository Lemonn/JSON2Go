package jsonMarshallerGenerators

import "github.com/Lemonn/JSON2Go/pkg/fieldData"

type Generator interface {
	Unmarshall(path fieldData.Path) ([]*fieldData.File, error)
	Marshall(path fieldData.Path) ([]*fieldData.File, error)
	GlobalFiles() []*fieldData.File
}
