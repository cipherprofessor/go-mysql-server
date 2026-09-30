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
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
)

// Elt joins several strings together.
type Elt struct {
	args         []sql.Expression
	collation    sql.CollationID
	coercibility byte
}

var _ sql.FunctionExpression = (*Elt)(nil)
var _ sql.CollationCoercible = (*Elt)(nil)
var _ sql.CollationCoercibilityResolver = (*Elt)(nil)

// NewElt creates a new Elt UDF.
func NewElt(ctx *sql.Context, args ...sql.Expression) (sql.Expression, error) {
	if len(args) < 2 {
		return nil, sql.ErrInvalidArgumentNumber.New("ELT", "2 or more", len(args))
	}

	return &Elt{args: args}, nil
}

// ResolveCollationCoercibility implements
// [sql.CollationCoercibilityResolver] over |args|[1:].
func (e *Elt) ResolveCollationCoercibility(ctx *sql.Context) error {
	collation, coercibility, err := sql.ResolveCoercibilityExpressions(ctx, sql.CollationAggregationDefault, e.args[1:]...)
	if err != nil {
		return err
	}
	e.collation = collation
	e.coercibility = coercibility
	return nil
}

// FunctionName implements sql.FunctionExpression
func (e *Elt) FunctionName() string {
	return "elt"
}

// Description implements sql.FunctionExpression
func (e *Elt) Description() string {
	return "returns the string at index number."
}

// Type implements the Expression interface.
func (e *Elt) Type(ctx *sql.Context) sql.Type {
	return types.LongText
}

// CollationCoercibility implements the interface sql.CollationCoercible.
func (e *Elt) CollationCoercibility(ctx *sql.Context) (sql.CollationID, byte) {
	return e.collation, e.coercibility
}

// IsNullable implements the Expression interface.
func (e *Elt) IsNullable(ctx *sql.Context) bool {
	return true
}

// String implements the Stringer interface.
func (e *Elt) String() string {
	var args = make([]string, len(e.args))
	for i, arg := range e.args {
		args[i] = arg.String()
	}
	return fmt.Sprintf("%s(%s)", e.FunctionName(), strings.Join(args, ","))
}

// WithChildren implements the Expression interface.
func (e *Elt) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) < 2 {
		return nil, sql.ErrInvalidArgumentNumber.New("ELT", "2 or more", len(children))
	}
	return &Elt{
		args:         children,
		collation:    e.collation,
		coercibility: e.coercibility,
	}, nil
}

// Resolved implements the Expression interface.
func (e *Elt) Resolved() bool {
	for _, arg := range e.args {
		if !arg.Resolved() {
			return false
		}
	}
	return true
}

// Children implements the Expression interface.
func (e *Elt) Children() []sql.Expression {
	return e.args
}

// Eval implements the Expression interface.
func (e *Elt) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	indexInt, ok, err := evalInt64(ctx, e.args[0], row)
	if err != nil || !ok {
		return nil, err
	}

	idx := int(indexInt)
	if idx <= 0 || idx >= len(e.args) {
		return nil, nil
	}

	str, err := e.args[idx].Eval(ctx, row)
	if err != nil {
		return nil, err
	}

	res, _, err := types.Text.Convert(ctx, str)
	if err != nil {
		return nil, err
	}

	return res, nil
}
