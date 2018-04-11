package cli

import (
	"os"
	"strings"
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
	var restore []string
	copy(restore, os.Args)
	defer func() {
		copy(os.Args, restore)
	}()

	sd := &sd{
		root: &cobra.Command{},
	}
	sd.initCompletions()

	t.Run("sets logrus level", func(t *testing.T) {
		assert.Equal(t, logrus.InfoLevel, logrus.GetLevel())

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

}
