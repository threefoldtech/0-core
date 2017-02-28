package bootstrap

import (
	"fmt"
	"os"
	"syscall"

	"github.com/g8os/core0/base/pm"
	"github.com/g8os/core0/base/pm/core"
	"github.com/g8os/core0/base/pm/process"
	"github.com/op/go-logging"
	"github.com/pborman/uuid"
)

var (
	log = logging.MustGetLogger("bootstrap")
)

type Bootstrap struct {
}

func NewBootstrap() *Bootstrap {
	return &Bootstrap{}
}

//Bootstrap registers extensions and startup system services.
func (b *Bootstrap) Bootstrap(hostname string) error {
	log.Infof("Mounting proc")
	if err := syscall.Mount("none", "/proc", "proc", 0, ""); err != nil {
		return err
	}

	if err := syscall.Mount("none", "/dev", "devtmpfs", 0, ""); err != nil {
		return err
	}

	if err := syscall.Mount("none", "/dev/pts", "devpts", 0, ""); err != nil {
		return err
	}

	if err := updateHostname(hostname); err != nil {
		return err
	}

	return nil
}

func updateHostname(hostname string) error {
	log.Infof("Set hostname to %s", hostname)

	// update /etc/hostname
	fHostname, err := os.OpenFile("/etc/hostname", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fHostname.Close()
	fmt.Fprint(fHostname, hostname)

	// update /etc/hosts
	fHosts, err := os.OpenFile("/etc/hosts", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fHosts.Close()
	fmt.Fprintf(fHosts, "127.0.0.1    %s.local %s\n", hostname, hostname)
	fmt.Fprint(fHosts, "127.0.0.1    localhost.localdomain localhost\n")

	// call hostname command
	runner, err := pm.GetManager().RunCmd(&core.Command{
		ID:      uuid.New(),
		Command: process.CommandSystem,
		Arguments: core.MustArguments(process.SystemCommandArguments{
			Name: "hostname",
			Args: []string{hostname},
		}),
	})

	if err != nil {
		return err
	}
	result := runner.Wait()
	if result.State != core.StateSuccess {
		return fmt.Errorf("failed to set hostname: %v", result.Streams)
	}

	return nil
}
