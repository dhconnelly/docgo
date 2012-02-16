package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"blackfriday"
)

type section struct {
	docsText string
	docsHTML string
	codeText string
	codeHTML string
}

var commentLine = regexp.MustCompile("^\\s*//[^\n]")

func parse(content string) []*section {
	sections := make([]*section, 0)
	lines := strings.Split(content, "\n")
	docs, code := "", ""

	saveSection := func() {
		if code == "" { return }
		sec := section{docsText: docs, codeText: code}
		sections = append(sections, &sec)
		docs, code = "", ""
	}

	for _, line := range lines {
		if commentLine.FindString(line) != "" {
			saveSection()
			docs += commentLine.ReplaceAllString(line, "") + "\n"
		} else {
			code += line + "\n"
		}
	}
	
	saveSection()
	return sections
}

func highlight(sections []*section) []*section {
	for _, section := range sections {
		section.codeHTML = section.codeText
	}
	return sections
}

func markdown(sections []*section) []*section {
	for _, section := range sections {
		md := blackfriday.MarkdownBasic([]byte(section.docsText))
		section.docsHTML = string(md)
	}
	return sections
}

func html(sections []*section) string {
	out := ""
	for _, section := range sections {
		out += section.docsHTML
	}
	return out
}

func GenerateDocs(content string) string {
	return html(highlight(markdown(parse(content))))
}

func main() {
	files := os.Args[1:]
	for _, filename := range files {
		content, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			continue
		}
		fmt.Print(GenerateDocs(string(content)))
	}
}
