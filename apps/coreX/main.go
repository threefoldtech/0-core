package main

import (
	"fmt"
	"os"

	logging "github.com/op/go-logging"
	"github.com/threefoldtech/0-core/apps/coreX/bootstrap"
	"github.com/threefoldtech/0-core/apps/coreX/options"
	"github.com/threefoldtech/0-core/apps/plugins/base/core"
	"github.com/threefoldtech/0-core/apps/plugins/base/fs"
	"github.com/threefoldtech/0-core/apps/plugins/base/info"
	"github.com/threefoldtech/0-core/apps/plugins/base/ip"
	"github.com/threefoldtech/0-core/apps/plugins/base/job"
	"github.com/threefoldtech/0-core/apps/plugins/process"
	"github.com/threefoldtech/0-core/base"
	"github.com/threefoldtech/0-core/base/mgr"
	"github.com/threefoldtech/0-core/base/pm"

	"os/signal"
	"syscall"

	"encoding/json"
)

const (
	//UnlockMagic expected magic from core0 to unlock coreX process
	UnlockMagic = 0x280682
)

var (
	log = logging.MustGetLogger("main")
)

func init() {
	formatter := logging.MustStringFormatter("%{color}%{module} %{level:.1s} > %{message} %{color:reset}")
	logging.SetFormatter(formatter)
	logging.SetLevel(logging.DEBUG, "")
}

func handleSignal(bs *bootstrap.Bootstrap) {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM)
	go func(ch <-chan os.Signal, bs *bootstrap.Bootstrap) {
		<-ch
		log.Infof("Received SIGTERM, terminating.")
		bs.UnBootstrap()
		os.Exit(0)
	}(ch, bs)
}

func main() {
	var opt = options.Options
	fmt.Println(base.Version())
	if opt.Version() {
		os.Exit(0)
	}

	if errors := options.Options.Validate(); len(errors) != 0 {
		for _, err := range errors {
			log.Errorf("Validation Error: %s\n", err)
		}

		os.Exit(1)
	}
	bs := bootstrap.NewBootstrap()
	if err := bs.Start(); err != nil {
		log.Fatalf("failed to initialize container: %s", err)
	}

	//built in plugins
	router, err := NewRouter(
		&core.Plugin,
		&fs.Plugin,
		&info.Plugin,
		&ip.Plugin,
		&job.Plugin,
		&process.Plugin,
	)

	if err != nil {
		log.Fatalf("failed to create router: %s", err)
	}

	mgr.MaxJobs = opt.MaxJobs()
	mgr.New()
	mgr.RegisterExtension("bash", "sh", "", []string{"-c", "{script}"}, nil)

	input := os.NewFile(3, "|input")
	output := os.NewFile(4, "|output")

	dispatcher := NewDispatcher(output)

	//start process mgr.
	log.Infof("Starting process manager")

	mgr.AddHandle(dispatcher)
	mgr.AddRouter(router)
	dec := json.NewDecoder(input)
	//we need to block until we recieve the magic number
	//from core0 this means that the setup from core0 side is complete
	//this include adding the coreX process into the proper cgroups
	var magic int
	if err := dec.Decode(&magic); err != nil {
		log.Fatalf("failed to load unlock magic: %s", err)
	} else if magic != UnlockMagic {
		log.Fatal("invalid magic number")
	}

	log.Info("magic recieved .. continue coreX bootstraping")

	if err := bs.Bootstrap(opt.Hostname()); err != nil {
		log.Fatalf("Failed to bootstrap corex: %s", err)
	}

	handleSignal(bs)

	for {
		var cmd pm.Command
		if err := dec.Decode(&cmd); err != nil {
			log.Errorf("failed to decode command message: %s", err)
		}

		_, err := mgr.Run(&cmd)

		if pm.IsUnknownCommand(err) {
			result := pm.NewJobResult(&cmd)
			result.State = pm.StateUnknownCmd
			dispatcher.Result(&cmd, result)
		} else if err != nil {
			log.Errorf("unknown error while queueing command: %s", err)
		}
	}
}
