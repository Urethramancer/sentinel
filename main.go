package main

import (
	"os/exec"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/jessevdk/go-flags"
)

const (
	PROGRAMNAME = "Sentinel"
	VERSION     = "0.1.1"
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
		CreateAction string `short:"C" long:"createaction" description:"Command to run when a file is created." value-name:"CMD"`
		WriteAction  string `short:"W" long:"writeaction" description:"Command to run when a file is edited." value-name:"CMD"`
		DeleteAction string `short:"D" long:"deleteaction" description:"Command to run when a file is deleted." value-name:"CMD"`
		RenameAction string `short:"R" long:"renameaction" description:"Command to run when a file is renamed." value-name:"CMD"`
		ChmodAction  string `short:"M" long:"chmodaction" description:"Command to run when a file's date or permissions change." value-name:"CMD"`
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
		pr("%s %s\n", PROGRAMNAME, VERSION)
		return
	}

	if len(opts.Args.Directory) == 0 {
		warn("No paths specified.")
	}
	paths := make([]string, 0)
	for _, d := range opts.Args.Directory {
		x, _ := filepath.Abs(d)
		if !exists(x) {
			warn("Path %s does not exist.", x)
		}
		paths = append(paths, x)
	}

	// Default: Watch for any changes
	var flags fsnotify.Op = 0

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
				if flags&event.Op&fsnotify.Create == fsnotify.Create {
					if opts.Commands.CreateAction != "" {
						cmd, _ := filepath.Abs(opts.Commands.CreateAction)
						v("Running '%s'\n", cmd)
						res := exec.Command(cmd, event.Name)
						res.Run()
					}
					done <- true
				}
				if flags&event.Op&fsnotify.Write == fsnotify.Write {
					if opts.Commands.WriteAction != "" {
						cmd, _ := filepath.Abs(opts.Commands.WriteAction)
						v("Running '%s'\n", cmd)
						res := exec.Command(cmd, event.Name)
						res.Run()
					}
					done <- true
				}
				if flags&event.Op&fsnotify.Remove == fsnotify.Remove {
					if opts.Commands.DeleteAction != "" {
						cmd, _ := filepath.Abs(opts.Commands.DeleteAction)
						v("Running '%s'\n", cmd)
						res := exec.Command(cmd, event.Name)
						res.Run()
					}
					done <- true
				}
				if flags&event.Op&fsnotify.Rename == fsnotify.Rename {
					if opts.Commands.RenameAction != "" {
						cmd, _ := filepath.Abs(opts.Commands.RenameAction)
						v("Running '%s'\n", cmd)
						res := exec.Command(cmd, event.Name)
						res.Run()
					}
					done <- true
				}
				if flags&event.Op&fsnotify.Chmod == fsnotify.Chmod {
					if opts.Commands.ChmodAction != "" {
						cmd, _ := filepath.Abs(opts.Commands.ChmodAction)
						v("Running '%s'\n", cmd)
						res := exec.Command(cmd, event.Name)
						res.Run()
					}
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
