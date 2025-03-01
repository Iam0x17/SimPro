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

func TestLength(t *testing.T) {
	testCases := []struct {
		name      string
		input     interface{}
		inputType sql.Type
		fn        func(sql.Expression) sql.Expression
		expected  interface{}
	}{
		{
			"length string",
			"fóo",
			types.Text,
			NewLength,
			int32(4),
		},
		{
			"length binary",
			[]byte("fóo"),
			types.Blob,
			NewLength,
			int32(4),
		},
		{
			"length empty",
			"",
			types.Blob,
			NewLength,
			int32(0),
		},
		{
			"length empty binary",
			[]byte{},
			types.Blob,
			NewLength,
			int32(0),
		},
		{
			"length nil",
			nil,
			types.Blob,
			NewLength,
			nil,
		},
		{
			"char_length string",
			"fóo",
			types.LongText,
			NewCharLength,
			int32(3),
		},
		{
			"char_length binary",
			[]byte("fóo"),
			types.Blob,
			NewCharLength,
			int32(4),
		},
		{
			"char_length empty",
			"",
			types.Blob,
			NewCharLength,
			int32(0),
		},
		{
			"char_length empty binary",
			[]byte{},
			types.Blob,
			NewCharLength,
			int32(0),
		},
		{
			"char_length nil",
			nil,
			types.Blob,
			NewCharLength,
			nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			result, err := tt.fn(expression.NewGetField(0, tt.inputType, "foo", false)).Eval(
				sql.NewEmptyContext(),
				sql.Row{tt.input},
			)

			require.NoError(err)
			require.Equal(tt.expected, result)
		})
	}
}
