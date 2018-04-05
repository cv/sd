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

func main() {
	root := &cobra.Command{
		Use: "sd",
	}

	completions := &cobra.Command{
		Use:   "completions",
		Short: "Generate completion scripts",
	}
	completions.Run = func(cmd *cobra.Command, args []string) {
		completions.Usage()
	}

	completions.AddCommand(&cobra.Command{
		Use:   "bash",
		Short: "Generate completions for bash",
		RunE: func(cmd *cobra.Command, args []string) error {
			return root.GenBashCompletion(os.Stdout)
		},
	})

	completions.AddCommand(&cobra.Command{
		Use:   "zsh",
		Short: "Generate completions for zsh",
		RunE: func(cmd *cobra.Command, args []string) error {
			return root.GenZshCompletion(os.Stdout)
		},
	})

	err := loadCommandsInto(root)
	if err != nil {
		panic(err)
	}

	root.AddCommand(completions)
	root.ExecuteC()
}

func loadCommandsInto(root *cobra.Command) error {
	baseDir := filepath.Join(os.Getenv("HOME"), ".sd")
	parent := root

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		switch {

		case err != nil, path == baseDir:
			return nil

		case strings.HasPrefix(filepath.Base(path), "."):
			return filepath.SkipDir

		case filepath.Base(path) == "README":
			cmd, err := commandFromReadme(path)
			if err != nil {
				return err
			}
			parent.AddCommand(cmd)
			parent = cmd

		case info.Mode()&0100 != 0 && !info.IsDir():
			cmd, err := commandFromScript(path)
			if err != nil {
				return err
			}
			parent.AddCommand(cmd)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
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

	cmd := cobra.Command{
		Use:     filepath.Base(path),
		Short:   shortDesc,
		Example: example,
		Run: func(cmd *cobra.Command, args []string) {
			sh(path, args)
		},
	}

	return &cmd, nil
}

/*

First line of README is the short description, and everything else is the long description.

 */
func commandFromReadme(path string) (*cobra.Command, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cmd := &cobra.Command{
		Use:   filepath.Base(filepath.Dir(path)),
		Short: strings.Split(string(file), "\n")[0],
		Long:  string(file),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	return cmd, nil
}

func sh(cmd string, args []string) error {
	return syscall.Exec(cmd, append([]string{cmd}, args...), os.Environ())
}
