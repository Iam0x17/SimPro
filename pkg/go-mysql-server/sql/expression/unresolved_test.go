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

package expression

import (
	"testing"

	"github.com/stretchr/testify/require"

	"SimPro/pkg/go-mysql-server/sql"
)

func TestUnresolvedExpression(t *testing.T) {
	require := require.New(t)
	var e sql.Expression = NewUnresolvedColumn("test_col")
	require.NotNil(e)
	var o sql.Expression = NewEquals(e, e)
	require.NotNil(o)
	o = NewNot(e)
	require.NotNil(o)
}
