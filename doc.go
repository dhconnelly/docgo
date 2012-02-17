package main

import (
	"fmt"
	"os"
	"regexp"
	"io/ioutil"
	"litebrite"
	"text/template"
	"strings"
	"blackfriday"
)

var match = regexp.MustCompile("^\\s*//[^\n]\\s?")
var t = template.Must(template.ParseFiles("doc.templ"))

type section struct {
	Doc string
	Code string
}

func extractSections(source string) []*section {
	sections := make([]*section, 0)
	current := new(section)
	
	for _, line := range strings.Split(source, "\n") {
		if match.FindString(line) != "" {
			if current.Code != "" {
				sections = append(sections, current)
				current = new(section)
			}
			repl := match.ReplaceAllString(line, "")
			current.Doc += repl + "\n"
		} else {
			current.Code += line + "\n"
		}
	}

	return append(sections, current)
}

func markdownComments(sections []*section) {
	for _, section := range sections {
		md := blackfriday.MarkdownBasic([]byte(section.Doc))
		section.Doc = string(md)
	}
	return
}

type File struct {
	Title string
	Sections []*section
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
		out := string(h.Highlight(src))

		sections := extractSections(out)
		markdownComments(sections)
		errt := t.Execute(os.Stdout, File{filename, sections})
		if errt != nil {
			fmt.Fprintf(os.Stderr, errt.Error())
		}
	}
}
