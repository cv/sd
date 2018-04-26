package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

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

Looks for lines like this:

# usage: foo arg1 arg2
# usage: foo [arg1] [arg2]

*/
func usageFrom(path string) (string, cobra.PositionalArgs, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", cobra.ArbitraryArgs, err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			logrus.Error(err)
		}
	}()

	r := regexp.MustCompile(`^# usage: (.*)$`)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		match := r.FindStringSubmatch(scanner.Text())
		if len(match) == 2 {
			line := match[1]
			logrus.Debug("Found usage line: ", filepath.Join(path), ", set to: ", line)

			parts := strings.Split(line, " ")
			if len(parts) == 1 {
				logrus.Debug("No args allowed")
				return line, cobra.NoArgs, nil
			}

			var required, optional int
			for _, i := range parts[1:] {
				if i == "..." {
					continue
				}
				if strings.HasPrefix(i, "[") && strings.HasSuffix(i, "]") {
					logrus.Debug("Found optional arg: ", i)
					optional++
				} else {
					logrus.Debug("Found required arg: ", i)
					required++
				}
			}
			if parts[len(parts)-1] == "..." {
				logrus.Debug("Minimum of ", required, " arguments set")
				return match[1], cobra.MinimumNArgs(required), nil
			}
			logrus.Debug("Arg range of ", required, " and ", required+optional, " set")
			return match[1], cobra.RangeArgs(required, required+optional), nil
		}
	}
	logrus.Debug("Any args allowed")
	return filepath.Base(path), cobra.ArbitraryArgs, nil
}

/*

Looks for a line like this:

# example: foo bar 1 2 3

*/
func exampleFrom(path string, cmd *cobra.Command) (string, error) {
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
			return fmt.Sprintf("  %s %s", cmd.UseLine(), match[1]), nil
		}
	}
	return "", nil
}
