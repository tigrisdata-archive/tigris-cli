// Copyright 2022-2023 Tigris Data, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//nolint:golint,funlen
package schema

import (
	"encoding/json"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tigrisdata/tigris-client-go/schema"
)

func TestSchemaInference(t *testing.T) {
	cases := []struct {
		name string
		exp  *schema.Schema
		in   [][]byte
	}{
		{
			name: "types", in: [][]byte{
				[]byte(`{
	"str_field" : "str_value",
	"int_field" : 1,
	"float_field" : 1.1,
	"uuid_field" : "1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1",
	"time_field" : "2022-11-04T16:17:23.967964263-07:00",
	"bool_field" : true,
	"binary_field": "cGVlay1hLWJvbwo=",
	"object" : {
		"str_field" : "str_value",
		"int_field" : 1,
		"float_field" : 1.1,
		"uuid_field" : "1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1",
		"time_field" : "2022-11-04T16:17:23.967964263-07:00",
		"bool_field" : true,
		"binary_field": "cGVlay1hLWJvbwo="
	},
	"array" : [ {
		"str_field" : "str_value",
		"int_field" : 1,
		"float_field" : 1.1,
		"uuid_field" : "1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1",
		"time_field" : "2022-11-04T16:17:23.967964263-07:00",
		"bool_field" : true,
		"binary_field": "cGVlay1hLWJvbwo="
	} ],
    "prim_array" : [ "str" ],
    "array_uuid" : [ "1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1" ]
}`),
			}, exp: &schema.Schema{
				Name: "types",
				Fields: map[string]*schema.Field{
					"uuid_field":   {Type: "string", Format: "uuid"},
					"time_field":   {Type: "string", Format: "date-time"},
					"str_field":    {Type: "string"},
					"bool_field":   {Type: "boolean"},
					"float_field":  {Type: "number"},
					"int_field":    {Type: "integer"},
					"binary_field": {Type: "string", Format: "byte"},
					"object": {Type: "object", Fields: map[string]*schema.Field{
						"uuid_field":   {Type: "string", Format: "uuid"},
						"time_field":   {Type: "string", Format: "date-time"},
						"str_field":    {Type: "string"},
						"bool_field":   {Type: "boolean"},
						"float_field":  {Type: "number"},
						"int_field":    {Type: "integer"},
						"binary_field": {Type: "string", Format: "byte"},
					}},
					"array": {
						Type: "array",
						Items: &schema.Field{
							Type: "object",
							Fields: map[string]*schema.Field{
								"uuid_field":   {Type: "string", Format: "uuid"},
								"time_field":   {Type: "string", Format: "date-time"},
								"str_field":    {Type: "string"},
								"bool_field":   {Type: "boolean"},
								"float_field":  {Type: "number"},
								"int_field":    {Type: "integer"},
								"binary_field": {Type: "string", Format: "byte"},
							},
						},
					},
					"prim_array": {
						Type: "array",
						Items: &schema.Field{
							Type: "string",
						},
					},
					"array_uuid": {
						Type: "array",
						Items: &schema.Field{
							Type:   typeString,
							Format: formatUUID,
						},
					},
				},
			},
		},

		{
			name: "empty_obj_arr", in: [][]byte{
				[]byte(`{
	"int_field" : 1,
	"empty_object" : {},
    "empty_array" : [],
	"object_with_empty_arr" : { "empty_array_emb": [] },
	"array_with_empty_obj" : [ {} ],
	"str_field" : "str_value"
}`),
			}, exp: &schema.Schema{
				Name: "empty_obj_arr",
				Fields: map[string]*schema.Field{
					"int_field": {Type: "integer"},
					"str_field": {Type: "string"},
				},
			},
		},

		{
			name: "broaden_type", in: [][]byte{
				[]byte(`{
	"int_field" : 1,
	"uuid_field" : "1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1",
	"time_field" : "2022-11-04T16:17:23.967964263-07:00",
	"binary_field": "cGVlay1hLWJvbwo=",
	"object" : {
		"int_field" : 1,
		"uuid_field" : "1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1",
		"time_field" : "2022-11-04T16:17:23.967964263-07:00",
		"binary_field": "cGVlay1hLWJvbwo="
	},
	"array" : [ {
		"int_field" : 1,
		"uuid_field" : "1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1",
		"time_field" : "2022-11-04T16:17:23.967964263-07:00",
		"binary_field": "cGVlay1hLWJvbwo="
	} ]
}`),
				[]byte(`{
	"int_field" : 1.1,
	"uuid_field" : "str9cd3a2ea3d58e9d1",
	"time_field" : "20221104T16172396796426307:00",
	"binary_field": "not_base64",
	"object" : {
		"int_field" : 1.2,
		"uuid_field" : "str9cd3a2ea3d58e9d1111",
		"time_field" : "20221104T1617239679642630700111",
		"binary_field": "not_base6411111"
	},
	"array" : [ {
		"int_field" : 1.3,
		"uuid_field" : "nnnnn1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1",
		"time_field" : "ttttt2022-11-04T16:17:23.967964263-07:00",
		"binary_field": "notbase64"
	} ]
}`),
			}, exp: &schema.Schema{
				Name: "broaden_type",
				Fields: map[string]*schema.Field{
					"uuid_field":   {Type: "string"},
					"time_field":   {Type: "string"},
					"int_field":    {Type: "number"},
					"binary_field": {Type: "string"},
					"object": {Type: "object", Fields: map[string]*schema.Field{
						"uuid_field":   {Type: "string"},
						"time_field":   {Type: "string"},
						"int_field":    {Type: "number"},
						"binary_field": {Type: "string"},
					}},
					"array": {
						Type: "array",
						Items: &schema.Field{
							Type: "object",
							Fields: map[string]*schema.Field{
								"uuid_field":   {Type: "string"},
								"time_field":   {Type: "string"},
								"int_field":    {Type: "number"},
								"binary_field": {Type: "string"},
							},
						},
					},
				},
			},
		},
		{
			name: "broaden_array_type", in: [][]byte{
				[]byte(`{
	"array" : [ {
		"int_field" : 1,
		"uuid_field" : "1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1",
		"time_field" : "2022-11-04T16:17:23.967964263-07:00",
		"binary_field": "cGVlay1hLWJvbwo="
	},
	 {
		"int_field" : 1.3,
		"uuid_field" : "nnnnn1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1",
		"time_field" : "ttttt2022-11-04T16:17:23.967964263-07:00",
		"binary_field": "notbase64"
	} ]
}`),
			}, exp: &schema.Schema{
				Name: "broaden_array_type",
				Fields: map[string]*schema.Field{
					"array": {
						Type: "array",
						Items: &schema.Field{
							Type: "object",
							Fields: map[string]*schema.Field{
								"uuid_field":   {Type: "string"},
								"time_field":   {Type: "string"},
								"int_field":    {Type: "number"},
								"binary_field": {Type: "string"},
							},
						},
					},
				},
			},
		},
		{
			name: "adding_fields", in: [][]byte{
				[]byte(`{
	"int_field" : 1,
	"object" : {
		"int_field" : 1
	},
	"array" : [ {
		"int_field" : 1
	} ]
}`),
				[]byte(`{
	"uuid_field" : "1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1",
	"object" : {
		"uuid_field" : "1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1"
	},
	"array" : [ {
		"uuid_field" : "1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1"
	} ]
}`),
			}, exp: &schema.Schema{
				Name: "adding_fields",
				Fields: map[string]*schema.Field{
					"uuid_field": {Type: "string", Format: formatUUID},
					"int_field":  {Type: "integer"},
					"object": {Type: "object", Fields: map[string]*schema.Field{
						"uuid_field": {Type: "string", Format: formatUUID},
						"int_field":  {Type: "integer"},
					}},
					"array": {
						Type: "array",
						Items: &schema.Field{
							Type: "object",
							Fields: map[string]*schema.Field{
								"uuid_field": {Type: "string", Format: formatUUID},
								"int_field":  {Type: "integer"},
							},
						},
					},
				},
			},
		},
		{
			name: "adding_array_object_fields", in: [][]byte{
				[]byte(`{ "array" : [ { "int_field" : 1 }, { "int_field_two" : 1 } ] }`),
			}, exp: &schema.Schema{
				Name: "adding_array_object_fields",
				Fields: map[string]*schema.Field{
					"array": {
						Type: "array",
						Items: &schema.Field{
							Type: "object",
							Fields: map[string]*schema.Field{
								"int_field":     {Type: "integer"},
								"int_field_two": {Type: "integer"},
							},
						},
					},
				},
			},
		},
	}

	DetectByteArrays = true

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var sch schema.Schema
			ptr := unsafe.Pointer(&c.in)
			err := Infer(&sch, c.name, *(*[]json.RawMessage)(ptr), nil, nil, 0)
			require.NoError(t, err)
			assert.Equal(t, c.exp, &sch)
		})
	}
}

