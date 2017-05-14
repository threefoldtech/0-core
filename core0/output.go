package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	LogPath = "/var/log/core.log"
)

var (
	output *os.File
)

func Redirect(p string) error {
	f, err := os.Create(p)
	if err != nil {
		return err
	}

	output = f

	if err := syscall.Dup2(int(f.Fd()), int(os.Stdout.Fd())); err != nil {
		return err
	}

	if err := syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd())); err != nil {
		return err
	}

	return nil
}

func Rotate(p string) error {
	if output != nil {
		output.Close()
		os.Rename(
			output.Name(),
			fmt.Sprintf("%s.%v", output.Name(), time.Now().Format("20060102-0405")),
		)
	}

	return Redirect(p)
}

func HandleRotation() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGUSR1)
	go func() {
		for _ = range ch {
			if err := Rotate(LogPath); err != nil {
				log.Errorf("failed to rotate logs")
			}
		}
	}()
}
