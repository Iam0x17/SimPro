// Copyright 2022 Dolthub, Inc.
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

package plan

import (
	"strings"

	"SimPro/pkg/go-mysql-server/memory"
	"SimPro/pkg/go-mysql-server/sql"
)

// DualTableName is empty string because no table with empty name can be created
const DualTableName = ""

// IsDualTable returns whether the given table is the "dual" table.
func IsDualTable(t sql.Table) bool {
	if t == nil {
		return false
	}
	return strings.ToLower(t.Name()) == DualTableName && t.Schema().Equals(memory.DualTableSchema.Schema)
}

var dualTable = func() sql.Table {
	return memory.NewDualTable()
}()

// NewDualSqlTable creates a new Dual table.
func NewDualSqlTable() sql.Table {
	return dualTable
}
