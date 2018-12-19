package mgr

// #cgo LDFLAGS: -lcap
// #include <sys/capability.h>
import "C"
import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/stream"
)

type jobImb struct {
	command        *pm.Command
	factory        ProcessFactory
	signal         chan syscall.Signal
	unschedule     chan struct{}
	unscheduleOnce sync.Once

	process     pm.Process
	hooks       []pm.RunnerHook
	startTime   time.Time
	backlog     *stream.Buffer
	subscribers []stream.MessageHandler

	o      sync.Once
	result *pm.JobResult
	wg     sync.WaitGroup

	registerPID func(GetPID) (int, error)
	waitPID     func(int) syscall.WaitStatus
	running     int32
}

/*
NewRunner creates a new r object that is bind to this PM instance.

:manager: Bind this r to this PM instance
:command: Command to run
:factory: Process factory associated with command type.
:hooksDelay: Fire the hooks after this delay in seconds if the r is still running. Basically it's a delay for if the
            command didn't exit by then we assume it's running successfully
            values are:
            	- 1 means hooks are only called when the command exits
            	0   means use default delay (default is 2 seconds)
            	> 0 Use that delay

:hooks: Optionals hooks that are called if the r is considered RUNNING successfully.
        The r is considered running, if it ran with no errors for 2 seconds, or exited before the 2 seconds passes
        with SUCCESS exit code.
*/
func newJob(command *pm.Command, factory ProcessFactory, hooks ...pm.RunnerHook) *jobImb {
	job := &jobImb{
		command:    command,
		factory:    factory,
		signal:     make(chan syscall.Signal, 5), //enough buffer for 5 signals
		unschedule: make(chan struct{}),
		hooks:      hooks,
		backlog:    stream.NewBuffer(pm.GenericStreamBufferSize),

		registerPID: registerPID,
		waitPID:     waitPID,
	}

	job.wg.Add(1)
	return job
}

func newTestJob(command *pm.Command, factory ProcessFactory, hooks ...pm.RunnerHook) *jobImb {
	job := newJob(command, factory, hooks...)
	var testTable TestingPIDTable
	job.registerPID = testTable.RegisterPID
	job.waitPID = testTable.WaitPID

	return job
}

func (r *jobImb) Command() *pm.Command {
	return r.command
}

func (r *jobImb) timeout() <-chan time.Time {
	var timeout <-chan time.Time
	if r.command.MaxTime > 0 {
		timeout = time.After(time.Duration(r.command.MaxTime) * time.Second)
	}
	return timeout
}

//set the bounding set for current thread, of course this is un-reversable once set on the
//pm it affects all threads from now on.
func (r *jobImb) setUnprivileged() {
	//drop bounding set for children.
	bound := []uintptr{
		C.CAP_SETPCAP,
		C.CAP_SYS_MODULE,
		C.CAP_SYS_RAWIO,
		C.CAP_SYS_PACCT,
		C.CAP_SYS_ADMIN,
		C.CAP_SYS_NICE,
		C.CAP_SYS_RESOURCE,
		C.CAP_SYS_TIME,
		C.CAP_SYS_TTY_CONFIG,
		C.CAP_AUDIT_CONTROL,
		C.CAP_MAC_OVERRIDE,
		C.CAP_MAC_ADMIN,
		C.CAP_NET_ADMIN,
		C.CAP_SYSLOG,
		C.CAP_DAC_READ_SEARCH,
		C.CAP_LINUX_IMMUTABLE,
		C.CAP_NET_BROADCAST,
		C.CAP_IPC_LOCK,
		C.CAP_IPC_OWNER,
		C.CAP_SYS_PTRACE,
		C.CAP_SYS_BOOT,
		C.CAP_LEASE,
		C.CAP_WAKE_ALARM,
		C.CAP_BLOCK_SUSPEND,
	}

	for _, c := range bound {
		syscall.Syscall6(syscall.SYS_PRCTL, syscall.PR_CAPBSET_DROP,
			c, 0, 0, 0, 0)
	}
}

func (r *jobImb) Subscribe(listener stream.MessageHandler) {
	//TODO: a race condition might happen here because, while we send the backlog
	//a new message might arrive and missed by this listener
	for l := r.backlog.Front(); l != nil; l = l.Next() {
		switch v := l.Value.(type) {
		case *stream.Message:
			listener(v)
		}
	}
	r.subscribers = append(r.subscribers, listener)
}

func (r *jobImb) callback(msg *stream.Message) {
	defer func() {
		//protection against subscriber crashes.
		if err := recover(); err != nil {
			log.Warningf("error in subsciber: %v", err)
		}
	}()

	//check subscribers here.
	msgCallback(r.command, msg)
	for _, sub := range r.subscribers {
		sub(msg)
	}
}

