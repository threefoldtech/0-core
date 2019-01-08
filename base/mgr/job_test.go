package mgr

import (
	"fmt"
	"syscall"
	"testing"
	"time"

	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/stream"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	New()
}
func TestJob(t *testing.T) {
	stdin := "hello world"
	cmd := pm.Command{
		Command: pm.CommandSystem,
		Arguments: pm.MustArguments(
			pm.SystemCommandArguments{
				Name:  "cat",
				StdIn: stdin,
			},
		),
	}

	job := newTestJob(&cmd, newSystemProcess)

	job.start(false)

	result := job.Wait()
	if ok := assert.Equal(t, pm.StateSuccess, result.State); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, stdin, result.Streams.Stdout()); !ok {
		t.Error()
	}
}

func TestJobMaxRestart(t *testing.T) {

	var counter int
	var action = func(ctx pm.Context) (interface{}, error) {
		counter++
		return nil, fmt.Errorf("error")
	}

	cmd := pm.Command{
		MaxRestart: 3,
	}

	job := newTestJob(&cmd, NewInternalProcess(action))

	job.start(false)

	result := job.Wait()
	if ok := assert.Equal(t, pm.StateError, result.State); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, 3, counter); !ok {
		t.Error()
	}
}

func TestJobMessages(t *testing.T) {

	var action = func(ctx pm.Context) (interface{}, error) {
		ctx.Log("debug message", pm.LevelDebug)
		ctx.Log("stdout message", pm.LevelStdout)
		ctx.Log("stderr message", pm.LevelStderr)

		return "result data", nil
	}

	cmd := pm.Command{}

	job := newTestJob(&cmd, NewInternalProcess(action))

	var logs []*stream.Message
	var subscriber = func(msg *stream.Message) {
		logs = append(logs, msg)
	}

	job.Subscribe(subscriber)
	job.start(false)

	log.Info("waiting for command to exit")
	result := job.Wait()
	if ok := assert.Equal(t, pm.StateSuccess, result.State); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, `"result data"`, result.Data); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, "stdout message\n", result.Streams.Stdout()); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, "stderr message\n", result.Streams.Stderr()); !ok {
		t.Error()
	}

	//Other levels (like debug) are forwarded to the logger instead, only stdout, stderr and return messages are captured
	//by the result object
	//also any subscriber to the job will get the messages, as we did above
	if ok := assert.Len(t, logs, 4); !ok {
		t.Fatal()
	}

	msg := logs[0] //first message
	if ok := assert.Equal(t, "debug message\n", msg.Message); !ok {
		t.Error()
	}
}

func TestJobTimeout(t *testing.T) {

	cmd := pm.Command{
		Command: pm.CommandSystem,
		Arguments: pm.MustArguments(
			pm.SystemCommandArguments{
				Name: "sleep",
				Args: []string{"10s"},
			},
		),
		MaxTime: 1,
	}

	job := newTestJob(&cmd, newSystemProcess)

	job.start(false)

	result := job.Wait()
	if ok := assert.Equal(t, pm.StateTimeout, result.State); !ok {
		t.Error()
	}
}
func TestJobSignal(t *testing.T) {

	cmd := pm.Command{
		Command: pm.CommandSystem,
		Arguments: pm.MustArguments(
			pm.SystemCommandArguments{
				Name: "sleep",
				Args: []string{"10s"},
			},
		),
	}

	job := newTestJob(&cmd, newSystemProcess)

	go func() {
		time.Sleep(time.Second)
		job.Signal(syscall.SIGINT)
	}()

	job.start(false)

	result := job.Wait()
	if ok := assert.Equal(t, pm.StateError, result.State); !ok {
		t.Error()
	}
}

func TestJobMaxRecurring(t *testing.T) {
	var counter int
	var action = func(ctx pm.Context) (interface{}, error) {
		counter++
		return nil, nil
	}

	cmd := pm.Command{
		RecurringPeriod: 1,
	}

	job := newTestJob(&cmd, NewInternalProcess(action))

	go func() {
		time.Sleep(4 * time.Second)
		job.Unschedule()
	}()

	job.start(false)

	result := job.Wait()
	if ok := assert.Equal(t, pm.StateSuccess, result.State); !ok {
		t.Error()
	}

	if ok := assert.InDelta(t, 4, counter, 1); !ok {
		t.Error()
	}
}

func TestJobHooks(t *testing.T) {
	t.Skip()

	cmd := pm.Command{
		Command: pm.CommandSystem,
		Arguments: pm.MustArguments(
			pm.SystemCommandArguments{
				Name: "sleep",
				Args: []string{"2s"},
			},
		),
	}

	var delayCalled, exitCalled, pidCalled bool
	var delay = func() {
		delayCalled = true
	}

	var exit = func(s bool) {
		exitCalled = true
	}

	var pid = func(i int) {
		pidCalled = true
	}

	hooks := []pm.RunnerHook{
		&pm.DelayHook{Delay: time.Second, Action: delay},
		&pm.ExitHook{Action: exit},
		&pm.PIDHook{Action: pid},
	}

	job := newTestJob(&cmd, newSystemProcess, hooks...)

	job.start(false)

	result := job.Wait()
	if ok := assert.Equal(t, pm.StateSuccess, result.State); !ok {
		t.Error()
	}

	if ok := assert.True(t, delayCalled && exitCalled && pidCalled); !ok {
		t.Error()
	}
}
