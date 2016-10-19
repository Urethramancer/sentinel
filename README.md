# Sentinel [![Build Status](https://travis-ci.org/Urethramancer/sentinel.svg)](https://travis-ci.org/Urethramancer/sentinel)
This command-line tool watches one or more directories for files being created, renamed, modified or deleted, then optionally runs a script. It uses the directory change APIs of Linux, macOS and Windows, respectively, thanks to the fsnotify package.

## Dependencies
1. Go version 1.7 or later.
2. [fsnotify](https://github.com/fsnotify/fsnotify)
3. [go-flags](https://github.com/jessevdk/go-flags)

## Build
Clone the repository and use "go build", and any ldflags etc. you prefer.

## Platforms
Tested on Linux and macOS. Should work on Windows without modification.

## Usage
In its simplest use, Sentinel merely watches a directory until an expected type of change occurs, then returns:

```sh
sentinel -c /tmp
```

The above example waits for a file to be created in /tmp, then exits.

You can also tell it to run a simple shell script:

```sh
sentinel -c -C /usr/local/bin/watcher.sh /tmp
```

Or to keep running it again:

```sh
sentinel -c -C /usr/local/bin/watcher.sh /tmp -L
```

The specified script(s) will be given two environment variables, **SENTINEL_ACTION** and **SENTINEL_PATH**, to indicate what happened and to what.

The keyword in **SENTINEL_ACTION** will be **create**, **write**, **delete**, **rename** or **chmod**.

Any non-flag arguments found are taken as directories to watch, and any number you need is allowed. Multiple instances of Sentinel are required if you wish to run different scripts per folder.

### Event flags
The supported types of events to watch for are as follows:

+ -c: Create. New files trigger this.
+ -w: Write. Triggered when an existing file is written to.
+ -d: Delete. Deleting a file from the watched directory will trigger this. Deleting the directory being watched has undetermined results (may vary per platform).
+ -r: Rename.
+ -m: Modification. Changing file permissions, date or attributes will trigger this.

### Command parameters
Each event flag has a corresponding uppercase version to specify a script to run when triggered. The first argument passed to this script will be the name of the directory where something happened.

If you return exit code 1 from a script, Sentinel will quit even in loop mode.

### Other flags
You can get a bit more output by passing the -v flag, or see the current version of Sentinel by passing -V.

## Licence
MIT.