func (r *jobImb) run(unprivileged bool) (jobresult *pm.JobResult) {
	r.startTime = time.Now()
	jobresult = pm.NewJobResult(r.command)
	jobresult.State = pm.StateError

	defer func() {
		jobresult.StartTime = int64(time.Duration(r.startTime.UnixNano()) / time.Millisecond)
		endtime := time.Now()

		jobresult.Time = endtime.Sub(r.startTime).Nanoseconds() / int64(time.Millisecond)

		if err := recover(); err != nil {
			jobresult.State = pm.StateError
			jobresult.Critical = fmt.Sprintf("PANIC(%v)", err)
		}
	}()

	r.process = r.factory(r, r.command)

	ps := r.process
	runtime.LockOSThread()
	if unprivileged {
		r.setUnprivileged()
	}
	channel, err := ps.Run()
	runtime.UnlockOSThread()

	if err != nil {
		var code uint32
		if err, ok := err.(pm.RunError); ok {
			code = err.Code()
		}
		//this basically means r couldn't spawn
		//which indicates a problem with the command itself. So restart won't
		//do any good. It's better to terminate it immediately.
		jobresult.Code = code
		jobresult.Data = err.Error()
		return jobresult
	}

	var result *stream.Message
	var critical string

	stdout := stream.NewBuffer(pm.StandardStreamBufferSize)
	stderr := stream.NewBuffer(pm.StandardStreamBufferSize)

	timeout := r.timeout()

	handlersTicker := time.NewTicker(1 * time.Second)
	defer handlersTicker.Stop()
loop:
	for {
		select {
		case sig := <-r.signal:
			if ps, ok := ps.(pm.Signaler); ok {
				ps.Signal(sig)
			}
		case <-timeout:
			if ps, ok := ps.(pm.Signaler); ok {
				ps.Signal(syscall.SIGKILL)
				jobresult.State = pm.StateTimeout
			}
		case <-handlersTicker.C:
			d := time.Now().Sub(r.startTime)
			for _, hook := range r.hooks {
				go hook.Tick(d)
			}
		case message := <-channel:
			r.backlog.Append(message)

			//messages with Exit flags are always the last.
			if message.Meta.Is(stream.ExitSuccessFlag) {
				jobresult.State = pm.StateSuccess
			}

			if message.Meta.Assert(pm.ResultMessageLevels...) {
				//a result message.
				result = message
			} else if message.Meta.Assert(pm.LevelStdout) {
				stdout.Append(message.Message)
			} else if message.Meta.Assert(pm.LevelStderr) {
				stderr.Append(message.Message)
			} else if message.Meta.Assert(pm.LevelCritical) {
				critical = message.Message
			}

			for _, hook := range r.hooks {
				hook.Message(message)
			}

			//FOR BACKWARD compatibility, we drop the code part from the message meta because watchers
			//like watchdog and such are not expecting a code part in the meta (yet)
			code := message.Meta.Code()
			message.Meta = message.Meta.Base()
			//END of BACKWARD compatibility code

			//by default, all messages are forwarded to the manager for further processing.
			r.callback(message)
			if message.Meta.Is(stream.ExitSuccessFlag | stream.ExitErrorFlag) {
				jobresult.Code = code
				break loop
			}
		}
	}

	r.process = nil

	//consume channel to the end to allow r to cleanup properly
	for range channel {
		//noop.
	}

	if result != nil {
		jobresult.Level = result.Meta.Level()
		jobresult.Data = result.Message
	}

	jobresult.Streams = pm.Streams{
		stdout.String(),
		stderr.String(),
	}

	jobresult.Critical = critical

	return jobresult
}

func (r *jobImb) start(unprivileged bool) {
	atomic.StoreInt32(&r.running, 1)

	runs := 0
	var result *pm.JobResult
	defer func() {
		atomic.StoreInt32(&r.running, 0)
		close(r.signal)
		r.Unschedule()

		if result != nil {
			r.result = result
			callback(r.command, result)

			r.o.Do(func() {
				r.wg.Done()
			})
		}

		cleanUp(r)
	}()

loop:
	for {
		result = r.run(unprivileged)

		for _, hook := range r.hooks {
			hook.Exit(result.State)
		}

		if r.command.Flags.Protected {
			//immediate restart
			log.Debugf("Re-spawning protected service '%s' in 1 second", r.command.ID)
			<-time.After(1 * time.Second)
			continue
		}

		restarting := false
		var restartIn time.Duration

		if result.State != pm.StateSuccess && r.command.MaxRestart > 0 {
			runs++
			if runs < r.command.MaxRestart {
				log.Debugf("Restarting '%s' due to abnormal exit status, trials: %d/%d", r.command, runs+1, r.command.MaxRestart)
				restarting = true
				restartIn = 1 * time.Second
			}
		}

		if r.command.RecurringPeriod > 0 {
			restarting = true
			restartIn = time.Duration(r.command.RecurringPeriod) * time.Second
		}

		if restarting {
			log.Debugf("recurring '%s' in %s", r.command, restartIn)
			select {
			case <-time.After(restartIn):
			case <-r.unschedule:
				break loop
			}
		} else {
			break
		}
	}
}

func (r *jobImb) Unschedule() {
	r.unscheduleOnce.Do(func() {
		close(r.unschedule)
	})
}

func (r *jobImb) Signal(sig syscall.Signal) error {
	if atomic.LoadInt32(&r.running) != 1 {
		return fmt.Errorf("job is not running")
	}

	select {
	case r.signal <- sig:
		return nil
	default:
		return fmt.Errorf("job not receiving signals")
	}
}

func (r *jobImb) Process() pm.Process {
	return r.process
}

func (r *jobImb) Wait() *pm.JobResult {
	r.wg.Wait()
	return r.result
}

//implement PIDTable
//intercept pid registration to fire the correct hooks.
func (r *jobImb) RegisterPID(g GetPID) (int, error) {
	pid, err := r.registerPID(g)
	if err != nil {
		return 0, err
	}

	for _, hook := range r.hooks {
		go hook.PID(pid)
	}

	return 0, nil
}

func (r *jobImb) WaitPID(pid int) syscall.WaitStatus {
	return r.waitPID(pid)
}

func (r *jobImb) StartTime() int64 {
	return int64(time.Duration(r.startTime.UnixNano()) / time.Millisecond)
}