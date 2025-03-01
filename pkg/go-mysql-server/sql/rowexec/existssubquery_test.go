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

package rowexec

import (
	"testing"

	"github.com/stretchr/testify/require"

	"SimPro/pkg/go-mysql-server/memory"
	"SimPro/pkg/go-mysql-server/sql"
	"SimPro/pkg/go-mysql-server/sql/expression"
	"SimPro/pkg/go-mysql-server/sql/plan"
	"SimPro/pkg/go-mysql-server/sql/types"
)

func TestExistsSubquery(t *testing.T) {
	db := memory.NewDatabase("test")
	pro := memory.NewDBProvider(db)
	ctx := newContext(pro)

	table := memory.NewTable(db.BaseDatabase, "foo", sql.NewPrimaryKeySchema(sql.Schema{
		{Name: "t", Source: "foo", Type: types.Text},
	}), nil)

	require.NoError(t, table.Insert(ctx, sql.Row{"one"}))
	require.NoError(t, table.Insert(ctx, sql.Row{"two"}))
	require.NoError(t, table.Insert(ctx, sql.Row{"three"}))

	emptyTable := memory.NewTable(db.BaseDatabase, "empty", sql.NewPrimaryKeySchema(sql.Schema{
		{Name: "t", Source: "empty", Type: types.Int64},
	}), nil)

	project := func(expr sql.Expression, tbl *memory.Table) sql.Node {
		return plan.NewProject([]sql.Expression{
			expr,
		}, plan.NewResolvedTable(tbl, nil, nil))
	}

	testCases := []struct {
		name     string
		subquery sql.Node
		row      sql.Row
		result   interface{}
	}{
		{
			"Null returns as true",
			project(
				expression.NewGetField(1, types.Text, "foo", false), table,
			),
			sql.NewRow(nil),
			true,
		},
		{
			"Non NULL evaluates as true",
			project(
				expression.NewGetField(1, types.Text, "foo", false), table,
			),
			sql.NewRow("four"),
			true,
		},
		{
			"Empty Set Passes",
			project(
				expression.NewGetField(1, types.Text, "foo", false), emptyTable,
			),
			sql.NewRow(),
			false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			result, err := plan.NewExistsSubquery(
				plan.NewSubquery(tt.subquery, "").WithExecBuilder(DefaultBuilder),
			).Eval(ctx, tt.row)
			require.NoError(err)
			require.Equal(tt.result, result)

			// Test Not Exists
			result, err = expression.NewNot(plan.NewExistsSubquery(
				plan.NewSubquery(tt.subquery, "").WithExecBuilder(DefaultBuilder),
			)).Eval(ctx, tt.row)

			require.NoError(err)
			require.Equal(tt.result, !result.(bool))
		})
	}
}
