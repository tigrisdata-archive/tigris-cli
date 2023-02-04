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

package schema

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/tigrisdata/tigris-client-go/schema"
)

const (
	typeInteger = "integer"
	typeString  = "string"
	typeBoolean = "boolean"
	typeNumber  = "number"
	typeArray   = "array"
	typeObject  = "object"
)

// Supported subtypes.
const (
	formatByte     = "byte"
	formatDateTime = "date-time"
	formatUUID     = "uuid"
)

var (
	ErrIncompatibleSchema = fmt.Errorf("error incompatible schema")
	ErrExpectedString     = fmt.Errorf("expected string type")
	ErrExpectedNumber     = fmt.Errorf("expected json.Number")
	ErrUnsupportedType    = fmt.Errorf("unsupported type")
)

func parseDateTime(s string) bool {
	if _, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return true
	}

	// FIXME: Support more formats. Need server side support or convert it here in CLI

	return false
}

func translateStringType(v interface{}) (string, string, error) {
	t := reflect.TypeOf(v)

	if t.PkgPath() == "encoding/json" && t.Name() == "Number" {
		n, ok := v.(json.Number)
		if !ok {
			return "", "", ErrExpectedNumber
		}

		if _, err := n.Int64(); err != nil {
			_, err = n.Float64()
			if err != nil {
				return "", "", err
			}

			return typeNumber, "", nil
		}

		return typeInteger, "", nil
	}

	s, ok := v.(string)
	if !ok {
		return "", "", ErrExpectedString
	}

	if parseDateTime(s) {
		return typeString, formatDateTime, nil
	}

	if _, err := uuid.Parse(s); err == nil {
		return typeString, formatUUID, nil
	}

	b := make([]byte, base64.StdEncoding.DecodedLen(len(s)))
	if _, err := base64.StdEncoding.Decode(b, []byte(s)); err == nil {
		return typeString, formatByte, nil
	}

	return typeString, "", nil
}

func translateType(v interface{}) (string, string, error) {
	t := reflect.TypeOf(v)

	//nolint:golint,exhaustive
	switch t.Kind() {
	case reflect.Bool:
		return typeBoolean, "", nil
	case reflect.Float64:
		return typeNumber, "", nil
	case reflect.String:
		return translateStringType(v)
	case reflect.Slice, reflect.Array:
		return typeArray, "", nil
	case reflect.Map:
		return typeObject, "", nil
	default:
		return "", "", errors.Wrapf(ErrUnsupportedType, "name='%s' kind='%s'", t.Name(), t.Kind())
	}
}

func extendedStringType(oldType string, oldFormat string, newType string, newFormat string) (string, string, error) {
	if newFormat == "" {
		switch {
		case oldFormat == formatByte:
			return newType, newFormat, nil
		case oldFormat == formatUUID:
			return newType, newFormat, nil
		case oldFormat == formatDateTime:
			return newType, newFormat, nil
		}
	} else if newFormat == formatByte && oldFormat == "" {
		return oldType, oldFormat, nil
	}

	return "", "", ErrIncompatibleSchema
}

// this is only matter for initial schema inference, where we have luxury to extend the type
// if not detected properly from the earlier data records
// we extend:
// int -> float
// byte -> string
// time -> string
// uuid => string.
func extendedType(oldType string, oldFormat string, newType string, newFormat string) (string, string, error) {
	if oldType == typeInteger && newType == typeNumber {
		return newType, newFormat, nil
	}

	if oldType == typeString && newType == typeString {
		if t, f, err := extendedStringType(oldType, oldFormat, newType, newFormat); err == nil {
			return t, f, nil
		}
	}

	if newType == oldType && newFormat == oldFormat {
		return newType, newFormat, nil
	}

	log.Debug().Str("oldType", oldType).Str("newType", newType).Msg("incompatible schema")
	log.Debug().Str("oldFormat", oldFormat).Str("newFormat", newFormat).Msg("incompatible schema")

	return "", "", ErrIncompatibleSchema
}

func traverseObject(existingField *schema.Field, newField *schema.Field, values map[string]any) error {
	switch {
	case existingField == nil:
		newField.Fields = make(map[string]*schema.Field)
	case existingField.Type == typeObject:
		if existingField.Fields == nil {
			newField.Fields = make(map[string]*schema.Field)
		} else {
			newField.Fields = existingField.Fields
		}
	default:
		return ErrIncompatibleSchema
	}

	return traverseFields(newField.Fields, values, nil)
}

