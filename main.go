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

// These variables are set by the build process (see Makefile)
var (
	// Version of the application (SemVer)
	version string
)

func main() {
	root := &cobra.Command{
		Use:     "sd",
		Version: version,
	}
	root.AddCommand(completions(root))
	root.PersistentFlags().BoolP("debug", "d", false, "Turn debugging on/off")

	err := root.ParseFlags(os.Args)
	if err != nil {
		panic(err)
	}

	debug, err := root.PersistentFlags().GetBool("debug")
	if err != nil {
		panic(err)
	}

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	err = loadCommandsInto(root)
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

	for _, path := range append([]string{home, current}, paths...) {
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return sh(path, args)
		},
	}

	logrus.Debug("Created command: ", filepath.Base(path))

	return cmd, nil
}

func sh(cmd string, args []string) error {
	logrus.Debug("Exec: ", cmd, " with args: ", args)
	return syscall.Exec(cmd, append([]string{cmd}, args...), os.Environ())
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
