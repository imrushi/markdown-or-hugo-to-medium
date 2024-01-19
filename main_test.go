package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractPostName(t *testing.T) {
	testCases := []struct {
		desc      string
		commitMsg string
		out       []string
	}{
		{
			desc:      "extract post name",
			commitMsg: "PUBLISH: test.md",
			out:       []string{"Test"},
		},
		{
			desc:      "extract post name by separating from hyphen",
			commitMsg: "PUBLISH: this-is-my-first-blog.md",
			out:       []string{"This Is My First Blog"},
		},
		{
			desc:      "extract post name by separating from underscore",
			commitMsg: "PUBLISH: this_is_my_first_blog.md",
			out:       []string{"This_is_my_first_blog"},
		},
		{
			desc:      "extract post name with symbols",
			commitMsg: "PUBLISH: this^is-my-#first-$blog!!!.md",
			out:       []string{"This^Is My #First $Blog!!!"},
		},
		{
			desc:      "empty commit message",
			commitMsg: "",
			out:       nil,
		},
		{
			desc:      "commit message with no 'PUBLISH:' prefix",
			commitMsg: "Some other message",
			out:       nil,
		},
		{
			desc:      "commit message with no extension",
			commitMsg: "PUBLISH: post",
			out:       []string{"Post"},
		},
		{
			desc:      "commit message with Single 'PUBLISH:' occurrence and comma separated",
			commitMsg: "PUBLISH: post1.md, post2.md",
			out:       []string{"Post1", "Post2"},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			_, postName := extractPostName(tC.commitMsg)
			assert.Equal(t, tC.out, postName)
		})
	}
}

func TestParseHeader(t *testing.T) {
	testCases := []struct {
		desc       string
		in         string
		outContent string
		outTitle   string
		outTags    []string
	}{
		{
			desc: "Parse YAML header (content, title, tags) and remove frontmatter",
			in: `---
title: "Go: Basics and a Dash of Clean Code"
date: "2023-09-03T19:31:57+05:30"
author: "Gopher"
authorTwitter: "golang"
coverCredit: "Go Gopher created by Renee French and Ashley McNamara"
categories: ["Go"]
tags: ["go", "basics", "clean code"]
keywords:
  [
    "go",
    "golang",
    "basics of go",
    "variables",
  ]
description: "Welcome to guide on Go programming! Whether you're beginner looking to grasp the basics of the language or an experience developer seeking to enhance your clean code skills, our blog has you covered. Explore the fundamentals of Go syntax and discover how to elevate your coding style with a touch of clean code principles."
showFullContent: false
readingTime: true
hideComments: false
draft: false
toc: true
---

---
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed id odio nisl. Vestibulum sed urna ut nulla molestie finibus eu vitae nisl. Donec eget venenatis turpis. Nulla nec molestie dolor, quis consequat lectus. Pellentesque faucibus leo vel erat fringilla tincidunt. Sed auctor magna sed lacus vestibulum, ac lobortis orci volutpat. Fusce tincidunt est in felis sollicitudin mollis. Fusce eget iaculis dolor. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Sed vehicula ligula nec sem facilisis lobortis. Phasellus quis ornare nunc. Suspendisse potenti. Nam quis nisl vel urna scelerisque auctor. Pellentesque laoreet molestie egestas. Ut in.
`,
			outContent: `

---
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed id odio nisl. Vestibulum sed urna ut nulla molestie finibus eu vitae nisl. Donec eget venenatis turpis. Nulla nec molestie dolor, quis consequat lectus. Pellentesque faucibus leo vel erat fringilla tincidunt. Sed auctor magna sed lacus vestibulum, ac lobortis orci volutpat. Fusce tincidunt est in felis sollicitudin mollis. Fusce eget iaculis dolor. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Sed vehicula ligula nec sem facilisis lobortis. Phasellus quis ornare nunc. Suspendisse potenti. Nam quis nisl vel urna scelerisque auctor. Pellentesque laoreet molestie egestas. Ut in.
`,
			outTitle: "Go: Basics and a Dash of Clean Code",
			outTags:  []string{"go", "basics", "clean code"},
		},
		{
			desc: "Parse YAML header (content, title, tags) and remove frontmatter",
			in: `+++
categories = ['Development', 'VIM']
date = '2012-04-06'
description = 'spf13-vim is a cross platform distribution of vim plugins and resources for Vim.'
slug = 'spf13-vim-3-0-release-and-new-website'
tags = ['.vimrc', 'plugins', 'spf13-vim', 'vim']
title = 'spf13-vim 3.0 release and new website'
+++

# tss

Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.`,
			outContent: `

# tss

Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.`,
			outTitle: "spf13-vim 3.0 release and new website",
			outTags:  []string{".vimrc", "plugins", "spf13-vim", "vim"},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			content, title, tags := parseHeader(tC.in)

			assert.Equal(t, tC.outContent, content)
			assert.Equal(t, tC.outTitle, title)
			assert.Equal(t, tC.outTags, tags)
		})
	}
}

func TestAddTrailingSlash(t *testing.T) {
	testCases := []struct {
		desc   string
		input  string
		output string
	}{
		{desc: "URL without trailing slash", input: "https://example.com/path", output: "https://example.com/path/"},
		{desc: "URL with trailing slash", input: "https://example.com/path/", output: "https://example.com/path/"},
		{desc: "Empty string", input: "", output: "/"},
		{desc: "Root path", input: "/", output: "/"},
		{desc: "Single Character", input: "x", output: "x/"},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			result := addURLTrailingSlash(tC.input)
			assert.Equal(t, tC.output, result)
		})
	}
}
