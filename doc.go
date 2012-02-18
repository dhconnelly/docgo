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

var match = regexp.MustCompile(`^\s*//[^\n]\s?`)
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

const SEP = "// [docgoseparator]\n"
var unsep = regexp.MustCompile(`<div class="comment">// \[docgoseparator\]\n</div>`)

func highlightSections(sections []*section) {
	// rejoin the source code fragments, using SEP as delimiter
	code := make([]byte, 0)
	for i := 0; i < len(sections) - 1; i++ {
		code = append(code, sections[i].Code...)
		code = append(code, SEP...)
	}
	code = append(code, sections[len(sections)-1].Code...)

	// highlight the joined source
	h := new(litebrite.Highlighter)
	h.IdentClass = "ident"
	h.LiteralClass = "literal"
	h.KeywordClass = "keyword"
	h.OperatorClass = "operator"
	h.CommentClass = "comment"
	hlcode := string(h.Highlight(code))

	// regexp package doesn't support splitting, so first replace all
	// UNSEP matches with SEP, which we can strings.Split around.
	segments := strings.Split(unsep.ReplaceAllString(hlcode, SEP), SEP)
	if len(segments) != len(sections) {
		panic("Failed to recover all source fragments!")
	}
	
	for i, segment := range segments {
		sections[i].Code = segment
	}
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

		sections := extractSections(string(src))
		highlightSections(sections)
		markdownComments(sections)
		
		errt := t.Execute(os.Stdout, File{filename, sections})
		if errt != nil {
			fmt.Fprintf(os.Stderr, errt.Error())
		}
	}
}
