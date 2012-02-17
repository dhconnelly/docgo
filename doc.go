package main

import (
	"fmt"
	"os"
	"regexp"
	"io/ioutil"
	"litebrite"
	"text/template"
)

var match = regexp.MustCompile("^\\s*//[^\n]")
var t = template.Must(template.ParseFiles("doc.templ"))

type File struct {
	Title string
	Source string
}

func main() {
	files := os.Args[1:]
	for _, filename := range files {
		src, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		}
		
		out := litebrite.Highlight(string(src))
		t.Execute(os.Stdout, File{filename, out})
	}
}
