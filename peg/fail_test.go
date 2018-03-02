// Copyright 2018 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package peg

import (
	"reflect"
	"testing"

	"github.com/eaburns/pretty"
)

func TestDedupFails(t *testing.T) {
	x := &Fail{Name: "x"}
	z := &Fail{Name: "z"}
	y := &Fail{Name: "y", Kids: []*Fail{z, z}}
	root := &Fail{
		Kids: []*Fail{
			x,
			&Fail{
				Kids: []*Fail{
					y,
					y,
				},
			},
			x,
		},
	}
	DedupFails(root)
	want := &Fail{
		Kids: []*Fail{
			&Fail{Name: "x"},
			&Fail{
				Kids: []*Fail{
					&Fail{
						Name: "y",
						Kids: []*Fail{
							&Fail{Name: "z"},
						},
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(root, want) {
		t.Errorf("DedupFails()=%v, want %v",
			pretty.String(root), pretty.String(want))
	}
}

func TestLeafFails(t *testing.T) {
	x0 := &Fail{Name: "x0", Pos: 10}
	x1 := &Fail{Name: "x1", Pos: 10}
	y0 := &Fail{Name: "y0", Pos: 15}
	y1 := &Fail{Name: "y1", Pos: 15}
	z0 := &Fail{Name: "z0", Pos: 20}
	z1 := &Fail{Name: "z1", Pos: 20}

	root := &Fail{
		Kids: []*Fail{
			x0,
			y0,
			z0,
			&Fail{
				Kids: []*Fail{
					x1,
					y1,
					z1,
					z0,
				},
			},
			z1,
			x0,
			y1,
		},
	}

	got := LeafFails(root)
	want := []*Fail{z0, z1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("LeafFails()=%s, want %s",
			pretty.String(got), pretty.String(want))
	}
}

func TestSimpleError_1(t *testing.T) {
	text := "123456789\nabcdefg"
	root := &Fail{
		Kids: []*Fail{
			&Fail{Pos: 10, Want: "A"},
		},
	}
	err := SimpleError(text, root)
	want := ":2.1: want A; got 'abcdefg'"
	if err.Error() != want {
		t.Errorf("err.Error()=%q, want %q", err.Error(), want)
	}
}

func TestSimpleError_2(t *testing.T) {
	text := "123456789\nabcdefg"
	root := &Fail{
		Kids: []*Fail{
			&Fail{Pos: 10, Want: "A"},
			&Fail{Pos: 10, Want: "B"},
		},
	}
	err := SimpleError(text, root)
	want := ":2.1: want A or B; got 'abcdefg'"
	if err.Error() != want {
		t.Errorf("err.Error()=%q, want %q", err.Error(), want)
	}
}

func TestSimpleError_3(t *testing.T) {
	text := "123456789\nabcdefg"
	root := &Fail{
		Kids: []*Fail{
			&Fail{Pos: 10, Want: "A"},
			&Fail{Pos: 10, Want: "B"},
			&Fail{Pos: 10, Want: "C"},
		},
	}
	err := SimpleError(text, root)
	want := ":2.1: want A, B, or C; got 'abcdefg'"
	if err.Error() != want {
		t.Errorf("err.Error()=%q, want %q", err.Error(), want)
	}
}
