// Copyright 2022 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConvertValue(t *testing.T) {
	kase := []struct {
		val string
		typ string
	}{
		{"1", "int"},
		{"1", "tinyint"},
		{"1", "smallint"},
		{"1", "bigint"},
		{"1", "unsigned bigint"},
		{"1", "unsigned int"},
		{"1", "unsigned tinyint"},
		{"1", "unsigned smallint"},
		{"1.1", "float"},
		{"1.1", "double"},
		{"1.1", "decimal"},
		{"asa", "varchar"},
		{"asa", "char"},
		{"asa", "text"},
		{"asa", "blob"},
		{"asa", "uuid"},
		{"asa", "json"},
		{"2021-01-01", "date"},
		{"2021-01-01 00:00:00", "datetime"},
		{"2021-01-01 00:00:00", "timestamp"},
	}
	for _, v := range kase {
		s := convertValue(makeValue(v.val), v.typ)
		switch v.typ {
		case "int", "tinyint", "smallint", "bigint", "unsigned bigint", "unsigned int", "unsigned tinyint", "unsigned smallint", "float", "double":
			require.Equal(t, v.val, s)
		default:

			require.Equal(t, fmt.Sprintf("'%v'", v.val), s)
		}
	}
}

func makeValue(val string) interface{} {
	tmp := []byte(val)
	var ret interface{} = tmp
	return &ret
}
