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

func TestIfNull(t *testing.T) {
	testCases := []struct {
		expression interface{}
		value      interface{}
		expected   interface{}
	}{
		{"foo", "bar", "foo"},
		{"foo", "foo", "foo"},
		{nil, "foo", "foo"},
		{"foo", nil, "foo"},
		{nil, nil, nil},
		{"", nil, ""},
	}

	f := NewIfNull(
		expression.NewGetField(0, types.LongText, "expression", true),
		expression.NewGetField(1, types.LongText, "value", true),
	)
	require.Equal(t, types.LongText, f.Type())

	for _, tc := range testCases {
		v, err := f.Eval(sql.NewEmptyContext(), sql.NewRow(tc.expression, tc.value))
		require.NoError(t, err)
		require.Equal(t, tc.expected, v)
	}
}
