package main

import (
	"bytes"
	"fmt"
	"github.com/dhconnelly/blackfriday"
	"io/ioutil"
	"litebrite"
	"os"
	"regexp"
	"strings"
	"text/template"
)

var templ = template.Must(template.ParseFiles("doc.templ"))

func generateDocs(title, src string) string {
	var b bytes.Buffer
	sections := extractSections(src)
	highlightSections(sections)
	markdownComments(sections)
	templ.Execute(&b, docs{title, sections})
	return b.String()
}

type docs struct {
	Title    string
	Sections []*section
}

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

const SEP = "/*[docgoseparator]*/" // replacement for comment groups

var UNSEP = regexp.MustCompile(`<div class="comment">/\*\[docgoseparator\]\*/</div>`)

func highlightSections(sections []*section) {
	// Rejoin the source code fragments, using SEP as delimiter
	segments := make([]string, 0)
	for _, section := range sections {
		segments = append(segments, section.Code)
	}
	code := strings.Join(segments, SEP)

	// Highlight the joined source
	h := litebrite.Highlighter{"operator", "ident", "literal", "keyword", "comment"}
	hlcode := h.Highlight(code)

	// Collect the code between subsequent `UNSEP`s
	matches := append(UNSEP.FindAllStringIndex(hlcode, -1), []int{len(hlcode), 0})
	lastend := 0
	for i, match := range matches {
		sections[i].Code = hlcode[lastend:match[0]]
		lastend = match[1]
	}
}

// ## Running

func main() {
	if fi, err := os.Stat("docs"); err != nil || !fi.IsDir() {
		fmt.Fprintln("Failed to open directory docs")
		return
	}
	for _, filename := range os.Args[1:] {
		infile, err := os.Open(filename)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}

		fi, _ := infile.Stat()
		outname := "docs/" + strings.Replace(fi.Name(), ".go", ".html", 1)
		outfile, err2 := os.Create(outname)
		if err2 != nil {
			fmt.Fprintln(os.Stderr, err2.Error())
			return
		}

		src, _ := ioutil.ReadAll(infile)
		outfile.WriteString(generateDocs(fi.Name(), string(src)))
	}
}
