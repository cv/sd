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

func TestArgsFrom(t *testing.T) {
	var tests = []struct {
		name  string
		input string
		check func(t *testing.T, v cobra.PositionalArgs, err error)
	}{
		{
			"happy path with 2 args",
			"#\n# args: 2\n#\n",
			func(t *testing.T, v cobra.PositionalArgs, err error) {
				assert.NoError(t, err)
				assert.Error(t, v(&cobra.Command{}, []string{}))
				assert.Error(t, v(&cobra.Command{}, []string{"first"}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first", "second"}))
				assert.Error(t, v(&cobra.Command{}, []string{"first", "second", "third"}))
			},
		},
		{
			"happy path with no args",
			"#\n# args: 0\n#\n",
			func(t *testing.T, v cobra.PositionalArgs, err error) {
				assert.NoError(t, err)
				assert.NoError(t, v(&cobra.Command{}, []string{}))
				assert.Error(t, v(&cobra.Command{}, []string{"first"}))
				assert.Error(t, v(&cobra.Command{}, []string{"first", "second"}))
			},
		},
		{
			"missing",
			"",
			func(t *testing.T, v cobra.PositionalArgs, err error) {
				assert.NoError(t, err)
				assert.NoError(t, v(&cobra.Command{}, []string{}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first", "second"}))
			},
		},
		{
			"bad value: not numeric",
			"#\n# args: banana\n#\n",
			func(t *testing.T, v cobra.PositionalArgs, err error) {
				assert.NoError(t, err)
				assert.NoError(t, v(&cobra.Command{}, []string{}))
				assert.NoError(t, v(&cobra.Command{}, []string{"first", "second"}))
			},
		},
		{
			"bad value: too large",
			"#\n# args: 999999999999999999999999999999999999999999999999999999999999999999999999\n#\n",
			func(t *testing.T, v cobra.PositionalArgs, err error) {
				assert.Error(t, err)
			},
		},
		{
			"no input",
			"",
			func(t *testing.T, v cobra.PositionalArgs, err error) {
				assert.NoError(t, err)
				assert.NoError(t, v(&cobra.Command{}, []string{}))
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
				f.Close()
				os.Remove(f.Name())
			}()

			v, err := argsFrom(f.Name())
			test.check(t, v, err)
		})
	}
}
