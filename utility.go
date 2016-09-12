package main

import (
	"fmt"
	"os"
)

func pr(f string, v ...interface{}) {
	fmt.Printf(f, v...)
}

func v(f string, v ...interface{}) {
	if !opts.Verbose {
		return
	}
	fmt.Printf(f, v...)
}

func warn(f string, v ...interface{}) {
	fmt.Printf(f+"\n", v...)
	os.Exit(1)
}

func fatal(f string, v ...interface{}) {
	fmt.Printf(f+"\n", v...)
	os.Exit(2)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
