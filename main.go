package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

// These variables are set by the build process (see Makefile)
var (
	// Version of the application (SemVer)
	Version string

	// Commit hash (short SHA1 of HEAD)
	Commit string
)

func main() {
	root := &cobra.Command{
		Use:     "sd",
		Version: fmt.Sprintf("%s (%s)", Version, Commit),
	}

	err := loadCommandsInto(root)
	if err != nil {
		panic(err)
	}

	root.AddCommand(completions(root))
	root.Execute()
}

func loadCommandsInto(root *cobra.Command) error {
	home := filepath.Join(os.Getenv("HOME"), ".sd")

	wd, _ := os.Getwd()
	current := filepath.Join(wd, "scripts")

	for _, path := range []string{home, current} {
		cmds, err := visitDir(path)
		if err != nil {
			return err
		}

		for _, c := range cmds {
			root.AddCommand(c)
		}
	}

	return nil
}

func visitDir(path string) ([]*cobra.Command, error) {
	var cmds []*cobra.Command

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cmds, nil
	}

	items, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		switch {
		case strings.HasPrefix(item.Name(), "."):
			continue

		case item.IsDir():
			cmd := &cobra.Command{
				Use: item.Name(),
			}

			readme, err := ioutil.ReadFile(filepath.Join(path, item.Name(), "README"))
			if err == nil {
				cmd.Short = strings.Split(string(readme), "\n")[0]
				cmd.Long = string(readme)
			}

			subcmds, err := visitDir(filepath.Join(path, item.Name()))
			if err != nil {
				return nil, err
			}
			for _, i := range subcmds {
				cmd.AddCommand(i)
			}

			if cmd.HasSubCommands() {
				cmd.Run = func(cmd *cobra.Command, args []string) {
					cmd.Usage()
				}
			}
			cmds = append(cmds, cmd)

		case item.Mode()&0100 != 0:
			cmd, err := commandFromScript(filepath.Join(path, item.Name()))
			if err != nil {
				return nil, err
			}
			cmds = append(cmds, cmd)
		}
	}
	return cmds, nil
}

/*

Looks for a line like this:

# name-of-the-file: short description.

*/
func shortDescriptionFrom(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	r := regexp.MustCompile(fmt.Sprintf(`^# %s: (.*)$`, regexp.QuoteMeta(filepath.Base(path))))
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		match := r.FindStringSubmatch(scanner.Text())
		if len(match) == 2 {
			return match[1], nil
		}
	}
	return "", nil
}

/*

Looks for a line like this:

# example: foo bar 1 2 3

*/
func exampleFrom(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	r := regexp.MustCompile(`^# example: (.*)$`)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		match := r.FindStringSubmatch(scanner.Text())
		if len(match) == 2 {
			return "  sd " + match[1], nil
		}
	}
	return "", nil
}

func commandFromScript(path string) (*cobra.Command, error) {
	shortDesc, err := shortDescriptionFrom(path)
	if err != nil {
		return nil, err
	}

	example, err := exampleFrom(path)
	if err != nil {
		return nil, err
	}

	cmd := &cobra.Command{
		Use:     filepath.Base(path),
		Short:   shortDesc,
		Example: example,
		Run: func(cmd *cobra.Command, args []string) {
			sh(path, args)
		},
	}

	return cmd, nil
}

func sh(cmd string, args []string) error {
	return syscall.Exec(cmd, append([]string{cmd}, args...), os.Environ())
}

func completions(root *cobra.Command) *cobra.Command {
	c := &cobra.Command{
		Use:   "completions",
		Short: "Generate completion scripts",
	}

	c.Run = func(cmd *cobra.Command, args []string) {
		c.Usage()
	}

	c.AddCommand(&cobra.Command{
		Use:   "bash",
		Short: "Generate completions for bash",
		RunE: func(cmd *cobra.Command, args []string) error {
			return root.GenBashCompletion(os.Stdout)
		},
	})

	c.AddCommand(&cobra.Command{
		Use:   "zsh",
		Short: "Generate completions for zsh",
		RunE: func(cmd *cobra.Command, args []string) error {
			return root.GenZshCompletion(os.Stdout)
		},
	})

	return c
}
