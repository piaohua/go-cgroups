package cgroups

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// Run a command in cgroups
func Run(command, path string) {
	if path == "" {
		path = randomHash()
	}
	path = sha256Sum(path)[:12]

	cg := NewCGroup(path)

	if err := cg.setCPU(); err != nil {
		panic(err)
	}

	if err := cg.setCpuset(); err != nil {
		panic(err)
	}

	if err := cg.setMemory(); err != nil {
		panic(err)
	}

	if err := cg.setPids(); err != nil {
		panic(err)
	}

	//TODO defer cg remove

	cg.startCmd(command)
}

// randomHash creates random Length 64 hash
func randomHash() string {
	randBuffer := make([]byte, 32)
	rand.Read(randBuffer)
	sha := sha256.New().Sum(randBuffer)
	return fmt.Sprintf("%x", sha)[:64]
}

func sha256Sum(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

type exitStatus struct {
	Signal os.Signal
	Code   int
}

func (cg *CGroups) startCmd(command string) {
	restart := make(chan exitStatus, 1)

	runner := func() {
		cmd := exec.Cmd{
			Path: command,
		}
		//cmd.SysProcAttr = &syscall.SysProcAttr{
		//	Cloneflags: syscall.CLONE_NEWNS |
		//		syscall.CLONE_NEWUTS |
		//		syscall.CLONE_NEWPID |
		//		syscall.CLONE_NEWNET |
		//		syscall.CLONE_NEWIPC,
		//}

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// start app
		if err := cmd.Start(); err != nil {
			panic(err)
		}

		fmt.Println("add pid", cmd.Process.Pid, "to file cgroup.procs")

		// set cgroup procs id
		cg.Pid = cmd.Process.Pid
		if err := cg.setProcs(); err != nil {
			panic(err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		signalChannel := make(chan os.Signal, 2)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGCHLD)
		go func(ctx context.Context) {
			select {
			case <-ctx.Done():
				return
			case sig := <-signalChannel:
				switch sig {
				case os.Interrupt, syscall.SIGTERM:
					cmd.Process.Signal(sig)
				case syscall.SIGCHLD:
					//TODO handle SIGCHLD
				}
			}
		}(ctx)

		if err := cmd.Wait(); err != nil {
			fmt.Println("cmd return with error:", err)
		}

		status := cmd.ProcessState.Sys().(syscall.WaitStatus)

		options := exitStatus{
			Code: status.ExitStatus(),
		}

		if status.Signaled() {
			options.Signal = status.Signal()
		}

		cmd.Process.Kill()

		restart <- options
	}

	go runner()

	for {
		status := <-restart

		switch status.Signal {
		case os.Kill:
			fmt.Println("app is killed by system")
		default:
			fmt.Println("app exit with code:", status.Code)
		}

		fmt.Println("restart app..")

		go runner()
	}
}
