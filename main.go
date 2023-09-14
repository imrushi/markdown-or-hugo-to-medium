package main

import (
	"encoding/json"
	"flag"
	"fmt"
)

type ShortCodes struct {
	Name    string `json:"name"`
	Regex   string `json:"regex,omitempty"`
	Replace string `json:"replace,omitempty"`
}

// Set implements flag.Value.
func (this *ShortCodes) Set(s string) error {
	return json.Unmarshal([]byte(s), this)
}

// String implements flag.Value.
func (this *ShortCodes) String() string {
	b, _ := json.Marshal(*this)
	return string(b)
}

func main() {
	var shortcodes ShortCodes
	var markdownOrHugo string
	var replaceHyperlinkToLink bool

	flag.StringVar(&markdownOrHugo, "markdownOrHugo", "markdown", "Set the flag for parsing hugo markdown or simple markdown.")
	flag.Var(&shortcodes, "shortCodes", "Pass JSON object as string for parsing shortcode to markdown")
	flag.BoolVar(&replaceHyperlinkToLink, "replaceHyperlinkToLink", false, "replace markdown hyperlink syntax with just link")
	flag.Parse()

	fmt.Println(shortcodes, markdownOrHugo, replaceHyperlinkToLink)
}