func TestSchemaInferenceNegative(t *testing.T) {
	cases := []struct {
		name string
		err  error
		in   [][]byte
	}{
		{
			name: "incompatible_primitive",
			in: [][]byte{
				[]byte(`{ "incompatible_field" : 1 }`),
				[]byte(`{ "incompatible_field" : "1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1" }`),
			},
			err: newInompatibleSchemaError("incompatible_field", "integer", "", "string", "uuid"),
		},
		{
			name: "incompatible_prim_to_object",
			in: [][]byte{
				[]byte(`{ "incompatible_field" : 1 }`),
				[]byte(`{ "incompatible_field" : { "field1": "1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1" } }`),
			},
			err: newInompatibleSchemaError("incompatible_field", "integer", "", "object", ""),
		},
		{
			name: "incompatible_object_to_prim",
			in: [][]byte{
				[]byte(`{ "incompatible_field" : { "field1": "1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1" } }`),
				[]byte(`{ "incompatible_field" : 1 }`),
			},
			err: newInompatibleSchemaError("incompatible_field", "object", "", "integer", ""),
		},
		{
			name: "incompatible_array_to_prim",
			in: [][]byte{
				[]byte(`{ "incompatible_field" : ["1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1"] }`),
				[]byte(`{ "incompatible_field" : 1 }`),
			},
			err: newInompatibleSchemaError("incompatible_field", "array", "", "integer", ""),
		},
		{
			name: "incompatible_prim_to_array",
			in: [][]byte{
				[]byte(`{ "incompatible_field" : 1 }`),
				[]byte(`{ "incompatible_field" : ["1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1"] }`),
			},
			err: newInompatibleSchemaError("incompatible_field", "integer", "", "array", ""),
		},
		{
			name: "incompatible_array_mixed",
			in: [][]byte{
				[]byte(`{ "incompatible_field" : ["1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1", 1] }`),
			},
			err: newInompatibleSchemaError("incompatible_field", "string", "uuid", "integer", ""),
		},
		{
			name: "incompatible_array",
			in: [][]byte{
				[]byte(`{ "incompatible_field" : [ 1 ] }`),
				[]byte(`{ "incompatible_field" : ["1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1"] }`),
			},
			err: newInompatibleSchemaError("incompatible_field", "integer", "", "string", "uuid"),
		},
		{
			name: "incompatible_array_object_mixed",
			in: [][]byte{
				[]byte(`{ "incompatible_field" : [ { "one" : "1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1" }, { "one" : 1 } ] }`),
			},
			err: newInompatibleSchemaError("one", "string", "uuid", "integer", ""),
		},
		{
			name: "incompatible_array_object",
			in: [][]byte{
				[]byte(`{ "incompatible_field" : [ { "one" : "1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1" } ] }`),
				[]byte(`{ "incompatible_field" : [ { "one" : 1 } ] }`),
			},
			err: newInompatibleSchemaError("one", "string", "uuid", "integer", ""),
		},
		{
			name: "incompatible_object",
			in: [][]byte{
				[]byte(`{ "incompatible_field" : { "one" : 1 } }`),
				[]byte(`{ "incompatible_field" : { "one" : "1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1" } }`),
			},
			err: newInompatibleSchemaError("one", "integer", "", "string", "uuid"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var sch schema.Schema
			ptr := unsafe.Pointer(&c.in)
			err := Infer(&sch, c.name, *(*[]json.RawMessage)(ptr), nil, nil, 0)
			assert.Equal(t, c.err, err)
		})
	}
}

