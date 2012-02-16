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
	docs string
	docsHTML string
	code string
	codeHTML string
}

var match = regexp.MustCompile("^\\s*//[^\n]")

func parse(content string) []*section {
	lines := strings.Split(content, "\n")
	sections := make([]*section, 0)
	current := new(section)

	for _, line := range lines {
		if match.FindString(line) != "" {
			if current.code != "" {
				sections = append(sections, current)
				current = new(section)
			}
			current.docs += match.ReplaceAllString(line, "") + "\n"
		} else {
			current.code += line + "\n"
		}
	}
	
	return append(sections, current)
}

func highlight(sections []*section) []*section {
	for _, section := range sections {
		section.codeHTML = section.code
	}
	return sections
}

func markdown(sections []*section) []*section {
	for _, section := range sections {
		md := blackfriday.MarkdownBasic([]byte(section.docs))
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
