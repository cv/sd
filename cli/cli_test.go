package cli

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestInitAliasing(t *testing.T) {
	sd := &sd{
		root: &cobra.Command{},
	}
	sd.initAliasing()

	t.Run("flag is hidden", func(t *testing.T) {
		assert.True(t, sd.root.PersistentFlags().Lookup("alias").Hidden)
	})

	t.Run("adds a default alias flag", func(t *testing.T) {
		sd.root.ParseFlags([]string{""})

		v, err := sd.root.PersistentFlags().GetString("alias")
		assert.NoError(t, err)
		assert.Equal(t, "sd", v)
	})

	t.Run("sets the name of the root command when aliased", func(t *testing.T) {
		sd.root.ParseFlags([]string{"-a", "quack"})
		sd.root.PersistentPreRunE(sd.root, []string{})

		_, err := sd.root.PersistentFlags().GetString("alias")
		assert.NoError(t, err)
		assert.Equal(t, "quack", sd.root.Use)
	})
}

func TestInitCompletions(t *testing.T) {
	sd := &sd{
		root: &cobra.Command{},
	}
	sd.initCompletions()

	t.Run("adds completion command", func(t *testing.T) {
		assert.Len(t, sd.root.Commands(), 1)
		cmd := sd.root.Commands()[0]
		assert.Equal(t, "completions", cmd.Use)

		t.Run("subcommands", func(t *testing.T) {
			assert.Len(t, cmd.Commands(), 2)
		})

		t.Run("bash", func(t *testing.T) {
			assert.Equal(t, "bash", cmd.Commands()[0].Use)
		})

		t.Run("zsh", func(t *testing.T) {
			assert.Equal(t, "zsh", cmd.Commands()[1].Use)
		})
	})
}

func TestInitDebugging(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	var restore []string
	copy(restore, os.Args)
	defer func() {
		copy(os.Args, restore)
	}()

	sd := &sd{
		root: &cobra.Command{},
	}

	t.Run("sets logrus level", func(t *testing.T) {
		assert.Equal(t, logrus.InfoLevel, logrus.GetLevel())
		defer logrus.SetLevel(logrus.InfoLevel)

		os.Args = []string{"-d"}
		sd.initDebugging()

		assert.Equal(t, logrus.DebugLevel, logrus.GetLevel())
	})
}

func TestInitEditing(t *testing.T) {
	sd := &sd{
		root: &cobra.Command{},
	}
	sd.initEditing()

	t.Run("creates flag", func(t *testing.T) {
		assert.NotNil(t, sd.root.PersistentFlags().Lookup("edit"))
	})
}

func TestDeduplicate(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		assert.Equal(t, []string{"a"}, deduplicate(strings.Split("aaaaaa", "")))
	})

	t.Run("two elements", func(t *testing.T) {
		assert.Equal(t, []string{"a", "b"}, deduplicate(strings.Split("aaaabb", "")))
	})

	t.Run("two elements repeated", func(t *testing.T) {
		assert.Equal(t, []string{"a", "b"}, deduplicate(strings.Split("ababab", "")))
	})

	t.Run("maintains ordering", func(t *testing.T) {
		assert.Equal(t, []string{"a", "c", "b"}, deduplicate(strings.Split("acbabab", "")))
	})
}

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

func TestCommandFromScript(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		f, err := ioutil.TempFile("", "test-command-from-script")
		assert.NoError(t, err)

		f.WriteString(fmt.Sprintf("#\n# %s: blah\n# example: one two three\n#\n", filepath.Base(f.Name())))
		defer func() {
			f.Close()
			os.Remove(f.Name())
		}()

		c, err := commandFromScript(f.Name())
		assert.NoError(t, err)
		assert.Equal(t, filepath.Base(f.Name()), c.Use)
		assert.Equal(t, "blah", c.Short)
		assert.Equal(t, "  sd one two three", c.Example)
		assert.Equal(t, f.Name(), c.Annotations["Source"])
	})
}

