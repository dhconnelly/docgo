package main

import (
	"flag"
	"fmt"
	"github.com/dhconnelly/blackfriday"
	"io/ioutil"
	"litebrite"
	"os"
	"regexp"
	"strings"
	"text/template"
)

var outdir *string = flag.String("outdir", ".", "directory for generated docs")

var t = template.Must(template.ParseFiles("doc.templ"))

type section struct {
	Doc  string
	Code string
}

var match = regexp.MustCompile(`^\s*//[^\n]\s?`) // Pattern for extracted comments

func extractSections(source string) []*section {
	sections := make([]*section, 0)
	current := new(section)

	// Collect lines up to the next comment group in a section
	for _, line := range strings.Split(source, "\n") {
		if match.FindString(line) != "" {
			if current.Code != "" {
				sections = append(sections, current)
				current = new(section)
			}
			// Strip out the comment delimiters
			current.Doc += match.ReplaceAllString(line, "") + "\n"
		} else {
			current.Code += line + "\n"
		}
	}

	return append(sections, current)
}

func markdownComments(sections []*section) {
	for _, section := range sections {
		section.Doc = string(blackfriday.MarkdownBasic([]byte(section.Doc)))
	}
	return
}

const SEP = "/*[docgoseparator]*/"

var UNSEP = regexp.MustCompile(`<div class="comment">/\*\[docgoseparator\]\*/\s*</div>`)
var BEGINWS = len("<div class=\"comment\">") + len(SEP)
var ENDWS = len("</div>")

func highlightSections(sections []*section) {
	// Rejoin the source code fragments, using SEP as delimiter
	segments := make([]string, 0)
	for _, section := range sections {
		segments = append(segments, section.Code)
	}
	code := strings.Join(segments, SEP)

	// Highlight the joined source
	h := litebrite.Highlighter{"operator", "ident", "literal", "keyword", "comment"}
	hlcode := []byte(h.Highlight(code))

	// Collect the code between subsequent `UNSEP`s.  Some whitespace from
	// the source might be in the `<div>...</div>` that wraps SEP, so we
	// we will add it back when we find it.
	matches := UNSEP.FindAllIndex(hlcode, -1)
	lastend := 0
	lastws := ""
	for i, match := range matches {
		sections[i].Code = lastws + string(hlcode[lastend:match[0]])
		// Extra whitespace comes between `SEP` and the closing `</div>`
		lastws = string(hlcode[match[0]+BEGINWS : match[1]-ENDWS])
		lastend = match[1]
	}
	sections[len(sections)-1].Code = lastws + string(hlcode[lastend:])
}

type File struct {
	Title    string
	Sections []*section
}

func processFile(filename string) {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return
	}
	fi, _ := os.Stat(filename) // Ignore error--already read the file

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
	for _, filename := range flag.Args() {
		processFile(filename)
	}
}
