# sd: Scripts Dir

A tool to keep utility scripts neatly organized.

`sd` scans and provides completion for a nested tree of executable script files. For example, if you want to be able to run:

```shell
sd foo bar 123
```

All you need to do is create the following structure:

```
~/.sd/
  |- foo/
  |  |- README
     |- bar
```

The first line of `README` is a short description of `foo`. Like this:

```
$ sd --help
Usage:
  sd [command]

Available Commands:
  completions Generate completion scripts
  foo         Commands related to foo
  help        Help about any command

Flags:
  -h, --help   help for sd

Use "sd [command] --help" for more information about a command.
```

The rest of the `README` file gets displayed when the user asks for further help:

```
$ sd foo --help
Commands related to foo

This is the longer text description of all the subcommands, switches
and examples of foo. It is displayed when `sd foo --help` is called.

Usage:
  sd foo [flags]
  sd foo [command]

Available Commands:
  bar         Bars the frobs.

Flags:
  -h, --help   help for foo

Use "sd foo [command] --help" for more information about a command.
```

The `bar` script *must* be marked executable (`chmod +x`). The help text for it looks like this:

```
$ sd foo bar --help
Bars the foos.

Usage:
  sd foo bar [flags]

Examples:
  sd foo bar 123

Flags:
  -h, --help   help for bar
```

In order to document the script, `sd` pays attention to a few special comments:

```shell
#!/bin/sh
#
# bar: Bars the foos.
#
# example: foo bar 123
#

echo "sd foo bar"
```

More will be added in the future, so you'll be able to specify and document flags, environment variables, and so on.

## Installing

If you have a Go development environment installed, this should be a piece of cake:

```
$ go get -u github.com/cv/sd
```

APT, Homebrew and other packages coming up soon.

## Completions

To enable shell completions, making `sd` much more pleasant to use, run:

```bash
source <(sd completions bash)
```

Or add it to `/etc/bash-completion.d`, as documented [in this guide](https://debian-administration.org/article/316/An_introduction_to_bash_completion_part_1).

## Contributing

Yes, please.
