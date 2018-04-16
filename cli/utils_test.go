package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeduplicate(t *testing.T) {
	var tests = []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			"simple",
			strings.Split("aaaaaa", ""),
			[]string{"a"},
		},
		{
			"two elements",
			[]string{"a", "b"},
			deduplicate(strings.Split("aaaabb", "")),
		},
		{
			"two elements repeated",
			deduplicate(strings.Split("ababab", "")),
			[]string{"a", "b"},
		},
		{
			"maintains ordering",
			deduplicate(strings.Split("acbaabab", "")),
			[]string{"a", "c", "b"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, deduplicate(test.input))
		})
	}
}
