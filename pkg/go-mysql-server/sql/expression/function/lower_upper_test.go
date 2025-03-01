// Copyright 2020-2021 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package function

import (
	"testing"

	"github.com/stretchr/testify/require"

	"SimPro/pkg/go-mysql-server/sql"
	"SimPro/pkg/go-mysql-server/sql/expression"
	"SimPro/pkg/go-mysql-server/sql/types"
)

func TestLower(t *testing.T) {
	testCases := []struct {
		name     string
		rowType  sql.Type
		row      sql.Row
		expected interface{}
	}{
		{"text nil", types.LongText, sql.NewRow(nil), nil},
		{"text ok", types.LongText, sql.NewRow("LoWeR"), "lower"},
		{"binary ok", types.Blob, sql.NewRow([]byte("LoWeR")), "LoWeR"},
		{"other type", types.Int32, sql.NewRow(int32(1)), "1"},
	}

	for _, tt := range testCases {
		f := NewLower(expression.NewGetField(0, tt.rowType, "", true))

		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, eval(t, f, tt.row))
		})

		req := require.New(t)
		req.True(f.IsNullable())
		req.Equal(tt.rowType, f.Type())
	}
}

func TestUpper(t *testing.T) {
	testCases := []struct {
		name     string
		rowType  sql.Type
		row      sql.Row
		expected interface{}
	}{
		{"text nil", types.LongText, sql.NewRow(nil), nil},
		{"text ok", types.LongText, sql.NewRow("UpPeR"), "UPPER"},
		{"binary ok", types.Blob, sql.NewRow([]byte("UpPeR")), "UpPeR"},
		{"other type", types.Int32, sql.NewRow(int32(1)), "1"},
	}

	for _, tt := range testCases {
		f := NewUpper(expression.NewGetField(0, tt.rowType, "", true))

		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, eval(t, f, tt.row))
		})

		req := require.New(t)
		req.True(f.IsNullable())
		req.Equal(tt.rowType, f.Type())
	}
}
