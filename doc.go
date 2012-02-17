package main

import (
	"fmt"
	"os"
	"regexp"
	"io/ioutil"
	"litebrite"
	"text/template"
	"strings"
)

var match = regexp.MustCompile("^\\s*//[^\n]\\s?")
var t = template.Must(template.ParseFiles("doc.templ"))

type section struct {
	Doc string
	Code string
}

func extractSections(source string) []section {
	sections := make([]section, 0)
	var current section
	
	for _, line := range strings.Split(source, "\n") {
		if match.FindString(line) != "" {
			if current.Code != "" {
				sections = append(sections, current)
				current = section{}
			}
			repl := match.ReplaceAllString(line, "")
			current.Doc += repl + "\n"
		} else {
			current.Code += line + "\n"
		}
	}

	return append(sections, current)
}

type File struct {
	Title string
	Sections []section
}

func main() {
	files := os.Args[1:]
	for _, filename := range files {
		src, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		}

		h := new(litebrite.Highlighter)
		h.IdentClass = "ident"
		h.LiteralClass = "literal"
		h.KeywordClass = "keyword"
		h.OperatorClass = "operator"
		out := h.Highlight(string(src))

		sections := extractSections(out)
		errt := t.Execute(os.Stdout, File{filename, sections})
		if errt != nil {
			fmt.Fprintf(os.Stderr, errt.Error())
		}
	}
}
