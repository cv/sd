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
	t.Run("happy path", func(t *testing.T) {
		f, err := ioutil.TempFile("", "test-short-description-ok")
		assert.NoError(t, err)

		f.WriteString(fmt.Sprintf("#\n# %s: blah\n#\n", filepath.Base(f.Name())))
		defer func() {
			f.Close()
			os.Remove(f.Name())
		}()

		v, err := shortDescriptionFrom(f.Name())
		assert.NoError(t, err)
		assert.Equal(t, "blah", v)
	})
	t.Run("missing", func(t *testing.T) {
		f, err := ioutil.TempFile("", "test-short-description-missing")
		assert.NoError(t, err)

		f.WriteString("#\n#\n#\n")
		defer func() {
			f.Close()
			os.Remove(f.Name())
		}()

		v, err := shortDescriptionFrom(f.Name())
		assert.NoError(t, err)
		assert.Equal(t, "", v)
	})
}

func TestExampleFrom(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		f, err := ioutil.TempFile("", "test-example-ok")
		assert.NoError(t, err)

		f.WriteString("#\n# example: blah\n#\n")
		defer func() {
			_ = f.Close()
			_ = os.Remove(f.Name())
		}()

		v, err := exampleFrom(f.Name())
		assert.NoError(t, err)
		assert.Equal(t, "  sd blah", v)
	})
	t.Run("missing", func(t *testing.T) {
		f, err := ioutil.TempFile("", "test-example-missing")
		assert.NoError(t, err)

		f.WriteString("#\n#\n#\n")
		defer func() {
			_ = f.Close()
			_ = os.Remove(f.Name())
		}()

		v, err := exampleFrom(f.Name())
		assert.NoError(t, err)
		assert.Equal(t, "", v)
	})
}

func TestArgsFrom(t *testing.T) {
	t.Run("happy path with 3 args", func(t *testing.T) {
		f, err := ioutil.TempFile("", "test-args3-ok")
		assert.NoError(t, err)

		f.WriteString("#\n# args: 2\n#\n")
		defer func() {
			f.Close()
			os.Remove(f.Name())
		}()

		v, err := argsFrom(f.Name())
		assert.NoError(t, err)
		assert.Error(t, v(&cobra.Command{}, []string{}))
		assert.Error(t, v(&cobra.Command{}, []string{"first"}))
		assert.NoError(t, v(&cobra.Command{}, []string{"first", "second"}))
		assert.Error(t, v(&cobra.Command{}, []string{"first", "second", "third"}))
	})

	t.Run("happy path with no args", func(t *testing.T) {
		f, err := ioutil.TempFile("", "test-args0-ok")
		assert.NoError(t, err)

		f.WriteString("#\n# args: 0\n#\n")
		defer func() {
			f.Close()
			os.Remove(f.Name())
		}()

		v, err := argsFrom(f.Name())
		assert.NoError(t, err)
		assert.NoError(t, v(&cobra.Command{}, []string{}))
		assert.Error(t, v(&cobra.Command{}, []string{"first"}))
		assert.Error(t, v(&cobra.Command{}, []string{"first", "second"}))
	})

	t.Run("missing", func(t *testing.T) {
		f, err := ioutil.TempFile("", "test-args-missing")
		assert.NoError(t, err)

		f.WriteString("#\n#\n#\n")
		defer func() {
			f.Close()
			os.Remove(f.Name())
		}()

		v, err := argsFrom(f.Name())
		assert.NoError(t, err)
		assert.NoError(t, v(&cobra.Command{}, []string{}))
		assert.NoError(t, v(&cobra.Command{}, []string{"first", "second"}))
	})

	t.Run("bad value: not numeric", func(t *testing.T) {
		f, err := ioutil.TempFile("", "test-args-not-numeric")
		assert.NoError(t, err)

		f.WriteString("#\n# args: banana\n#\n")
		defer func() {
			f.Close()
			os.Remove(f.Name())
		}()

		v, err := argsFrom(f.Name())
		assert.NoError(t, err)
		assert.NoError(t, v(&cobra.Command{}, []string{}))
		assert.NoError(t, v(&cobra.Command{}, []string{"first", "second"}))
	})

	t.Run("bad value: too large", func(t *testing.T) {
		f, err := ioutil.TempFile("", "test-args-too-large")
		assert.NoError(t, err)

		f.WriteString("#\n# args: 999999999999999999999999999999999999999999999999999999999999999999999999\n#\n")
		defer func() {
			f.Close()
			os.Remove(f.Name())
		}()

		v, err := argsFrom(f.Name())
		if !assert.Error(t, err) {
			assert.NoError(t, v(&cobra.Command{}, []string{}))
			assert.NoError(t, v(&cobra.Command{}, []string{"first", "second"}))
		}
	})
}
