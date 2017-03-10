// Copyright 2017 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import (
	"fmt"
	"sort"
)

// Located is an interface representing anything located within the input stream.
type Located interface {
	Begin() Loc
	End() Loc
}

// Errors implements error, containing multiple errors.
type Errors struct {
	Errs []Error
}

func (err *Errors) ret() error {
	if len(err.Errs) == 0 {
		return nil
	}
	sort.Slice(err.Errs, func(i, j int) bool {
		return err.Errs[i].Begin().Less(err.Errs[j].Begin())
	})
	return err
}

func (err *Errors) add(loc Located, format string, args ...interface{}) {
	err.Errs = append(err.Errs, Err(loc, format, args...))
}

// Error returns the string representation of the Errors,
// which is the string of each Error, one per-line.
func (err Errors) Error() string {
	var s string
	for i, e := range err.Errs {
		if i > 0 {
			s += "\n"
		}
		s += e.Error()
	}
	return s
}

// Error is an error tied to an element of the Peggy input file.
type Error struct {
	Located
	Msg string
}

func (err Error) Error() string {
	b, e := err.Begin(), err.End()
	l0, c0 := b.Line, b.Col
	l1, c1 := e.Line, e.Col
	switch {
	case l0 == l1 && c0 == c1:
		return fmt.Sprintf("%s:%d.%d: %s", b.File, l0, c0, err.Msg)
	default:
		return fmt.Sprintf("%s:%d.%d,%d.%d: %s", b.File, l0, c0, l1, c1, err.Msg)
	}
}

// Err returns an error containing the location and formatted message.
func Err(loc Located, format string, args ...interface{}) Error {
	return Error{Located: loc, Msg: fmt.Sprintf(format, args...)}
}
