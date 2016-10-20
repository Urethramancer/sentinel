package main

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/jessevdk/go-flags"
)

const (
	program = "Sentinel"
	version = "0.4.2"
	// ACTION is the environment variable for the type of notification triggered.
	ACTION = "SENTINEL_ACTION"
	// PATH is the environment variable for the type of notification triggered.
	PATH = "SENTINEL_PATH"
)

var opts struct {
	Verbose bool `short:"v" long:"verbose" description:"Print more details during operation, otherwise remain quiet until an error occurs."`
	Version bool `short:"V" long:"version" description:"Show program version and exit."`
	Flags   struct {
		Create bool `short:"c" long:"create" description:"Watch for new files."`
		Write  bool `short:"w" long:"write" description:"Watch for changed files."`
		Delete bool `short:"d" long:"delete" description:"Watch for deletion."`
		Rename bool `short:"r" long:"rename" description:"Watch for renamed files."`
		Chmod  bool `short:"m" long:"chmod" description:"Watch for attribute changes (date or permissions)."`
	} `group:"Trigger flags"`
	Other struct {
		Loop bool `short:"L" long:"loop" description:"Don't quit after each triggered event."`
	}
	Commands struct {
		CreateAction string `short:"C" long:"createaction" description:"Script to run when a file is created. Implies -c." value-name:"SCRIPT"`
		WriteAction  string `short:"W" long:"writeaction" description:"Script to run when a file is edited. Implies -w." value-name:"SCRIPT"`
		DeleteAction string `short:"D" long:"deleteaction" description:"Script to run when a file is deleted. Implies -d." value-name:"SCRIPT"`
		RenameAction string `short:"R" long:"renameaction" description:"Script to run when a file is renamed. Implies -r." value-name:"SCRIPT"`
		ChmodAction  string `short:"M" long:"chmodaction" description:"Script to run when a file's date or permissions change. Implies -m." value-name:"SCRIPT"`
		ScriptAction string `short:"S" long:"scriptaction" description:"Script to run for all events. Requires any of the trigger flags. Overrides the other scripts." value-name:"SCRIPT"`
	} `group:"Scripts"`
	Args struct {
		Directory []string `positional-arg-name:"PATH"`
	} `positional-args:"yes"`
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		return
	}

	if opts.Version {
		pr("%s %s\n", program, version)
		return
	}

	if len(opts.Args.Directory) == 0 {
		warn("No paths specified.")
	}
	var paths []string
	for _, d := range opts.Args.Directory {
		if !exists(d) {
			warn("Path %s does not exist.", d)
		}
		paths = append(paths, d)
	}

	// Default: Watch for any changes
	var flags fsnotify.Op

	if opts.Commands.CreateAction != "" {
		opts.Flags.Create = true
	}
	if opts.Flags.Create {
		v("Watching for creation.\n")
		flags |= fsnotify.Create
	}

	if opts.Commands.WriteAction != "" {
		opts.Flags.Write = true
	}
	if opts.Flags.Write {
		v("Watching for write.\n")
		flags |= fsnotify.Write
	}

	if opts.Commands.DeleteAction != "" {
		opts.Flags.Delete = true
	}
	if opts.Flags.Delete {
		v("Watching for delete.\n")
		flags |= fsnotify.Remove
	}

	if opts.Commands.RenameAction != "" {
		opts.Flags.Rename = true
	}
	if opts.Flags.Rename {
		v("Watching for rename.\n")
		flags |= fsnotify.Rename
	}

	if opts.Commands.ChmodAction != "" {
		opts.Flags.Chmod = true
	}
	if opts.Flags.Chmod {
		v("Watching for permission changes.\n")
		flags |= fsnotify.Chmod
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if flags&event.Op&fsnotify.Create == fsnotify.Create {
					if opts.Commands.CreateAction != "" {
						v("CREATE: Running '%s'\n", opts.Commands.CreateAction)
						os.Setenv(ACTION, "create")
						os.Setenv(PATH, event.Name)
						runCommand(opts.Commands.CreateAction)
					}
					if !opts.Other.Loop {
						done <- true
					}
				}
				if flags&event.Op&fsnotify.Write == fsnotify.Write {
					if opts.Commands.WriteAction != "" {
						v("WRITE: Running '%s'\n", opts.Commands.WriteAction)
						os.Setenv(ACTION, "write")
						os.Setenv(PATH, event.Name)
						runCommand(opts.Commands.WriteAction)
					}
					if !opts.Other.Loop {
						done <- true
					}
				}
				if flags&event.Op&fsnotify.Remove == fsnotify.Remove {
					if opts.Commands.DeleteAction != "" {
						v("REMOVE: Running '%s'\n", opts.Commands.DeleteAction)
						os.Setenv(ACTION, "delete")
						os.Setenv(PATH, event.Name)
						runCommand(opts.Commands.DeleteAction)
					}
					if !opts.Other.Loop {
						done <- true
					}
				}
				if flags&event.Op&fsnotify.Rename == fsnotify.Rename {
					if opts.Commands.RenameAction != "" {
						v("RENAME: Running '%s'\n", opts.Commands.RenameAction)
						os.Setenv(ACTION, "rename")
						os.Setenv(PATH, event.Name)
						runCommand(opts.Commands.RenameAction)
					}
					if !opts.Other.Loop {
						done <- true
					}
				}
				if flags&event.Op&fsnotify.Chmod == fsnotify.Chmod {
					if opts.Commands.ChmodAction != "" {
						v("CHMOD: Running '%s'\n", opts.Commands.ChmodAction)
						os.Setenv(ACTION, "chmod")
						os.Setenv(PATH, event.Name)
						runCommand(opts.Commands.ChmodAction)
					}
					if !opts.Other.Loop {
						done <- true
					}
				}
			case err := <-watcher.Errors:
				if err.Error() != "" {
					fatal("Error: ", err.Error())
				}
				done <- true
			}
		}
	}()

	for _, dir := range paths {
		v("* %s\n", dir)
		err = watcher.Add(dir)
		if err != nil {
			fatal(err.Error())
		}
	}

	// We'll never return from this without a break signal if in loop mode
	<-done
}

func runCommand(script string) {
	cmd := exec.Command("bash", script)
	err := cmd.Start()
	if err != nil {
		v("Error: %s\n", err)
	}

	err = cmd.Wait()
	if err != nil {
		exit, ok := err.(*exec.ExitError)
		if ok {
			status, ok := exit.Sys().(syscall.WaitStatus)
			if ok {
				if status == 256 || status == 512 {
					os.Exit(0)
					v("Exit code: %d\n", status)
				}
			}
		} else {
			v("Error: %s\n", err)
		}
	}

}
