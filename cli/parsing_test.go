package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestShortDescriptionFrom(t *testing.T) {
	var tests = []struct {
		name        string
		inputFormat string
		expected    string
	}{
		{
			"happy path",
			"#\n# %s: blah\n#\n",
			"blah",
		},
		{
			"missing",
			"#\n#\n#\n",
			"",
		},
		{
			"no input",
			"",
			"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f, err := ioutil.TempFile("", test.name)
			assert.NoError(t, err)

			f.WriteString(fmt.Sprintf(test.inputFormat, filepath.Base(f.Name())))
			defer func() {
				f.Close()
				os.Remove(f.Name())
			}()

			v, err := shortDescriptionFrom(f.Name())
			assert.NoError(t, err)
			assert.Equal(t, test.expected, v)
		})
	}
}

func TestUsageFrom(t *testing.T) {
	var tests = []struct {
		name       string
		input      string
		checkUsage func(t *testing.T, name string, actual string)
		checkArgs  func(t *testing.T, v cobra.PositionalArgs)
	}{
		{
			"no arguments",
			"#\n# usage: blah\n#\n",
			func(t *testing.T, name string, actual string) {
				assert.Equal(t, "blah", actual)
			},
			func(t *testing.T, v cobra.PositionalArgs) {
				assert.NoError(t, v(&cobra.Command{}, []string{}))
				assert.Error(t, v(&cobra.Command{}, []string{"first"}))
				assert.Error(t, v(&cobra.Command{}, []string{"first", "second"}))
			},
		},
		{
			"mandatory argument",
			"#\n# usage: blah foo\n#\n",
			func(t *testing.T, name string, actual string) {
				assert.Equal(t, "blah foo", actual)
			},
			func(t *testing.T, v cobra.PositionalArgs) {
				assert.Error(t, v(&cobra.Command{}, []string{}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first"}))
				assert.Error(t, v(&cobra.Command{}, []string{"first", "second"}))
			},
		},
		{
			"optional argument",
			"#\n# usage: blah [foo]\n#\n",
			func(t *testing.T, name string, actual string) {
				assert.Equal(t, "blah [foo]", actual)
			},
			func(t *testing.T, v cobra.PositionalArgs) {
				assert.NoError(t, v(&cobra.Command{}, []string{}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first"}))
				assert.Error(t, v(&cobra.Command{}, []string{"first", "second"}))
			},
		},
		{
			"mandatory and optional arguments",
			"#\n# usage: blah foo [bar]\n#\n",
			func(t *testing.T, name string, actual string) {
				assert.Equal(t, "blah foo [bar]", actual)
			},
			func(t *testing.T, v cobra.PositionalArgs) {
				assert.Error(t, v(&cobra.Command{}, []string{}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first"}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first", "second"}))
				assert.Error(t, v(&cobra.Command{}, []string{"first", "second", "third"}))
			},
		},
		{
			"unlimited arguments, mixed",
			"#\n# usage: blah foo [bar] ...\n#\n",
			func(t *testing.T, name string, actual string) {
				assert.Equal(t, "blah foo [bar] ...", actual)
			},
			func(t *testing.T, v cobra.PositionalArgs) {
				assert.Error(t, v(&cobra.Command{}, []string{}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first"}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first", "second"}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first", "second", "third"}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first", "second", "third", "fourth"}))
			},
		},
		{
			"unlimited arguments, required",
			"#\n# usage: blah foo bar ...\n#\n",
			func(t *testing.T, name string, actual string) {
				assert.Equal(t, "blah foo bar ...", actual)
			},
			func(t *testing.T, v cobra.PositionalArgs) {
				assert.Error(t, v(&cobra.Command{}, []string{}))
				assert.Error(t, v(&cobra.Command{}, []string{"first"}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first", "second"}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first", "second", "third"}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first", "second", "third", "fourth"}))
			},
		},
		{
			"unlimited arguments, optional",
			"#\n# usage: blah [bar] ...\n#\n",
			func(t *testing.T, name string, actual string) {
				assert.Equal(t, "blah [bar] ...", actual)
			},
			func(t *testing.T, v cobra.PositionalArgs) {
				assert.NoError(t, v(&cobra.Command{}, []string{}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first"}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first", "second"}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first", "second", "third"}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first", "second", "third", "fourth"}))
			},
		},
		{
			"missing",
			"#\n#\n#\n",
			func(t *testing.T, name string, actual string) {
				assert.Equal(t, name, actual)
			},
			func(t *testing.T, v cobra.PositionalArgs) {
				assert.NoError(t, v(&cobra.Command{}, []string{}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first"}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first", "second"}))
			},
		},
		{
			"no input",
			"",
			func(t *testing.T, name string, actual string) {
				assert.Equal(t, name, actual)
			},
			func(t *testing.T, v cobra.PositionalArgs) {
				assert.NoError(t, v(&cobra.Command{}, []string{}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first"}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first", "second"}))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f, err := ioutil.TempFile("", test.name)
			assert.NoError(t, err)

			f.WriteString(test.input)
			defer func() {
				_ = f.Close()
				_ = os.Remove(f.Name())
			}()

			usage, args, err := usageFrom(f.Name())
			assert.NoError(t, err)
			test.checkUsage(t, filepath.Base(f.Name()), usage)
			test.checkArgs(t, args)
		})
	}
}

func TestExampleFrom(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected string
	}{
		{
			"happy path",
			"#\n# example: blah\n#\n",
			"  sd blah",
		},
		{
			"missing",
			"#\n#\n#\n",
			"",
		},
		{
			"no input",
			"",
			"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f, err := ioutil.TempFile("", test.name)
			assert.NoError(t, err)

			f.WriteString(test.input)
			defer func() {
				_ = f.Close()
				_ = os.Remove(f.Name())
			}()

			v, err := exampleFrom(f.Name())
			assert.NoError(t, err)
			assert.Equal(t, test.expected, v)
		})
	}
}
