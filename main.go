// Copyright 2017 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

//go:generate goyacc -o grammar.go -p "peggy" grammar.y

var (
	out    = flag.String("o", "", "output file path")
	prefix = flag.String("p", "_", "identifier prefix")
)

func main() {
	flag.Parse()
	args := flag.Args()

	in := bufio.NewReader(os.Stdin)
	file := "<stdin>"
	if len(args) > 0 {
		f, err := os.Open(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		in = bufio.NewReader(f)
		file = args[0]
	}

	g, err := Parse(in, file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := Check(g); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	w := bufio.NewWriter(os.Stdout)
	if *out != "" {
		f, err := os.Create(*out)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer func() {
			if err := f.Close(); err != nil {
				fmt.Println(err)
			}
		}()
		w = bufio.NewWriter(f)
	}

	cfg := Config{Prefix: *prefix}
	if err := cfg.Generate(w, file, g); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := w.Flush(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
