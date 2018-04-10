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

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	version string
)

func main() {
	root := &cobra.Command{
		Use:     "sd",
		Version: version,
	}
	root.AddCommand(completions(root))
	root.PersistentFlags().BoolP("debug", "d", false, "Turn debugging on/off")
	root.PersistentFlags().BoolP("edit", "e", false, "Edit command")

	// Flags haven't been parsed yet, we need to do it ourselves
	for _, arg := range os.Args {
		if arg == "-d" || arg == "--debug" {
			logrus.SetLevel(logrus.DebugLevel)
		}
	}

	err := loadCommandsInto(root)
	if err != nil {
		panic(err)
	}

	err = root.Execute()
	if err != nil {
		panic(err)
	}
}

func loadCommandsInto(root *cobra.Command) error {
	logrus.Debug("Loading commands started")

	home := filepath.Join(os.Getenv("HOME"), ".sd")
	logrus.Debug("HOME is set to: ", home)

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	logrus.Debug("Current working dir is set to: ", wd)

	current := filepath.Join(wd, "scripts")
	logrus.Debug("Looking for ./scripts in: ", current)

	sdPath := os.Getenv("SD_PATH")
	paths := filepath.SplitList(sdPath)
	logrus.Debug("SD_PATH is set to:", sdPath, ", parsed as: ", paths)

	for _, path := range deduplicate(append([]string{home, current}, paths...)) {
		cmds, err := visitDir(path)
		if err != nil {
			return err
		}

		for _, c := range cmds {
			root.AddCommand(c)
		}
	}

	logrus.Debug("Loading commands done")
	return nil
}

func visitDir(path string) ([]*cobra.Command, error) {
	logrus.Debug("Visiting path: ", path)
	var cmds []*cobra.Command

	if _, err := os.Stat(path); os.IsNotExist(err) {
		logrus.Debug("Path does not exist: ", path)
		return cmds, nil
	}

	items, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		switch {
		case strings.HasPrefix(item.Name(), "."):
			logrus.Debug("Ignoring hidden path: ", filepath.Join(path, item.Name()))
			continue

		case item.IsDir():
			logrus.Debug("Found directory: ", filepath.Join(path, item.Name()))
			cmd := &cobra.Command{
				Use: item.Name(),
			}

			readmePath := filepath.Join(path, item.Name(), "README")
			readme, err := ioutil.ReadFile(readmePath)
			if err == nil {
				logrus.Debug("Found README at: ", readmePath)
				cmd.Short = strings.Split(string(readme), "\n")[0]
				cmd.Long = string(readme)
				cmd.Args = cobra.NoArgs
			}

			subcmds, err := visitDir(filepath.Join(path, item.Name()))
			if err != nil {
				return nil, err
			}
			for _, i := range subcmds {
				cmd.AddCommand(i)
			}

			if cmd.HasSubCommands() {
				logrus.Debug("Directory has scripts (subcommands) inside it: ", filepath.Join(path, item.Name()))
				cmd.RunE = func(cmd *cobra.Command, args []string) error {
					return cmd.Usage()
				}
			}
			cmds = append(cmds, cmd)

		case item.Mode()&0100 != 0:
			logrus.Debug("Script found: ", filepath.Join(path, item.Name()))

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
			logrus.Debug("Found short description line: ", filepath.Join(path), ", set to: ", match[1])
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
			logrus.Debug("Found example line: ", filepath.Join(path), ", set to: ", match[1])
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
		Annotations: map[string]string{
			"Source": path,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			src := cmd.Annotations["Source"]
			edit, err := cmd.Flags().GetBool("edit")
			if err != nil {
				return err
			}
			if edit {
				editor := os.Getenv("VISUAL")
				if editor == "" {
					logrus.Debug("$VISUAL not set, trying $EDITOR...")
					editor = os.Getenv("EDITOR")
					if editor == "" {
					logrus.Debug("$EDITOR not set, trying $(which vim)...")
						editor = "$(which vim)"
					}
				}
				cmdline := []string{"sh", "-c", strings.Join([]string{editor, src}, " ")}
				logrus.Debug("Running ", cmdline)
				return syscall.Exec("/bin/sh", cmdline, os.Environ())

			} else {
				logrus.Debug("Exec: ", src, " with args: ", args)
				return syscall.Exec(src, append([]string{src}, args...), os.Environ())
			}
		},
	}

	logrus.Debug("Created command: ", filepath.Base(path))

	return cmd, nil
}

func completions(root *cobra.Command) *cobra.Command {
	c := &cobra.Command{
		Use:   "completions",
		Short: "Generate completion scripts",
	}

	c.RunE = func(cmd *cobra.Command, args []string) error {
		return c.Usage()
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

	logrus.Debug("Completions (bash/zsh) commands added")
	return c
}

/*
 * deduplicate a slice of strings, keeping the order of the elements
 */
func deduplicate(input []string) []string {
	var output []string
	unique := map[string]interface{}{}
	for _, i := range input {
		unique[i] = new(interface{})
	}
	for _, i := range input {
		if _, ok := unique[i]; ok {
			output = append(output, i)
			delete(unique, i)
		}
	}
	return output
}
