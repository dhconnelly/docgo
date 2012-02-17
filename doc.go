package main

import (
	"fmt"
	"os"
	"regexp"
	//"github.com/dhconnelly/blackfriday"
	"io/ioutil"
	"litebrite"
)

var match = regexp.MustCompile("^\\s*//[^\n]")

func main() {
	files := os.Args[1:]
	for _, filename := range files {
		src, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		}
		fmt.Println(litebrite.Highlight(string(src)))
	}
}
