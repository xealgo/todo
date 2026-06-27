package scanner

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewScanner(t *testing.T) {
	cases := []struct {
		keywordLower string
		keywordUpper string
	}{
		{keywordLower: "todo:", keywordUpper: "TODO:"},
		{keywordLower: "note:", keywordUpper: "NOTE:"},
		{keywordLower: "hack:", keywordUpper: "HACK:"},
	}

	keywords := []string{}
	for _, c := range cases {
		keywords = append(keywords, strings.TrimSuffix(c.keywordLower, ":"))
	}

	scanner := NewScanner("./")
	scanner.SetKeywords(keywords...)

	assert.Equal(t, len(scanner.keywords), len(keywords)*2)

	for i := range cases {
		assert.Equal(t, string(scanner.keywords[i*2]), cases[i].keywordLower)
		assert.Equal(t, string(scanner.keywords[i*2+1]), cases[i].keywordUpper)
	}
}
