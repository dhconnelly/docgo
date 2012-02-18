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
	"flag"
)

var outdir *string = flag.String("outdir", ".", "directory for generated docs")

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

const SEP = "/*[docgoseparator]*/"
var UNSEP = regexp.MustCompile(`<div class="comment">/\*\[docgoseparator\]\*/\s*</div>`)
var BEGINWS = len("<div class=\"comment\">") + len(SEP)
var ENDWS = len("</div>")

func highlightSections(sections []*section) {
	// rejoin the source code fragments, using SEP as delimiter
	code := make([]byte, 0)
	for i := 0; i < len(sections) - 1; i++ {
		code = append(code, sections[i].Code...)
		code = append(code, SEP...)
	}
	code = append(code, sections[len(sections)-1].Code...)

	// highlight the joined source
	h := litebrite.Highlighter{"operator", "ident", "literal", "keyword", "comment"}
	hlcode := h.Highlight(code)

	// split the highlighted code around unsep.  some whitespace from
	// the source might be in the `<div>...</div>` that wraps SEP, so we
	// we will add it back when we find it.
	matches := UNSEP.FindAllIndex(hlcode, -1)
	lastend := 0
	lastws := ""
	for i, match := range matches {
		begin, end := match[0], match[1]
		segment := string(hlcode[lastend:begin])
		sections[i].Code = lastws + segment
		beginws := begin + BEGINWS
		endws := end - ENDWS
		lastws = string(hlcode[beginws:endws])
		lastend = end
	}
	sections[len(sections)-1].Code = lastws + string(hlcode[lastend:])
}

type File struct {
	Title string
	Sections []*section
}

func processFile(filename string) {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return
	}
	fi, _ := os.Stat(filename) // ignore error--already read the file

	sections := extractSections(string(src))
	highlightSections(sections)
	markdownComments(sections)

	name := strings.Replace(fi.Name(), ".go", "", -1)
	out, err2 := os.Create(name + ".html") // TODO use outdir
	if err2 != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err2.Error())
		return
	}
	
	errt := t.Execute(out, File{name, sections})
	if errt != nil {
		fmt.Fprintf(os.Stderr, "%s\n", errt.Error())
		return
	}
}

func main() {
	flag.Parse()

	// check if we can write to outdir
	fi, err := os.Stat(*outdir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return
	}
	if !fi.Mode().IsDir() {
		fmt.Fprintf(os.Stderr, "outdir must be a valid directory!\n")
		return
	}
	if (fi.Mode().Perm() & 0200) == 0 {
		fmt.Fprintf(os.Stderr, "can't write to outdir!\n")
		return
	}
	
	// process all input files
	for _, filename := range flag.Args() {
		processFile(filename)
	}
}
