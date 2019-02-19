package power

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
	"github.com/threefoldtech/0-core/base/stream"
)

const (

	//RedisJobID avoid terminating this process
	RedisJobID      = "redis"
	RedisProxyJobID = "redis-proxy"
	ZeroFSID        = "zfs:*"

	//BaseUpdateURL location to download update image
	BaseUpdateURL = "https://bootstrap.grid.tf/kernel"
	//BaseDownloadLocation download location
	BaseDownloadLocation = "/var/cache"
)

func (m *Manager) restart(ctx pm.Context) (interface{}, error) {
	log.Info("rebooting")
	m.api.Shutdown(RedisJobID, RedisProxyJobID, ZeroFSID)
	m.api.Shutdown(RedisJobID, RedisProxyJobID)
	syscall.Sync()

	//we send the message to signal client that job finished
	//before it's actually done
	ctx.Message(&stream.Message{
		Message: "rebooting...\n",
		Meta:    stream.NewMeta(1, stream.ExitSuccessFlag),
	})

	syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
	return nil, nil
}

func (m *Manager) poweroff(ctx pm.Context) (interface{}, error) {
	log.Info("shutting down")
	m.api.Shutdown(RedisJobID, RedisProxyJobID, ZeroFSID)
	m.api.Shutdown(RedisJobID, RedisProxyJobID)
	syscall.Sync()

	//we send the message to signal client that job finished
	//before it's actually done
	ctx.Message(&stream.Message{
		Message: "powering off...\n",
		Meta:    stream.NewMeta(1, stream.ExitSuccessFlag),
	})

	syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
	return nil, nil
}

func (m *Manager) wget(ctx pm.Context, file, url string) error {
	msgs := pm.MessageHook{
		Action: func(msg *stream.Message) {
			//drop the flag part so stream reader (client) doesn't think
			//no more logs are coming on wget termination
			if msg.Meta.Is(stream.ExitErrorFlag | stream.ExitSuccessFlag) {
				return
			}

			ctx.Message(msg)
		},
	}

	job, err := m.api.Run(&pm.Command{
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

func (m *Manager) update(ctx pm.Context) (interface{}, error) {
	var args struct {
		Image string `json:"image"`
	}
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}
	img := path.Base(args.Image)
	url := fmt.Sprintf("%s/%s", BaseUpdateURL, img)

	ctx.Log(fmt.Sprintf("downloading image: %s", url))
	file := path.Join(BaseDownloadLocation, img)

	if err := m.wget(ctx, file, url); err != nil {
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

	ctx.Log("terminating all running process. point of no return...")

	m.api.Shutdown(RedisJobID, RedisProxyJobID, ZeroFSID)
	m.api.Shutdown(RedisJobID, RedisProxyJobID)
	syscall.Sync()

	ctx.Message(&stream.Message{
		Message: "rebooting...\n",
		Meta:    stream.NewMeta(1, stream.ExitSuccessFlag),
	})

	return nil, syscall.Reboot(syscall.LINUX_REBOOT_CMD_KEXEC)
}
