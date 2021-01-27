// +build ignore

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/piaohua/go-cgroups"
)

var (
	// command
	exec string

	// cgroups path name
	name string

	// BuildTime build time
	BuildTime = ""

	// Commit will be overwritten automatically by the build system
	Commit = "HEAD"

	// command line arguments
	args argFlags

	// help
	h bool
)

type argFlags []string

func (a *argFlags) Set(s string) error {
	*a = append(*a, s)
	return nil
}

func (a *argFlags) String() string {
	return strings.Join(*a, " ")
}

func init() {
	flag.Var(&args, "args", "command line arguments")
	flag.BoolVar(&h, "h", true, "help")
	flag.StringVar(&exec, "exec", "", "Run a command in cgroups")
	flag.StringVar(&name, "name", "", "cgroups path name")
}

func main() {
	flag.Parse()
	fmt.Printf("%s@%s\n", BuildTime, Commit)
	if h {
		flag.PrintDefaults()
		return
	}
	if exec == "" {
		panic("missing command to be executed")
	}

	go cgroups.Run(name, exec, args...)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	s := <-c
	fmt.Println("Got signal:", s)
}
