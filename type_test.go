// Copyright 2018 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import (
	"strings"
	"testing"
)

func TestType(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "any",
			in:   "A <- .",
			want: "string",
		},
		{
			name: "char class",
			in:   "A <- [abc]",
			want: "string",
		},
		{
			name: "literal",
			in:   "A <- \"abc\"",
			want: "string",
		},
		{
			name: "pred code",
			in:   "A <- &{ true }",
			want: "bool",
		},
		{
			name: "subexpr",
			in:   "A <- (&{ true })",
			want: "bool",
		},
		{
			name: "ident",
			in:   "A <- B\nB <- &{ true }",
			want: "bool",
		},
		{
			name: "opt expr",
			in:   "A <- \"abc\"?",
			want: "*string",
		},
		{
			name: "rep expr",
			in:   "A <- \"abc\"+",
			want: "[]string",
		},
		{
			name: "pred expr",
			in:   "A <- &B\nB <- \"abc\"",
			want: "bool",
		},
		{
			name: "label expr",
			in:   "A <- l:(\"abc\"*)",
			want: "[]string",
		},
		{
			name: "sequence: same types",
			in:   "A <- \"abc\" \"def\"",
			want: "[]string",
		},
		{
			name: "sequence: different types",
			in:   "A <- \"abc\" &B \"def\"\nB <- \"xyz\"",
			want: "[]interface{}",
		},
		{
			name: "action",
			in:   "A <- \"abc\" T:{ return T{} }",
			want: "T",
		},
		{
			name: "action with string type",
			in:   `A <- "abc" "interface{}":{ return nil }`,
			want: "interface{}",
		},
		{
			name: "choice: same types",
			in:   "A <- \"abc\" / \"xyz\"",
			want: "string",
		},
		{
			name: "choice: different types",
			in:   "A <- \"abc\" / \"xyz\" T:{ return T{} }",
			want: "interface{}",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			in := strings.NewReader(test.in)
			g, err := Parse(in, "test.file")
			if err != nil {
				t.Errorf("Parse(%q, _)=_, %v, want _,nil", test.in, err)
				return
			}
			if err := Check(g); err != nil {
				t.Errorf("Check(%q)=%v, want nil", test.in, err)
				return
			}
			got := g.Rules[0].Expr.Type()
			if got != test.want {
				t.Errorf("%s.Type()=%s, want %s\n",
					g.Rules[0].Name, got, test.want)
			}
		})
	}
}
