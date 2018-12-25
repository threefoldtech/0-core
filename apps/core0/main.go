package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	logging "github.com/op/go-logging"
	"github.com/threefoldtech/0-core/apps/core0/assets"
	"github.com/threefoldtech/0-core/apps/core0/bootstrap"
	"github.com/threefoldtech/0-core/apps/core0/options"
	"github.com/threefoldtech/0-core/apps/core0/plugin"
	"github.com/threefoldtech/0-core/apps/core0/screen"
	"github.com/threefoldtech/0-core/base"
	"github.com/threefoldtech/0-core/base/mgr"
	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/settings"
)

var (
	log = logging.MustGetLogger("main")
)

func init() {
	formatter := logging.MustStringFormatter("%{time}: %{color}%{module} %{level:.1s} > %{color:reset} %{message}")
	logging.SetFormatter(formatter)

	//we don't use signal.Ignore because the Ignore is actually inherited by children
	//even after execve which makes all child process not exit when u send them a sigterm or sighup
	var signals []os.Signal
	for i := 1; i <= 31; i++ {
		if syscall.Signal(i) == syscall.SIGUSR1 ||
			syscall.Signal(i) == syscall.SIGCHLD {
			continue
		}
		signals = append(signals, syscall.Signal(i))
	}

	signal.Notify(make(chan os.Signal), signals...)
}

func updateHostnameOnScreen(hostSection *screen.TextSection) {
	for {
		time.Sleep(time.Second * 5)

		hostname, err := os.Hostname()
		if err != nil {
			log.Critical(err.Error())
		} else {
			hostSection.Text = fmt.Sprintf("Hostname: %s", hostname)
		}
	}

}

//Splash setup splash screen
func Splash() {
	if err := screen.New(2); err != nil {
		log.Critical(err)
	}

	screen.Push(&screen.CenteredText{
		TextSection: screen.TextSection{
			Attributes: screen.Attributes{screen.Bold},
			Text:       string(assets.MustAsset("text/logo.txt")),
		},
	})
	screen.Push(&screen.TextSection{})
	screen.Push(&screen.TextSection{
		Attributes: screen.Attributes{screen.Green},
		Text:       fmt.Sprintf("Version: %s", base.Version().Short()),
	})

	screen.Push(&screen.TextSection{
		Attributes: screen.Attributes{screen.Bold},
		Text: fmt.Sprintf("Boot Params: %s",
			options.Options.Kernel.String("debug", "organization", "zerotier", "quiet", "development", "support"), //flags we care about
		),
	})

	screen.Push(&screen.TextSection{})

	hostnameSection := &screen.TextSection{
		Attributes: screen.Attributes{screen.Bold},
		Text:       "",
	}
	screen.Push(hostnameSection)
	screen.Push(&screen.TextSection{})

	go updateHostnameOnScreen(hostnameSection)

	screen.Push(&screen.TextSection{
		Text: "[Alt+F1: Kernel logs view] [Alt+F2: This screen]",
	})
	screen.Push(&screen.TextSection{})
	screen.Refresh()
}

type console struct{}

func (*console) Result(cmd *pm.Command, result *pm.JobResult) {
	log.Debugf("Job result for command '%s' is '%s'", cmd, result.State)
}

func startEntropyGenerator() error {
	log.Debug("starte haveged to generate entropy")
	cmd := exec.Command("haveged", "-w 1024", "-d 32", "-i 32", "-v 1")
	_, err := cmd.CombinedOutput()
	return err
}

func main() {
	var options = options.Options
	fmt.Println(base.Version())
	if options.Version() {
		os.Exit(0)
	}

	if err := startEntropyGenerator(); err != nil {
		log.Fatalf("fail to start entropy generator: %v", err)
	}

	if err := settings.LoadSettings(options.Config()); err != nil {
		log.Fatal(err)
	}

	if errors := settings.Settings.Validate(); len(errors) > 0 {
		for _, err := range errors {
			log.Errorf("%s", err)
		}

		log.Fatalf("\nConfig validation error, please fix and try again.")
	}

	//start process mgr.
	log.Infof("Initialize process manager")
	mgr.New()
	mgr.RegisterExtension("bash", "sh", "", []string{"-c", "{script}"}, nil)

	if !options.Agent() {
		//Logging prepration
		if err := Rotate(); err != nil {
			log.Errorf("failed to setup logging: %s", err)
		}

		//Handle log rotation requests (SIGNALS)
		HandleRotation()
	}

	level := logging.INFO
	if options.Kernel.Is("debug") {
		level = logging.DEBUG
	}

	logging.SetLevel(level, "")

	log.Infof("Configure process manager")
	var config = settings.Settings
	mgr.MaxJobs = config.Main.MaxJobs
	mgr.AddHandle((*console)(nil))
	pluginMgr, err := plugin.New(settings.Settings.Main.Modules...)
	if err != nil {
		log.Fatalf("failed to initialize plugin manager: %s", err)
	}
	mgr.AddRouter(pluginMgr)
	if err := pluginMgr.Load(); err != nil {
		log.Fatalf("failed to load plugins")
	}

	//configure logging handlers from configurations
	//log.Infof("Configure logging")
	// cfg := transport.SinkConfig{Port: 6379}
	// sink, err := transport.NewSink(cfg)
	// if err != nil {
	// 	log.Errorf("failed to start command sink: %s", err)
	// }

	// logger.ConfigureLogging(sink)

	bs := bootstrap.NewBootstrap(options.Agent())
	bs.First()

	if !options.Agent() {
		//Only allow splash screen if debug is not set, or if not running in agent mode
		Splash()
	}

	screen.Push(&screen.TextSection{})
	screen.Push(&screen.SplitterSection{Title: "System Information"})

	row := &screen.RowSection{
		Cells: make([]screen.RowCell, 2),
	}
	screen.Push(row)

	// contMgr, err := containers.ContainerSubsystem(sink, &row.Cells[0])
	// if err != nil {
	// 	log.Fatal("failed to intialize container subsystem", err)
	// }

	bs.Second()

	// if err := kvm.KVMSubsystem(contMgr, &row.Cells[1]); err != nil {
	// 	log.Errorf("failed to initialize kvm subsystem: %s", err)
	// }

	// log.Infof("Starting local transport")
	// local, err := NewLocal(contMgr, "/var/run/core.sock")
	// if err != nil {
	// 	log.Errorf("Failed to start local transport: %s", err)
	// } else {
	// 	local.Start()
	// }

	//start jobs sinks.
	log.Infof("Starting Sinks")

	//sink.Start()
	screen.Refresh()

	// if config.Stats.Enabled {
	// 	aggregator := stats.NewLedisStatsAggregator(sink)
	// 	mgr.AddHandle(aggregator)
	// }

	select {}
}
