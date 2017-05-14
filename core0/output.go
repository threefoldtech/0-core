package main

import (
	"os"
	"os/signal"
	"syscall"
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