func TestExecCommand(t *testing.T) {
	t.Run("edit with VISUAL", func(t *testing.T) {
		sd := &sd{root: &cobra.Command{}}
		sd.initEditing()
		sd.root.PersistentFlags().Set("edit", "true")

		defer func() {
			syscallExec = syscall.Exec
			env = os.Getenv
		}()

		env = func(key string) string {
			if key == "VISUAL" {
				return "some-visual-editor"
			}
			return ""
		}

		called := false
		syscallExec = func(argv0 string, argv []string, envv []string) error {
			called = true
			assert.Equal(t, "/bin/sh", argv0)
			assert.Equal(t, []string{"sh", "-c", "some-visual-editor /path/to/foo"}, argv)
			return nil
		}

		cmd := &cobra.Command{
			Use: "foo",
			Annotations: map[string]string{
				"Source": "/path/to/foo",
			},
		}
		sd.root.AddCommand(cmd)

		err := execCommand(cmd, []string{})
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("edit with EDITOR", func(t *testing.T) {
		sd := &sd{root: &cobra.Command{}}
		sd.initEditing()
		sd.root.PersistentFlags().Set("edit", "true")

		defer func() {
			syscallExec = syscall.Exec
			env = os.Getenv
		}()

		env = func(key string) string {
			if key == "VISUAL" {
				return ""
			}
			if key == "EDITOR" {
				return "some-editor"
			}
			return ""
		}

		called := false
		syscallExec = func(argv0 string, argv []string, envv []string) error {
			called = true
			assert.Equal(t, "/bin/sh", argv0)
			assert.Equal(t, []string{"sh", "-c", "some-editor /path/to/foo"}, argv)
			return nil
		}

		cmd := &cobra.Command{
			Use: "foo",
			Annotations: map[string]string{
				"Source": "/path/to/foo",
			},
		}
		sd.root.AddCommand(cmd)

		err := execCommand(cmd, []string{})
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("edit with default vim", func(t *testing.T) {
		sd := &sd{root: &cobra.Command{}}
		sd.initEditing()
		sd.root.PersistentFlags().Set("edit", "true")

		defer func() {
			syscallExec = syscall.Exec
			env = os.Getenv
		}()

		env = func(key string) string {
			return ""
		}

		called := false
		syscallExec = func(argv0 string, argv []string, envv []string) error {
			called = true
			assert.Equal(t, "/bin/sh", argv0)
			assert.Equal(t, []string{"sh", "-c", "$(which vim) /path/to/foo"}, argv)
			return nil
		}

		cmd := &cobra.Command{
			Use: "foo",
			Annotations: map[string]string{
				"Source": "/path/to/foo",
			},
		}
		sd.root.AddCommand(cmd)

		err := execCommand(cmd, []string{})
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("exec script", func(t *testing.T) {
		sd := &sd{root: &cobra.Command{}}
		sd.initEditing()

		defer func() {
			syscallExec = syscall.Exec
			env = os.Getenv
		}()

		env = func(key string) string {
			return ""
		}

		called := false
		syscallExec = func(argv0 string, argv []string, envv []string) error {
			called = true
			assert.Equal(t, "/path/to/foo", argv0)
			assert.Equal(t, []string{"/path/to/foo", "bar"}, argv)
			return nil
		}

		cmd := &cobra.Command{
			Use: "foo",
			Annotations: map[string]string{
				"Source": "/path/to/foo",
			},
		}
		sd.root.AddCommand(cmd)

		err := execCommand(cmd, []string{"bar"})
		assert.NoError(t, err)
		assert.True(t, called)
	})
}

func TestInit(t *testing.T) {
	t.Run("sets the initialized flag", func(t *testing.T) {
		sd := &sd{root: &cobra.Command{}}
		sd.init()
		assert.True(t, sd.initialized)
	})
}

func TestNew(t *testing.T) {
	t.Run("creates valid instane", func(t *testing.T) {
		sd := New("1.0.0").(*sd)
		assert.True(t, sd.initialized)
		assert.Equal(t, "1.0.0", sd.root.Version)
	})
}

func TestShowUsage(t *testing.T) {
	cmd := &cobra.Command{}
	called := false
	cmd.SetUsageFunc(func(c *cobra.Command) error {
		called = true
		assert.Equal(t, cmd, c)
		return nil
	})
	showUsage(cmd, []string{})
	assert.True(t, called)
}

func TestRunCompletionsBash(t *testing.T) {
	restore := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	outC := make(chan string)

	var out string
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	var args []string
	copy(args, os.Args)
	defer func() {
		copy(os.Args, args)
	}()

	os.Args = []string{"sd", "completions", "bash"}

	err := New("1.0").Run()

	func() {
		w.Close()
		os.Stdout = restore
		out = <-outC
	}()

	assert.NoError(t, err)

	assert.Contains(t, out, "__sd_debug()")
}

func TestRunCompletionsZsh(t *testing.T) {
	restore := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	outC := make(chan string)

	var out string
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	var args []string
	copy(args, os.Args)
	defer func() {
		copy(os.Args, args)
	}()

	os.Args = []string{"sd", "completions", "zsh"}

	err := New("1.0").Run()

	func() {
		w.Close()
		os.Stdout = restore
		out = <-outC
	}()

	assert.NoError(t, err)
	assert.Contains(t, out, "#compdef sd")
}
