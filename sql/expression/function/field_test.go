// Copyright 2024 Dolthub, Inc.
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

	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/types"
)

func TestField(t *testing.T) {
	tests := []struct {
		name string
		args []sql.Expression
		exp  interface{}
		err  bool
		skip bool
	}{
		{
			name: "null argument",
			args: []sql.Expression{
				nil,
				nil,
			},
			exp: int64(0),
		},
		{
			name: "not found",
			args: []sql.Expression{
				expression.NewLiteral("abc", types.Text),
				expression.NewLiteral("def", types.Text),
			},
			exp: int64(0),
		},
		{
			name: "simple case",
			args: []sql.Expression{
				expression.NewLiteral("abc", types.Int32),
				expression.NewLiteral("abc", types.Int32),
				expression.NewLiteral("def", types.Int32),
				expression.NewLiteral("xyz", types.Int32),
			},
			exp: int64(1),
		},
		{
			name: "simple case again",
			args: []sql.Expression{
				expression.NewLiteral("def", types.Int32),
				expression.NewLiteral("abc", types.Int32),
				expression.NewLiteral("def", types.Int32),
				expression.NewLiteral("xyz", types.Int32),
			},
			exp: int64(2),
		},
		{
			name: "index is int",
			args: []sql.Expression{
				expression.NewLiteral(10, types.Int32),
				expression.NewLiteral("8", types.Text),
				expression.NewLiteral("8", types.Text),
				expression.NewLiteral("10", types.Text),
			},
			exp: int64(3),
		},
		{
			name: "index is float",
			args: []sql.Expression{
				expression.NewLiteral(2.9, types.Float64),
				expression.NewLiteral("1", types.Text),
				expression.NewLiteral("2.9", types.Text),
				expression.NewLiteral("3", types.Text),
			},
			exp: int64(2),
		},
		{
			name: "scientific string is truncated",
			args: []sql.Expression{
				expression.NewLiteral(1e2, types.Int32),
				expression.NewLiteral("100", types.Text),
			},
			exp: int64(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip()
			}

			ctx := sql.NewEmptyContext()
			f, err := NewField(sql.NewEmptyContext(), tt.args...)
			require.NoError(t, err)

			res, err := f.Eval(ctx, nil)
			if tt.err {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.exp, res)
		})
	}
}

func TestFieldCollationCoercibility(t *testing.T) {
	// https://github.com/dolthub/dolt/issues/11907
	ctx := sql.NewEmptyContext()
	latin1Type := types.MustCreateString(sqltypes.VarChar, 10, sql.Collation_latin1_swedish_ci)
	target := expression.NewLiteral("a", latin1Type)
	choice1 := expression.NewLiteral("b", types.LongText)
	choice2 := expression.NewLiteral("c", types.LongText)

	field, err := NewField(ctx, target, choice1, choice2)
	require.NoError(t, err)
	require.NoError(t, field.(sql.CollationCoercibilityResolver).ResolveCollationCoercibility(ctx))

	col, coer := field.(sql.CollationCoercible).CollationCoercibility(ctx)
	require.Equal(t, sql.Collation_latin1_swedish_ci, col)
	require.Equal(t, byte(4), coer)
}
