// Copyright 2018 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package testutils

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/sql/coltypes"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/types"
)

// ParseType parses a string describing a type.
// It supports tuples using the syntax "tuple{<type>, <type>, ...}" but does not
// support tuples of tuples.
func ParseType(typeStr string) (types.T, error) {
	// Special case for tuples for which there is no SQL syntax.
	if strings.HasPrefix(typeStr, "tuple{") && strings.HasSuffix(typeStr, "}") {
		s := strings.TrimPrefix(typeStr, "tuple{")
		s = strings.TrimSuffix(s, "}")
		// Hijack the PREPARE syntax which takes a list of types.
		// TODO(radu): this won't work for tuples of tuples; we would need to add
		// some special syntax.
		parsed, err := parser.ParseOne(fmt.Sprintf("PREPARE x ( %s ) AS SELECT 1", s))
		if err != nil {
			return nil, fmt.Errorf("cannot parse %s as a type: %s", typeStr, err)
		}
		colTypes := parsed.AST.(*tree.Prepare).Types
		res := types.TTuple{Types: make([]types.T, len(colTypes))}
		for i := range colTypes {
			res.Types[i] = coltypes.CastTargetToDatumType(colTypes[i])
		}
		return res, nil
	}
	colType, err := parser.ParseType(typeStr)
	if err != nil {
		return nil, err
	}
	return coltypes.CastTargetToDatumType(colType), nil
}

// ParseTypes parses a list of types.
func ParseTypes(colStrs []string) ([]types.T, error) {
	res := make([]types.T, len(colStrs))
	for i, s := range colStrs {
		var err error
		res[i], err = ParseType(s)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

// ParseScalarExpr parses a scalar expression and converts it to a
// tree.TypedExpr.
func ParseScalarExpr(sql string, ivc tree.IndexedVarContainer) (tree.TypedExpr, error) {
	expr, err := parser.ParseExpr(sql)
	if err != nil {
		return nil, err
	}

	sema := tree.MakeSemaContext()
	sema.IVarContainer = ivc

	return expr.TypeCheck(&sema, types.Any)
}

// GetTestFiles returns the set of test files that matches the Glob pattern.
func GetTestFiles(tb testing.TB, testdataGlob string) []string {
	paths, err := filepath.Glob(testdataGlob)
	if err != nil {
		tb.Fatal(err)
	}
	if len(paths) == 0 {
		tb.Fatalf("no testfiles found matching: %s", testdataGlob)
	}
	return paths
}

var _ = GetTestFiles
