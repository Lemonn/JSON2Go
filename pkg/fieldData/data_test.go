package fieldData

import (
	"reflect"
	"testing"
)

func TestTag_Combine(t *testing.T) {
	type fields struct {
		SeenValues              map[string]string
		CheckedNonMatchingTypes map[string]int64
		ParseFunctions          *ParseFunctions
		BaseType                *string
	}
	type args struct {
		j1 *FieldData
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *FieldData
		wantErr bool
	}{
		{name: "CombinationOfValidTags", fields: struct {
			SeenValues              map[string]string
			CheckedNonMatchingTypes map[string]int64
			ParseFunctions          *ParseFunctions
			BaseType                *string
		}{SeenValues: map[string]string{"01a8573c-11d9-3efb-9aa6-dcbc0880272a": "string"}, CheckedNonMatchingTypes: nil, ParseFunctions: &ParseFunctions{
			FromTypeParseFunction: "fromIDTest",
			ToTypeParseFunction:   "toIDTest",
		}, BaseType: func() *string {
			s := "string"
			return &s
		}()}, args: struct{ j1 *FieldData }{j1: &FieldData{
			SeenValues:              map[string]string{"1b1e4578-8d7c-4abe-a37e-697f29a484dd": "string"},
			CheckedNonMatchingTypes: map[string]int64{},
			ParseFunctions: &ParseFunctions{
				FromTypeParseFunction: "fromIDTest",
				ToTypeParseFunction:   "toIDTest",
			},
			BaseType: func() *string {
				s := "string"
				return &s
			}(),
		}}, want: &FieldData{
			SeenValues:              map[string]string{"01a8573c-11d9-3efb-9aa6-dcbc0880272a": "string", "1b1e4578-8d7c-4abe-a37e-697f29a484dd": "string"},
			CheckedNonMatchingTypes: nil,
			ParseFunctions: &ParseFunctions{
				FromTypeParseFunction: "fromIDTest",
				ToTypeParseFunction:   "toIDTest",
			},
			BaseType: func() *string {
				s := "string"
				return &s
			}(),
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &FieldData{
				SeenValues:              tt.fields.SeenValues,
				CheckedNonMatchingTypes: tt.fields.CheckedNonMatchingTypes,
				ParseFunctions:          tt.fields.ParseFunctions,
				BaseType:                tt.fields.BaseType,
			}
			got, err := j.Combine(tt.args.j1)
			if (err != nil) != tt.wantErr {
				t.Errorf("Combine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Combine() got = %v, want %v", got, tt.want)
			}
		})
	}
}
