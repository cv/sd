package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

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

/*
Argument validators. Looks for a line like this:

# args: 3

 */
func argsFrom(path string) (cobra.PositionalArgs, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			logrus.Error(err)
		}
	}()

	r := regexp.MustCompile(`^# args: (\d+)$`)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		match := r.FindStringSubmatch(scanner.Text())
		if len(match) == 2 {
			logrus.Debug("Found example line: ", filepath.Join(path), ", set to: ", match[1])
			i, err := strconv.ParseUint(match[1], 10, 32)
			if err != nil {
				return nil, err
			}
			if i == 0 {
				return cobra.NoArgs, nil
			}
			return cobra.ExactArgs(int(i)), nil
		}
	}
	return cobra.ArbitraryArgs, nil
}
