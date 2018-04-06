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
  bar         Bars the foos.

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

### Homebrew

The easiest way to install and keep `sd` up-to-date for MacOS users is through [Homebrew](https://brew.sh). First, add the `cv/taps` tap to your Homebrew install:

```
$ brew tap cv/taps git@github.com:cv/taps.git
==> Tapping cv/taps
Cloning into '/usr/local/Homebrew/Library/Taps/cv/homebrew-taps'...
remote: Counting objects: 5, done.
remote: Compressing objects: 100% (5/5), done.
remote: Total 5 (delta 0), reused 0 (delta 0), pack-reused 0
Receiving objects: 100% (5/5), done.
Tapped 1 formula (27 files, 23KB)
```

Then install `sd` with `brew install sd`:

```
$ brew install sd
==> Installing sd from cv/taps
==> Downloading https://github.com/cv/sd/releases/download/v0.1.1/sd_0.1.1_Darwin_x86_64.tar.gz
==> Downloading from https://github-production-release-asset-2e65be.s3.amazonaws.com/128149837/9149f9cc-39b3-11e8-98d8-b5bf16da23b7?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAIWNJYAX4CSVEH53A%2
######################################################################## 100.0%
🍺  /usr/local/Cellar/sd/0.1.1: 5 files, 3MB, built in 7 seconds
```

If you have a Go development environment installed, `go get` should work as expected:

```shell
$ go get -u github.com/cv/sd
```

Alternatively, you can grab one of the packages from the [Releases](https://github.com/cv/sd/releases) tab.

## Multiple sources

`sd` loads scripts and dirs in the following order:

- Your `$HOME/.sd` directory
- Script directories listed in `SD_PATH`
- The `scripts` directory under the current location

## Completions

To enable shell completions, making `sd` much more pleasant to use, run:

```shell
$ source <(sd completions bash)
```

Or add it to `/etc/bash-completion.d`, as documented [in this guide](https://debian-administration.org/article/316/An_introduction_to_bash_completion_part_1).

## Contributing

Yes, please! Check out the [issues](https://github.com/cv/sd/issues) and [pull requests](https://github.com/cv/sd/pulls). Any feedback is greatly appreciated!

## Thanks

- [Steve Francia](https://github.com/spf13) et al for [Cobra](https://github.com/spf13/cobra)
- [Fabio Rehm](https://github.com/fgrehm) for lots of feedback and ideas
