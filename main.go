// +build ignore

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/piaohua/go-cgroups"
)

var (
	// command
	exec string

	// cgroups path name
	name string

	// Version release version
	Version = "0.0.1"

	// Commit will be overwritten automatically by the build system
	Commit = "HEAD"
)

func init() {
	flag.StringVar(&exec, "exec", "", "Run a command in cgroups")
	flag.StringVar(&name, "name", "", "cgroups path name")
}

func main() {
	flag.Parse()
	fmt.Printf("%s@%s\n", Version, Commit)
	if exec == "" {
		panic("missing command to be executed")
	}

	go cgroups.Run(exec, name)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	s := <-c
	fmt.Println("Got signal:", s)
}
