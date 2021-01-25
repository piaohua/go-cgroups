// +build ignore

package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

const (
	MB = 1024 * 1024
)

func main() {
	fmt.Println("Child pid is", os.Getpid())

	go func() {
		blocks := make([][MB]byte, 0)

		for range time.Tick(5 * time.Second) {
			blocks = append(blocks, [MB]byte{})
			printMemUsage()
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	s := <-c
	fmt.Println("Got signal:", s)

	time.Sleep(5 * time.Second)
}

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tSys = %v MiB \n", bToMb(m.Sys))
}

func bToMb(b uint64) uint64 {
	return b / MB
}
