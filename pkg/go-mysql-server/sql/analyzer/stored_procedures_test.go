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

package analyzer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"SimPro/pkg/go-mysql-server/memory"
	"SimPro/pkg/go-mysql-server/sql"
	"SimPro/pkg/go-mysql-server/sql/plan"
	"SimPro/pkg/go-mysql-server/sql/transform"
)

func TestStoredProcedureNotFoundWithNoDatabaseSelected(t *testing.T) {
	db := memory.NewDatabase("mydb")
	a := NewBuilder(sql.NewDatabaseProvider(db)).Build()
	ctx := sql.NewContext(context.Background(), sql.WithSession(sql.NewBaseSession()))

	call := plan.NewCall(nil, "non_existent_procedure", []sql.Expression{}, nil, nil)
	scope, err := loadStoredProcedures(ctx, a, call, newTestScope(call), DefaultRuleSelector)
	require.NoError(t, err)

	node, identity, err := applyProceduresCall(ctx, a, call, scope, DefaultRuleSelector, nil)
	assert.Nil(t, node)
	assert.Equal(t, transform.SameTree, identity)
	assert.Contains(t, err.Error(), "stored procedure \"non_existent_procedure\" does not exist")
	assert.Contains(t, err.Error(), "this might be because no database is selected")
}
