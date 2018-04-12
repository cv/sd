package cli

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

// SD is the main interface to running sd
type SD interface {
	Run() error
}

type sd struct {
	root        *cobra.Command
	initialized bool
}

// New returns an instance of SD
func New(version string) SD {
	s := &sd{
		root: &cobra.Command{
			Use:     "sd",
			Version: version,
		},
	}
	s.init()
	return s
}

func (s *sd) init() {
	s.initAliasing()
	s.initCompletions()
	s.initDebugging()
	s.initEditing()

	s.initialized = true
}

func showUsage(cmd *cobra.Command, _ []string) error {
	return cmd.Usage()
}

func (s *sd) Run() error {
	if ! s.initialized {
		return fmt.Errorf("init() not called")
	}

	err := s.loadCommands()
	if err != nil {
		logrus.Debugf("Error loading commands: %v", err)
		return err
	}

	err = s.root.Execute()
	if err != nil {
		logrus.Debugf("Error executing command: %v", err)
		return err
	}

	return nil
}

func (s *sd) initAliasing() {
	s.root.PersistentFlags().StringP("alias", "a", "sd", "Use an alias in help text and completions")
	err := s.root.PersistentFlags().MarkHidden("alias")
	if err != nil {
		panic(err)
	}

	s.root.PersistentPreRunE = func(_ *cobra.Command, args []string) error {
		alias, err := s.root.PersistentFlags().GetString("alias")
		if err != nil {
			return err
		}
		s.root.Use = alias
		logrus.Debug("Aliasing: sd replaced with ", alias, " in help text")
		return nil
	}
	s.root.RunE = showUsage
}

func (s *sd) initCompletions() {
	c := &cobra.Command{
		Use:   "completions",
		Short: "Generate completion scripts",
		RunE:  showUsage,
	}

	c.AddCommand(&cobra.Command{
		Use:   "bash",
		Short: "Generate completions for bash",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenBashCompletion(os.Stdout)
		},
	})

	c.AddCommand(&cobra.Command{
		Use:   "zsh",
		Short: "Generate completions for zsh",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenZshCompletion(os.Stdout)
		},
	})

	logrus.Debug("Completions (bash/zsh) commands added")
	s.root.AddCommand(c)
}

func (s *sd) initDebugging() {
	s.root.PersistentFlags().BoolP("debug", "d", false, "Turn debugging on/off")

	// Flags haven't been parsed yet, we need to do it ourselves
	for _, arg := range os.Args {
		if arg == "-d" || arg == "--debug" {
			logrus.SetLevel(logrus.DebugLevel)
		}
	}
}

func (s *sd) initEditing() {
	s.root.PersistentFlags().BoolP("edit", "e", false, "Edit command")
}

func (s *sd) loadCommands() error {
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
			s.root.AddCommand(c)
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
				cmd.RunE = showUsage
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
				cmd.RunE = showUsage
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
	defer func() {
		err = file.Close()
		if err != nil {
			logrus.Error(err)
		}
	}()

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
	defer func() {
		err = file.Close()
		if err != nil {
			logrus.Error(err)
		}
	}()

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
		RunE: execCommand,
	}

	logrus.Debug("Created command: ", filepath.Base(path))
	return cmd, nil
}

// these get mocked in tests
var (
	syscallExec = syscall.Exec
	env = os.Getenv
)

func execCommand(cmd *cobra.Command, args []string) error {
	src := cmd.Annotations["Source"]
	edit, err := cmd.Root().PersistentFlags().GetBool("edit")
	if err != nil {
		return err
	}

	if edit {
		editor := env("VISUAL")
		if editor == "" {
			logrus.Debug("$VISUAL not set, trying $EDITOR...")
			editor = env("EDITOR")
			if editor == "" {
				logrus.Debug("$EDITOR not set, trying $(which vim)...")
				editor = "$(which vim)"
			}
		}
		cmdline := []string{"sh", "-c", strings.Join([]string{editor, src}, " ")}
		logrus.Debug("Running ", cmdline)
		return syscallExec("/bin/sh", cmdline, os.Environ())
	}

	logrus.Debug("Exec: ", src, " with args: ", args)
	return syscallExec(src, append([]string{src}, args...), os.Environ())
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
