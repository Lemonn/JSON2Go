package JSON2Go

import (
	"reflect"
	"testing"
)

func TestTag_Combine(t *testing.T) {
	type fields struct {
		SeenValues              map[string]string
		CheckedNonMatchingTypes []string
		ParseFunctions          *ParseFunctions
		BaseType                *string
	}
	type args struct {
		j1 *Tag
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Tag
		wantErr bool
	}{
		{name: "CombinationOfValidTags", fields: struct {
			SeenValues              map[string]string
			CheckedNonMatchingTypes []string
			ParseFunctions          *ParseFunctions
			BaseType                *string
		}{SeenValues: map[string]string{"01a8573c-11d9-3efb-9aa6-dcbc0880272a": "string"}, CheckedNonMatchingTypes: nil, ParseFunctions: &ParseFunctions{
			FromTypeParseFunction: "fromIDTest",
			ToTypeParseFunction:   "toIDTest",
		}, BaseType: func() *string {
			s := "string"
			return &s
		}()}, args: struct{ j1 *Tag }{j1: &Tag{
			SeenValues:              map[string]string{"1b1e4578-8d7c-4abe-a37e-697f29a484dd": "string"},
			CheckedNonMatchingTypes: nil,
			ParseFunctions: &ParseFunctions{
				FromTypeParseFunction: "fromIDTest",
				ToTypeParseFunction:   "toIDTest",
			},
			BaseType: func() *string {
				s := "string"
				return &s
			}(),
		}}, want: &Tag{
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
			j := &Tag{
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
