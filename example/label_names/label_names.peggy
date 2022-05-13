{
// Test labels with the same name but in different choice branches.
// 	peggy -o label_names.go label_names.peggy
package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/eaburns/peggy/peg"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		p, err := _NewParser(line)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if pos, perr := _ExprAccepts(p, 0); pos < 0 {
			_, fail := _ExprFail(p, 0 ,perr)
			fmt.Println(peg.SimpleError(line, fail))
			continue
		}
		_, result := _ExprAction(p, 0)
		fmt.Println(*result)
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
}

Expr <-
	letter:[a] { fmt.Printf("a=[%s]\n", letter); return string(letter) } /
	letter:[b] { fmt.Printf("b=[%s]\n", letter); return string(letter)  }
