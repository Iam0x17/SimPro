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

package aggregation

import (
	"testing"

	"github.com/stretchr/testify/require"

	"SimPro/pkg/go-mysql-server/sql"
	"SimPro/pkg/go-mysql-server/sql/expression"
)

func TestSum(t *testing.T) {
	sum := NewSum(expression.NewGetField(0, nil, "", false))

	testCases := []struct {
		name     string
		rows     []sql.Row
		expected interface{}
	}{
		{
			"string int values",
			[]sql.Row{{"1"}, {"2"}, {"3"}, {"4"}},
			float64(10),
		},
		{
			"string float values",
			[]sql.Row{{"1.5"}, {"2"}, {"3"}, {"4"}},
			float64(10.5),
		},
		{
			"string non-int values",
			[]sql.Row{{"a"}, {"b"}, {"c"}, {"d"}},
			float64(0),
		},
		{
			"float values",
			[]sql.Row{{1.}, {2.5}, {3.}, {4.}},
			float64(10.5),
		},
		{
			"no rows",
			[]sql.Row{},
			nil,
		},
		{
			"nil values",
			[]sql.Row{{nil}, {nil}},
			nil,
		},
		{
			"int64 values",
			[]sql.Row{{int64(1)}, {int64(3)}},
			float64(4),
		},
		{
			"int32 values",
			[]sql.Row{{int32(1)}, {int32(3)}},
			float64(4),
		},
		{
			"int32 and nil values",
			[]sql.Row{{int32(1)}, {int32(3)}, {nil}},
			float64(4),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			ctx := sql.NewEmptyContext()
			buf, _ := sum.NewBuffer()
			for _, row := range tt.rows {
				require.NoError(buf.Update(ctx, row))
			}

			result, err := buf.Eval(sql.NewEmptyContext())
			require.NoError(err)
			require.Equal(tt.expected, result)
		})
	}
}

func TestSumWithDistinct(t *testing.T) {
	require := require.New(t)

	ad := expression.NewDistinctExpression(expression.NewGetField(0, nil, "myfield", false))
	sum := NewSum(ad)

	// first validate that the expression's name is correct
	require.Equal("SUM(DISTINCT myfield)", sum.String())

	testCases := []struct {
		name     string
		rows     []sql.Row
		expected interface{}
	}{
		{
			"string int values",
			[]sql.Row{{"1"}, {"1"}, {"2"}, {"2"}, {"3"}, {"3"}, {"4"}, {"4"}},
			float64(10),
		},
		// TODO : DISTINCT returns incorrect result, it currently returns 11.00
		//        https://github.com/dolthub/dolt/issues/4298
		//{
		//	"string int values",
		//	[]sql.Row{{"1.00"}, {"1"}, {"2"}, {"2"}, {"3"}, {"3"}, {"4"}, {"4"}},
		//	float64(10),
		//},
		{
			"string float values",
			[]sql.Row{{"1.5"}, {"1.5"}, {"1.5"}, {"1.5"}, {"2"}, {"3"}, {"4"}},
			float64(10.5),
		},
		{
			"string non-int values",
			[]sql.Row{{"a"}, {"b"}, {"b"}, {"c"}, {"c"}, {"d"}},
			float64(0),
		},
		{
			"float values",
			[]sql.Row{{1.}, {2.5}, {3.}, {4.}},
			float64(10.5),
		},
		{
			"no rows",
			[]sql.Row{},
			nil,
		},
		{
			"nil values",
			[]sql.Row{{nil}, {nil}},
			nil,
		},
		{
			"int64 values",
			[]sql.Row{{int64(1)}, {int64(3)}, {int64(3)}, {int64(3)}},
			float64(4),
		},
		{
			"int32 values",
			[]sql.Row{{int32(1)}, {int32(1)}, {int32(1)}, {int32(3)}},
			float64(4),
		},
		{
			"int32 and nil values",
			[]sql.Row{{nil}, {int32(1)}, {int32(1)}, {int32(1)}, {int32(3)}, {nil}, {nil}},
			float64(4),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ad.Dispose()

			ctx := sql.NewEmptyContext()
			buf, _ := sum.NewBuffer()
			for _, row := range tt.rows {
				require.NoError(buf.Update(ctx, row))
			}

			result, err := buf.Eval(sql.NewEmptyContext())
			require.NoError(err)
			require.Equal(tt.expected, result)
		})
	}
}