func TestSchemaTags(t *testing.T) {
	cases := []struct {
		name         string
		primaryKey   []string
		autoGenerate []string
		exp          *schema.Schema
		in           [][]byte
	}{
		{
			name: "pk_autogenerate",
			in: [][]byte{
				[]byte(`{ "uuid_field" : "1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1",
"uuid_field1" : "1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1" }`),
			},
			primaryKey:   []string{"uuid_field"},
			autoGenerate: []string{"uuid_field1"},
			exp: &schema.Schema{
				Name: "pk_autogenerate",
				Fields: map[string]*schema.Field{
					"uuid_field":  {Type: "string", Format: "uuid"},
					"uuid_field1": {Type: "string", Format: "uuid", AutoGenerate: true},
				},
				PrimaryKey: []string{"uuid_field"},
			},
		},
		{
			name: "pk_autogenerate_multi",
			in: [][]byte{
				[]byte(`{ "uuid_field" : "1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1",
"uuid_field1" : "1ed6ff32-4c0f-4553-9cd3-a2ea3d58e9d1" }`),
			},
			primaryKey:   []string{"uuid_field", "uuid_field1"},
			autoGenerate: []string{"uuid_field1", "uuid_field"},
			exp: &schema.Schema{
				Name: "pk_autogenerate_multi",
				Fields: map[string]*schema.Field{
					"uuid_field":  {Type: "string", Format: "uuid", AutoGenerate: true},
					"uuid_field1": {Type: "string", Format: "uuid", AutoGenerate: true},
				},
				PrimaryKey: []string{"uuid_field", "uuid_field1"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var sch schema.Schema
			ptr := unsafe.Pointer(&c.in)
			err := Infer(&sch, c.name, *(*[]json.RawMessage)(ptr), c.primaryKey, c.autoGenerate, 0)
			require.NoError(t, err)
			assert.Equal(t, c.exp, &sch)
		})
	}
}
