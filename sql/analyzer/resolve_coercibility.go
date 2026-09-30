// Copyright 2026 Dolthub, Inc.
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
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
)

// resolveCollationCoercibility resolves and stores coercibility on
// each [sql.CollationCoercibilityResolver] in |n|.
func resolveCollationCoercibility(ctx *sql.Context, a *Analyzer, n sql.Node, scope *plan.Scope, sel RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	var err error
	transform.InspectExpressions(ctx, n, func(ctx *sql.Context, e sql.Expression) bool {
		if cr, ok := e.(sql.CollationCoercibilityResolver); ok {
			if rErr := cr.ResolveCollationCoercibility(ctx); rErr != nil {
				err = rErr
				return false
			}
		}
		return true
	})
	if err != nil {
		return nil, transform.SameTree, err
	}
	return n, transform.SameTree, nil
}
