package helper

import (
	"net/url"
	"strings"
	"unicode"
)

type PathSpec struct {
	RootURL string
}

func NewPathSpec(rootURL string) *PathSpec {
	return &PathSpec{
		RootURL: rootURL,
	}
}

// URLize is similar to MakePath, but with Unicode handling
// Example:
//
//	uri: Vim (text editor)
//	urlize: vim-text-editor
func (p *PathSpec) URLize(uri string) string {
	return p.URLEscape(p.MakePathSanitized(uri))
}

// URLEscape escapes unicode letters.
func (p *PathSpec) URLEscape(uri string) string {
	// escape unicode letters
	parsedURI, err := url.Parse(uri)
	if err != nil {
		// if net/url can not parse URL it means Sanitize works incorrectly
		panic(err)
	}
	x := parsedURI.String()
	return x
}

// MakePath takes a string with any characters and replace it
// so the string could be used in a path.
// It does so by creating a Unicode-sanitized string, with the spaces replaced,
// whilst preserving the original casing of the string.
// E.g. Social Media -> Social-Media
func (p *PathSpec) MakePath(s string) string {
	return p.UnicodeSanitize(s)
}

// MakePathSanitized creates a Unicode-sanitized string, with the spaces replaced
func (p *PathSpec) MakePathSanitized(s string) string {
	return strings.ToLower(p.MakePath(s))
}

// From https://golang.org/src/net/url/url.go
func ishex(c rune) bool {
	switch {
	case '0' <= c && c <= '9':
		return true
	case 'a' <= c && c <= 'f':
		return true
	case 'A' <= c && c <= 'F':
		return true
	}
	return false
}

// UnicodeSanitize sanitizes string to be used in Hugo URL's, allowing only
// a predefined set of special Unicode characters.
// If RemovePathAccents configuration flag is enabled, Unicode accents
// are also removed.
// Hyphens in the original input are maintained.
// Spaces will be replaced with a single hyphen, and sequential replacement hyphens will be reduced to one
func (p *PathSpec) UnicodeSanitize(s string) string {
	source := []rune(s)
	target := make([]rune, 0, len(source))
	var (
		prependHyphen bool
		wasHyphen     bool
	)

	for i, r := range source {
		isAllowed := r == '.' || r == '/' || r == '\\' || r == '_' || r == '#' || r == '+' || r == '~' || r == '-' || r == '@'
		isAllowed = isAllowed || unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r)
		isAllowed = isAllowed || (r == '%' && i+2 < len(source) && ishex(source[i+1]) && ishex(source[i+2]))

		if isAllowed {
			// track explicit hyphen in input; no need to add a new hyphen if
			// we just saw one.
			wasHyphen = r == '-'

			if prependHyphen {
				// if currently have a hyphen, don't prepend an extra one
				if !wasHyphen {
					target = append(target, '-')
				}
				prependHyphen = false
			}
			target = append(target, r)
		} else if len(target) > 0 && !wasHyphen && unicode.IsSpace(r) {
			prependHyphen = true
		}
	}
	return string(target)
}
