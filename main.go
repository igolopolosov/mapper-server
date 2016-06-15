package main

import (
	"fmt"
	"os"
)

func main() {
	tpl, _ := os.Open("template.txt")
	defer tpl.Close()
	_, _ = MakeAppl(tpl)
	fmt.Println("end")
}
