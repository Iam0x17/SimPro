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
	"SimPro/pkg/go-mysql-server/sql/plan"
	"SimPro/pkg/go-mysql-server/test"
)

func TestShowTableStatus(t *testing.T) {
	require := require.New(t)

	db1 := memory.NewDatabase("a")
	db1.AddTable("t1", memory.NewTable(db1, "t1", sql.PrimaryKeySchema{}, db1.GetForeignKeyCollection()))
	db1.AddTable("t2", memory.NewTable(db1, "t2", sql.PrimaryKeySchema{}, db1.GetForeignKeyCollection()))

	db2 := memory.NewDatabase("b")
	db2.AddTable("t3", memory.NewTable(db2, "t3", sql.PrimaryKeySchema{}, db2.GetForeignKeyCollection()))
	db2.AddTable("t4", memory.NewTable(db2, "t4", sql.PrimaryKeySchema{}, db2.GetForeignKeyCollection()))

	catalog := test.NewCatalog(sql.NewDatabaseProvider(db1, db2))
	pro := memory.NewDBProvider(db1, db2)
	ctx := newContext(pro)

	node := plan.NewShowTableStatus(db1)
	node.Catalog = catalog

	ctx.SetCurrentDatabase("a")
	iter, err := DefaultBuilder.Build(ctx, node, nil)
	require.NoError(err)

	rows, err := sql.RowIterToRows(ctx, iter)
	require.NoError(err)

	expected := []sql.Row{
		{"t1", "InnoDB", "10", "Fixed", uint64(0), uint64(0), uint64(0), uint64(0), int64(0), int64(0), nil, nil, nil, nil, sql.Collation_Default.String(), nil, nil, nil},
		{"t2", "InnoDB", "10", "Fixed", uint64(0), uint64(0), uint64(0), uint64(0), int64(0), int64(0), nil, nil, nil, nil, sql.Collation_Default.String(), nil, nil, nil},
	}

	require.ElementsMatch(expected, rows)
	node = plan.NewShowTableStatus(db2)
	node.Catalog = catalog

	iter, err = DefaultBuilder.Build(ctx, node, nil)
	require.NoError(err)

	rows, err = sql.RowIterToRows(ctx, iter)
	require.NoError(err)

	expected = []sql.Row{
		{"t3", "InnoDB", "10", "Fixed", uint64(0), uint64(0), uint64(0), uint64(0), int64(0), int64(0), nil, nil, nil, nil, sql.Collation_Default.String(), nil, nil, nil},
		{"t4", "InnoDB", "10", "Fixed", uint64(0), uint64(0), uint64(0), uint64(0), int64(0), int64(0), nil, nil, nil, nil, sql.Collation_Default.String(), nil, nil, nil},
	}

	require.ElementsMatch(expected, rows)
}
