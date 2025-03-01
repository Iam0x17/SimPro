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

package analyzer

import (
	"fmt"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/src-d/go-errors.v1"

	"SimPro/pkg/go-mysql-server/sql"
	"SimPro/pkg/go-mysql-server/sql/expression"
	"SimPro/pkg/go-mysql-server/sql/plan"
	"SimPro/pkg/go-mysql-server/sql/types"
)

func col(idx int, table, col string) sql.Expression {
	return expression.NewGetFieldWithTable(0, idx, types.Int64, "", table, col, false)
}

func and(left, right sql.Expression) sql.Expression {
	return expression.NewAnd(left, right)
}

func gt(left, right sql.Expression) sql.Expression {
	return expression.NewGreaterThan(left, right)
}

func lt(left, right sql.Expression) sql.Expression {
	return expression.NewLessThan(left, right)
}

func or(left, right sql.Expression) sql.Expression {
	return expression.NewOr(left, right)
}

func eq(left, right sql.Expression) sql.Expression {
	return expression.NewEquals(left, right)
}

func lit(n int64) sql.Expression {
	return expression.NewLiteral(n, types.Int64)
}

func litT(n interface{}, t sql.Type) sql.Expression {
	return expression.NewLiteral(n, t)
}

func gf(idx int, table, name string) *expression.GetField {
	return expression.NewGetFieldWithTable(0, idx, types.Int64, "", table, name, false)
}

func gfCol(idx int, col *sql.Column) *expression.GetField {
	return expression.NewGetFieldWithTable(0, idx, col.Type, col.DatabaseSource, col.Source, col.Name, true)
}

func gfColAlias(idx int, col *sql.Column, tableAlias string) *expression.GetField {
	return expression.NewGetFieldWithTable(0, idx, col.Type, col.DatabaseSource, tableAlias, col.Name, true)
}

func uc(name string) *expression.UnresolvedColumn {
	return expression.NewUnresolvedColumn(name)
}

func uqc(table, name string) *expression.UnresolvedColumn {
	return expression.NewUnresolvedQualifiedColumn(table, name)
}

func null(col sql.Expression) sql.Expression {
	return expression.NewIsNull(col)
}

func litNull() sql.Expression {
	return expression.NewLiteral(nil, types.Null)
}

// Creates a new top-level scope from the node given
func newTestScope(n sql.Node) *plan.Scope {
	return (*plan.Scope)(nil).NewScope(n)
}

var analyzeRules = [][]Rule{
	OnceBeforeDefault,
	DefaultRules,
	OnceAfterDefault,
	OnceAfterAll,
}

func getRule(id RuleId) Rule {
	for _, rules := range analyzeRules {
		rule := getRuleFrom(rules, id)
		if rule != nil {
			return *rule
		}
	}

	panic("missing rule")
}

func getRuleFrom(rules []Rule, id RuleId) *Rule {
	for _, rule := range rules {
		if rule.Id == id {
			return &rule
		}
	}

	return nil
}

// Common test struct for analyzer transformation tests. Name and node are required, other fields are optional.
// The expected node is optional: if omitted, the tests asserts that input == output. The optional err field is the
// kind of error expected, if any.
type analyzerFnTestCase struct {
	name     string
	node     sql.Node
	scope    *plan.Scope
	expected sql.Node
	err      *errors.Kind
}

func runTestCases(t *testing.T, ctx *sql.Context, testCases []analyzerFnTestCase, a *Analyzer, f ...Rule) {
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			context := ctx
			if context == nil {
				context = sql.NewEmptyContext()
			}

			var result = tt.node
			var err error
			for _, r := range f {
				result, _, err = r.Apply(context, a, result, tt.scope, DefaultRuleSelector, nil)
				if tt.err != nil {
					require.Error(t, err)
					require.True(t, tt.err.Is(err), fmt.Sprintf("Expected error of type %T but got %T", tt.err, err))
					return
				}
				require.NoError(t, err)
			}

			expected := tt.expected
			if expected == nil {
				expected = tt.node
			}

			assertNodesEqualWithDiff(t, expected, result)
		})
	}
}

// assertNodesEqualWithDiff asserts the two nodes given to be equal and prints any diff according to their DebugString
// methods.
func assertNodesEqualWithDiff(t *testing.T, expected, actual sql.Node) bool {
	if !assert.Equal(t, expected, actual) {
		expectedStr := sql.DebugString(expected)
		actualStr := sql.DebugString(actual)
		diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(expectedStr),
			B:        difflib.SplitLines(actualStr),
			FromFile: "expected",
			FromDate: "",
			ToFile:   "actual",
			ToDate:   "",
			Context:  1,
		})
		require.NoError(t, err)

		if len(diff) > 0 {
			fmt.Println(diff)
		} else {
			// No textual diff found, but not equal. Ugh. Let's at least figure out which node in the plans isn't equal.
		Top:
			for {
				for i := range expected.Children() {
					if !assertNodesEqualWithDiff(t, expected.Children()[i], actual.Children()[i]) {
						expected, actual = expected.Children()[i], actual.Children()[i]
						continue Top
					}
				}
				// Either no children, or all children were equal. This must the node that's different. Probably should add
				// enough information in DebugPrint for this node that it shows up in the textual diff.
				fmt.Printf("Non-textual difference found in node %s -- implement a better DebugPrint?\n", sql.DebugString(expected))
				break
			}
		}

		return false
	}
	return true
}
