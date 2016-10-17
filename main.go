package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/jessevdk/go-flags"
)

const (
	program = "Sentinel"
	version = "0.1.2"
	ACTION  = "SENTINEL_ACTION"
	PATH    = "SENTINEL_PATH"
)

var opts struct {
	Verbose bool `short:"v" long:"verbose" description:"Print more details during operation, otherwise remain quiet until an error occurs."`
	Version bool `short:"V" long:"version" description:"Show program version and exit."`
	Flags   struct {
		Create bool `short:"c" long:"create" description:"Watch for file creation."`
		Write  bool `short:"w" long:"write" description:"Watch for file editing."`
		Delete bool `short:"d" long:"delete" description:"Watch for file deletion."`
		Rename bool `short:"r" long:"rename" description:"Watch for file renaming."`
		Chmod  bool `short:"m" long:"chmod" description:"Watch for file attribute changes (date or permissions)."`
	} `group:"Flags"`
	Commands struct {
		CreateAction string `short:"C" long:"createaction" description:"Script to run when a file is created." value-name:"CMD"`
		WriteAction  string `short:"W" long:"writeaction" description:"Script to run when a file is edited." value-name:"CMD"`
		DeleteAction string `short:"D" long:"deleteaction" description:"Script to run when a file is deleted." value-name:"CMD"`
		RenameAction string `short:"R" long:"renameaction" description:"Script to run when a file is renamed." value-name:"CMD"`
		ChmodAction  string `short:"M" long:"chmodaction" description:"Script to run when a file's date or permissions change." value-name:"CMD"`
	} `group:"Commands"`
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
		x, _ := filepath.Abs(d)
		if !exists(x) {
			warn("Path %s does not exist.", x)
		}
		paths = append(paths, x)
	}

	// Default: Watch for any changes
	var flags fsnotify.Op

	if opts.Flags.Create {
		v("Watching for creation.\n")
		flags |= fsnotify.Create
	}

	if opts.Flags.Write {
		v("Watching for write.\n")
		flags |= fsnotify.Write
	}

	if opts.Flags.Delete {
		v("Watching for delete.\n")
		flags |= fsnotify.Remove
	}

	if opts.Flags.Rename {
		v("Watching for rename.\n")
		flags |= fsnotify.Rename
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
				if flags&event.Op&fsnotify.Create == fsnotify.Create && opts.Commands.CreateAction != "" {
					v("Running '%s'\n", opts.Commands.CreateAction)
					os.Setenv(ACTION, "create")
					os.Setenv(PATH, event.Name)
					runCommand(opts.Commands.CreateAction)
					done <- true
				}
				if flags&event.Op&fsnotify.Write == fsnotify.Write && opts.Commands.WriteAction != "" {
					v("Running '%s'\n", opts.Commands.WriteAction)
					os.Setenv(ACTION, "write")
					os.Setenv(PATH, event.Name)
					runCommand(opts.Commands.WriteAction)
					done <- true
				}
				if flags&event.Op&fsnotify.Remove == fsnotify.Remove && opts.Commands.DeleteAction != "" {
					v("Running '%s'\n", opts.Commands.DeleteAction)
					os.Setenv(ACTION, "delete")
					os.Setenv(PATH, event.Name)
					runCommand(opts.Commands.DeleteAction)
					done <- true
				}
				if flags&event.Op&fsnotify.Rename == fsnotify.Rename && opts.Commands.RenameAction != "" {
					v("Running '%s'\n", opts.Commands.RenameAction)
					os.Setenv(ACTION, "rename")
					os.Setenv(PATH, event.Name)
					runCommand(opts.Commands.RenameAction)
					done <- true
				}
				if flags&event.Op&fsnotify.Chmod == fsnotify.Chmod && opts.Commands.ChmodAction != "" {
					v("Running '%s'\n", opts.Commands.ChmodAction)
					os.Setenv(ACTION, "chmod")
					os.Setenv(PATH, event.Name)
					runCommand(opts.Commands.ChmodAction)
					done <- true
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

	<-done
}

func runCommand(cmd string) {
	err := exec.Command("bash", cmd).Run()
	if err != nil {
		v("Error: %s\n", err)
	}
}
