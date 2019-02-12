package builtin

import (
	"syscall"

	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/pm/stream"
)

const (
	cmdReboot   = "core.reboot"
	cmdPowerOff = "core.poweroff"

	//RedisJobID avoid terminating this process
	RedisJobID      = "redis"
	RedisProxyJobID = "redis-proxy"
)

func init() {
	pm.RegisterBuiltInWithCtx(cmdReboot, restart)
	pm.RegisterBuiltInWithCtx(cmdPowerOff, poweroff)
}

func restart(ctx *pm.Context) (interface{}, error) {
	log.Info("rebooting")
	pm.Shutdown(RedisJobID, RedisProxyJobID)
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

func poweroff(ctx *pm.Context) (interface{}, error) {
	log.Info("shutting down")
	pm.Shutdown(RedisJobID, RedisProxyJobID)
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
