// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package view

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	fieldsDepth = 2 // depth of the nested fields to examine
)

func extractFieldValues(objs []interface{}, fields []string) ([][]interface{}, error) {
	if len(objs) == 0 {
		return [][]interface{}{}, nil
	}

	knownFields := extractFieldNames(objs[0], []string{}, "", fieldsDepth)

	if len(fields) == 0 {
		fields = knownFields
	}

	if err := validateFields(knownFields, fields); err != nil {
		return nil, err
	}

	var result = make([][]interface{}, len(objs))
	for i, item := range objs {
		result[i] = make([]interface{}, len(fields))
		val := reflect.ValueOf(item)
		for j, field := range fields {
			nestedFields := splitFieldPath(field)
			var col interface{}
			for _, nField := range nestedFields {
				val = reflect.Indirect(val)
				val = val.FieldByName(nField)
				col = val.Interface()
				val = reflect.ValueOf(col)
			}
			result[i][j] = col

			val = reflect.ValueOf(item)
		}
	}

	return result, nil
}

func extractFieldNames(obj interface{}, fieldNames []string, parentField string, depth int) []string {
	if depth == 0 {
		return fieldNames
	}

	val := reflect.ValueOf(obj)
	val = reflect.Indirect(val)
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		if !isFieldExported(typ.Field(i)) {
			continue
		}

		fieldName := typ.Field(i).Name
		if parentField != "" {
			fieldName = parentField + "." + fieldName
		}
		fieldNames = append(fieldNames, fieldName)

		// recursively examine nested fields
		subval := val.FieldByName(fieldName)
		subval = reflect.Indirect(subval)
		isFieldValid := subval.Kind() == reflect.Struct && subval.CanInterface()

		if isFieldValid {
			subObj := subval.Interface()
			fieldNames = extractFieldNames(subObj, fieldNames, fieldName, depth-1)
		}
	}
	return fieldNames
}

func validateFields(allowedFields []string, fields []string) error {
	for _, f := range fields {
		contains := false
		for _, a := range allowedFields {
			if strings.Compare(f, a) == 0 {
				contains = true
				break
			}
		}
		if !contains {
			fieldsStr := `"` + strings.Join(allowedFields, `","`) + `"`
			return fmt.Errorf("unknown field %v.\nAvailable fields: %v", f, fieldsStr)
		}
	}
	return nil
}

func isFieldExported(field reflect.StructField) bool {
	return field.PkgPath == ""
}

func splitFieldPath(field string) []string {
	return strings.Split(field, ".") // results in ex. "Execution", "RunId"
}
