package builtin

// #include <sys/syscall.h>
// #include <linux/kexec.h>
import "C"
import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"syscall"
	"unsafe"

	"github.com/pborman/uuid"
	"github.com/threefoldtech/0-core/apps/core0/options"

	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/pm/stream"
)

const (
	cmdReboot   = "core.reboot"
	cmdPowerOff = "core.poweroff"
	cmdUpdate   = "core.update"

	//RedisJobID avoid terminating this process
	RedisJobID      = "redis"
	RedisProxyJobID = "redis-proxy"

	//BaseUpdateURL location to download update image
	BaseUpdateURL = "https://bootstrap.grid.tf/kernel"
	//BaseDownloadLocation download location
	BaseDownloadLocation = "/var/cache"
)

func init() {
	pm.RegisterBuiltIn(cmdReboot, restart)
	pm.RegisterBuiltIn(cmdPowerOff, poweroff)
	pm.RegisterBuiltInWithCtx(cmdUpdate, update)
}

func restart(cmd *pm.Command) (interface{}, error) {
	log.Info("rebooting")
	pm.Shutdown(RedisJobID, RedisProxyJobID)
	syscall.Sync()
	syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
	return nil, nil
}

func poweroff(cmd *pm.Command) (interface{}, error) {
	log.Info("shutting down")
	pm.Shutdown(RedisJobID, RedisProxyJobID)
	syscall.Sync()
	syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
	return nil, nil
}

func wget(ctx *pm.Context, file, url string) error {
	msgs := pm.MessageHook{
		Action: func(msg *stream.Message) {
			//drop the flag part so stream reader (client) doesn't think
			//no more logs are coming on wget termination
			msg.Meta = stream.NewMeta(msg.Meta.Level())
			ctx.Message(msg)
		},
	}

	job, err := pm.Run(&pm.Command{
		ID:      uuid.New(),
		Command: pm.CommandSystem,
		Arguments: pm.MustArguments(
			pm.SystemCommandArguments{
				Name: "wget",
				Args: []string{
					"-O", file, url,
				},
			},
		),
	}, &msgs)

	if err != nil {
		return err
	}

	result := job.Wait()
	if result.State != pm.StateSuccess {
		return fmt.Errorf("failed to download %s: %v", url, result.Streams)
	}

	return nil
}

func update(ctx *pm.Context) (interface{}, error) {
	var args struct {
		Image string `json:"image"`
	}

	if err := json.Unmarshal(*ctx.Command.Arguments, &args); err != nil {
		return nil, err
	}
	img := path.Base(args.Image)
	url := fmt.Sprintf("%s/%s", BaseUpdateURL, img)

	ctx.Log(fmt.Sprintf("downloading image: %s", url))
	file := path.Join(BaseDownloadLocation, img)

	if err := wget(ctx, file, url); err != nil {
		return nil, err
	}

	ctx.Log("download complete")

	fd, err := os.Open(file)
	if err != nil {
		return nil, pm.InternalError(err)
	}

	opts := options.Options
	cmdline := opts.Kernel.Cmdline()
	cmdline = append(cmdline, 0)

	_, _, errno := syscall.Syscall6(
		C.SYS_kexec_file_load,
		fd.Fd(),
		0,
		uintptr(len(cmdline)),
		uintptr(unsafe.Pointer(&cmdline[0])),
		C.KEXEC_FILE_NO_INITRAMFS,
		0,
	)

	if err != nil && uintptr(errno) != 0 {
		log.Error(err)
		return nil, pm.InternalError(err)
	}

	ctx.Log("terminating all running process... point of no return")

	pm.Shutdown(RedisJobID, RedisProxyJobID)
	syscall.Sync()

	ctx.Message(&stream.Message{
		Message: "rebooting...\n",
		Meta:    stream.NewMeta(1, stream.ExitSuccessFlag),
	})

	return nil, syscall.Reboot(syscall.LINUX_REBOOT_CMD_KEXEC)
}
