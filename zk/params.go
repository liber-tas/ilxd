// Copyright (c) 2024 The illium developers
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package zk

type Parameters interface {
	// ToExpr marshals the Parameters to a string
	// expression used by lurk.
	ToExpr() string
}

type Expr string

func (p Expr) ToExpr() string {
	return string(p)
}
