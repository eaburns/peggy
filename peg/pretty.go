// Copyright 2018 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package peg

import (
	"bytes"
	"io"
	"strconv"
)

type nodeOrFail interface {
	name() string
	numKids() int
	kid(int) nodeOrFail
	text() string
}

func (f *Node) name() string         { return f.Name }
func (f *Node) numKids() int         { return len(f.Kids) }
func (f *Node) kid(i int) nodeOrFail { return f.Kids[i] }
func (f *Node) text() string         { return f.Text }
func (f *Fail) name() string         { return f.Name }
func (f *Fail) numKids() int         { return len(f.Kids) }
func (f *Fail) kid(i int) nodeOrFail { return f.Kids[i] }
func (f *Fail) text() string         { return f.Want }

// Pretty returns a human-readable string of a Node or Fail
// and the subtree beneath it.
// The output looks like:
// 	<n.Name>{
// 		<Pretty(n.Kids)[0])>,
// 		<Pretty(n.Kids[1])>,
// 		…
// 		<Pretty(n.Kids[n-1])>,
// 	}
func Pretty(n nodeOrFail) string {
	b := bytes.NewBuffer(nil)
	PrettyWrite(b, n)
	return b.String()
}

// PrettyWrite is like Pretty but outputs to an io.Writer.
func PrettyWrite(w io.Writer, n nodeOrFail) error {
	return prettyWrite(w, "", n)
}

func prettyWrite(w io.Writer, tab string, n nodeOrFail) error {
	if _, err := io.WriteString(w, tab); err != nil {
		return err
	}
	if n.numKids() == 0 {
		if n.name() != "" {
			if _, err := io.WriteString(w, n.name()+"("); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(w, `"`+n.text()+`"`); err != nil {
			return err
		}
		if n.name() != "" {
			if _, err := io.WriteString(w, ")"); err != nil {
				return err
			}
		}
		return nil
	}
	if _, err := io.WriteString(w, n.name()); err != nil {
		return err
	}
	if f, ok := n.(*Fail); ok {
		pos := "[" + strconv.Itoa(f.Pos) + "]"
		if _, err := io.WriteString(w, pos); err != nil {
			return err
		}
	}
	if n.numKids() == 0 {
		if n.name() == "" {
			if _, err := io.WriteString(w, "{}"); err != nil {
				return err
			}
		}
		return nil
	}
	if _, err := io.WriteString(w, "{"); err != nil {
		return err
	}
	if n.numKids() == 1 && n.kid(0).numKids() == 0 {
		if err := prettyWrite(w, "", n.kid(0)); err != nil {
			return err
		}
		if _, err := io.WriteString(w, "}"); err != nil {
			return err
		}
		return nil
	}
	for i := 0; i < n.numKids(); i++ {
		if _, err := io.WriteString(w, "\n"); err != nil {
			return err
		}
		if err := prettyWrite(w, tab+"\t", n.kid(i)); err != nil {
			return err
		}
		if _, err := io.WriteString(w, ","); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, "\n"+tab+"}"); err != nil {
		return err
	}
	return nil
}