func traverseArray(existingField *schema.Field, newField *schema.Field, v any) error {
	for i := 0; i < reflect.ValueOf(v).Len(); i++ {
		t, format, err := translateType(reflect.ValueOf(v).Index(i).Interface())
		if err != nil {
			return err
		}

		if i == 0 {
			switch {
			case existingField == nil:
				newField.Items = &schema.Field{Type: t, Format: format}
			case existingField.Type == typeArray:
				newField.Items = existingField.Items
			default:
				return ErrIncompatibleSchema
			}
		}

		nt, nf, err := extendedType(newField.Items.Type, newField.Items.Format, t, format)
		if err != nil {
			return err
		}

		newField.Items.Type = nt
		newField.Items.Format = nf

		if t == typeObject {
			values, _ := reflect.ValueOf(v).Index(i).Interface().(map[string]any)
			if err := traverseObject(newField.Items, newField.Items, values); err != nil {
				return err
			}

			if len(newField.Items.Fields) == 0 {
				newField.Items = nil
			}
		}
	}

	return nil
}

func setAutoGenerate(autoGen []string, name string, field *schema.Field) {
	if autoGen != nil {
		// FIXME: Convert to O(1)
		for i := 0; i < len(autoGen); i++ {
			if autoGen[i] == name {
				field.AutoGenerate = true
			}
		}
	}
}

func traverseFieldsLow(t string, format string, k string, f *schema.Field, v any, sch map[string]*schema.Field,
) (bool, error) {
	switch {
	case t == typeObject:
		vm, _ := v.(map[string]any)
		if err := traverseObject(sch[k], f, vm); err != nil {
			return false, err
		}

		if len(f.Fields) == 0 {
			return true, nil
		}
	case t == typeArray:
		// FIXME: Support multidimensional arrays
		if reflect.ValueOf(v).Len() == 0 {
			return true, nil // empty array does not reflect in the schema
		}

		if err := traverseArray(sch[k], f, v); err != nil {
			return false, err
		}

		if f.Items == nil {
			return true, nil // empty object
		}
	case sch[k] != nil:
		nt, nf, err := extendedType(sch[k].Type, sch[k].Format, t, format)
		if err != nil {
			return false, ErrIncompatibleSchema
		}

		f.Type = nt
		f.Format = nf
	}

	return false, nil
}

func traverseFields(sch map[string]*schema.Field, fields map[string]any, autoGen []string) error {
	for k, v := range fields {
		// handle `null` JSON value
		if v == nil {
			continue
		}

		t, format, err := translateType(v)
		if err != nil {
			return err
		}

		f := &schema.Field{Type: t, Format: format}

		skip, err := traverseFieldsLow(t, format, k, f, v, sch)
		if err != nil {
			return err
		}

		if skip {
			continue
		}

		setAutoGenerate(autoGen, k, f)

		sch[k] = f
	}

	return nil
}

func docToSchema(sch *schema.Schema, name string, data []byte, pk []string, autoGen []string) error {
	var m map[string]interface{}

	dec := json.NewDecoder(bytes.NewBuffer(data))
	dec.UseNumber()

	if err := dec.Decode(&m); err != nil {
		return err
	}

	sch.Name = name
	if pk != nil {
		sch.PrimaryKey = pk
	}

	if sch.Fields == nil {
		sch.Fields = make(map[string]*schema.Field)
	}

	if err := traverseFields(sch.Fields, m, autoGen); err != nil {
		return err
	}

	// Implicit "id" primary key
	f := sch.Fields["id"]
	if sch.PrimaryKey == nil && f != nil && f.Format == formatUUID {
		f.AutoGenerate = true
		sch.PrimaryKey = []string{"id"}
	}

	return nil
}

func Infer(sch *schema.Schema, name string, docs []json.RawMessage, primaryKey []string, autoGenerate []string,
	depth int,
) error {
	for i := 0; (depth == 0 || i < depth) && i < len(docs); i++ {
		err := docToSchema(sch, name, docs[i], primaryKey, autoGenerate)
		if err != nil {
			return err
		}
	}

	return nil
}
