package bootstrap

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
	"syscall"
	"time"

	logging "github.com/op/go-logging"
	cache "github.com/patrickmn/go-cache"
	"github.com/threefoldtech/0-core/apps/core0/bootstrap/network"
	"github.com/threefoldtech/0-core/apps/core0/options"
	"github.com/threefoldtech/0-core/apps/core0/screen"
	"github.com/threefoldtech/0-core/base/mgr"
	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/settings"
	"github.com/threefoldtech/0-core/base/utils"
	"github.com/vishvananda/netlink"
)

const (
	InternetTestAddress = "google.com"
	InternetTestPort    = "80"

	screenStateLine = "->%25s: %s %s"
)

var (
	log = logging.MustGetLogger("bootstrap")
)

type Bootstrap struct {
	i     *settings.IncludedSettings
	t     settings.StartupTree
	agent bool
	rs    *cache.Cache
}

func NewBootstrap(agent bool) *Bootstrap {
	included, errors := settings.Settings.GetIncludedSettings()
	if len(errors) > 0 {
		for _, err := range errors {
			log.Errorf("%s", err)
		}
	}

	//startup services from [init, net[
	t, errors := included.GetStartupTree()

	if len(errors) > 0 {
		//print service tree errors (cyclic dependencies, or missing dependencies)
		for _, err := range errors {
			log.Errorf("%s", err)
		}
	}

	b := &Bootstrap{
		i:     included,
		t:     t,
		agent: agent,
		rs:    cache.New(20*time.Minute, 5*time.Minute),
	}

	return b
}

func (b *Bootstrap) registerExtensions(extensions map[string]settings.Extension) {
	for extKey, extCfg := range extensions {
		log.Infof("registering extension (%s)", extKey)
		if err := mgr.RegisterExtension(extKey, extCfg.Binary, extCfg.Cwd, extCfg.Args, extCfg.Env); err != nil {
			log.Error(err)
		}
	}
}

func (b *Bootstrap) startupServices(s, e settings.After) {
	log.Debugf("Starting up '%s' services", s)
	slice := b.t.Slice(s.Weight(), e.Weight())
	mgr.RunSlice(slice)
	log.Debugf("'%s' services are booted", s)
}

func (b *Bootstrap) resolve(host string) (string, error) {
	if addr, ok := b.rs.Get(host); ok {
		return addr.(string), nil
	}

	addresses, err := net.DefaultResolver.LookupHost(context.Background(), host)
	if err != nil {
		return "", err
	}

	b.rs.Set(host, addresses[0], cache.DefaultExpiration)
	return addresses[0], nil
}

func (b *Bootstrap) canReachInternet() bool {
	addr, err := b.resolve(InternetTestAddress)
	if err != nil {
		log.Debugf("failed to resolve '%s': %s", InternetTestAddress, err)
		return false
	}
	con, err := net.Dial("tcp", addr+":"+InternetTestPort)
	if err != nil {
		return false
	}
	con.Close()
	return true
}

func (b *Bootstrap) ipsAsString(ips []netlink.Addr) string {
	var s []string
	for _, ip := range ips {
		s = append(s, ip.IPNet.String())
	}

	return strings.Join(s, ", ")
}

func (b *Bootstrap) setupNetworking() error {
	if settings.Settings.Main.Network == "" {
		log.Warning("No network config file found, skipping network setup")
		return nil
	}

	netMgr, err := network.GetNetworkManager(settings.Settings.Main.Network)
	if err != nil {
		return err
	}

	if err := netMgr.Initialize(); err != nil {
		return err
	}

	interfaces, err := netMgr.Interfaces()
	if err != nil {
		return fmt.Errorf("failed to get network interfaces: %s", err)
	}

	ignore, _ := options.Options.Kernel.Get("noautonic")

	//apply the interfaces settings as configured.
	for _, inf := range interfaces {
		if utils.InString(ignore, inf.Name()) {
			log.Infof("skipping auto config for interface '%s'", inf.Name())
			continue
		}

		log.Infof("Setting up interface '%s'", inf.Name())

		inf.Clear()
		inf.SetUP(true)
		go func(inf network.Interface) {
			if err := inf.Configure(); err != nil {
				log.Errorf("%s", err)
				inf.SetUP(false)
			}
		}(inf)
	}

	log.Debugf("waiting for internet reachability")
	now := time.Now()
	for time.Since(now) < 2*time.Minute {
		if b.canReachInternet() {
			log.Info("can reach the internet")
			return nil
		}
		<-time.After(3 * time.Second)
	}

	log.Warning("can not reach interent, continue booting anyway")
	return nil
}

func (b *Bootstrap) screen() {
	section := &screen.GroupSection{
		Sections: []screen.Section{},
	}

	screen.Push(section)

	progress := &screen.ProgressSection{}
	reachable := "All Interfaces"
	if options.Options.Kernel.Is("zerotier") {
		reachable = "Zerotier Only"
	}
	reachability := &screen.TextSection{
		Text: fmt.Sprintf(screenStateLine, "Reachability", reachable, ""),
	}
	const refreshEvery = 5
	const refreshNetStateEvery = 10 * 60 / refreshEvery //10min
	netUpdate := refreshNetStateEvery
	for {
		links, err := netlink.LinkList()
		if err != nil {
			<-time.After(10 * time.Second)
			continue
		}
		section.Sections = []screen.Section{}

		for _, link := range links {
			name := link.Attrs().Name
			if name == "lo" || !utils.InString([]string{"device", "tun", "tap", "openvswitch"}, link.Type()) {
				continue
			}
			if strings.HasPrefix(name, "tun") || strings.HasPrefix(name, "tap") {
				continue
			}

			ips, _ := netlink.AddrList(link, netlink.FAMILY_V4)
			section.Sections = append(section.Sections, &screen.TextSection{
				Text: fmt.Sprintf(screenStateLine, name, link.Attrs().HardwareAddr, b.ipsAsString(ips)),
			})
		}

		section.Sections = append(section.Sections, progress, reachability)
		netUpdate++
		if netUpdate >= refreshNetStateEvery {
			progress.Enter()
			progress.Text = fmt.Sprintf(screenStateLine, "Internet Connectivity", "", "")
			if b.canReachInternet() {
				progress.Text = fmt.Sprintf(screenStateLine, "Internet Connectivity", "OK", "")
				netUpdate = 0 //reset counter
			} else {
				progress.Text = fmt.Sprintf(screenStateLine, "Internet Connectivity", "NOT OK", "")
			}
		}

		progress.Leave()
		screen.Refresh()
		<-time.After(refreshEvery * time.Second)
	}
}

func (b *Bootstrap) watchers() {
	//uptime watcher

	//zerotier watcher
	screen.Push(&screen.SplitterSection{
		Title: "Watchers",
	})

	zerotier := &screen.TextSection{}
	uptime := &screen.TextSection{}
	screen.Push(zerotier)
	screen.Push(uptime)

	go func() {
		for {
			result, err := mgr.System("zerotier-cli", "-D/tmp/zt", "info")
			if err == nil {
				ztstatus := result.Streams.Stdout()

				ztstatus = strings.TrimSpace(ztstatus)
				zerotier.Text = fmt.Sprintf(screenStateLine, "Zerotier", ztstatus, "")
			}

			result, err = mgr.System("uptime")
			if err == nil {
				uptimestatus := result.Streams.Stdout()
				uptime.Text = fmt.Sprintf(screenStateLine, "Uptime", uptimestatus, "")
			}

			screen.Refresh()
			<-time.After(30 * time.Second)
		}
	}()
}

func (b *Bootstrap) syslogd() {
	mgr.Run(&pm.Command{
		ID:      "syslogd",
		Command: pm.CommandSystem,
		Arguments: pm.MustArguments(
			pm.SystemCommandArguments{
				Name: "syslogd",
				Args: []string{
					"-n",
					"-O", "/var/log/messages",
				},
			},
		),
		Flags: pm.JobFlags{Protected: true},
	})

	mgr.Run(&pm.Command{
		ID:      "klogd",
		Command: pm.CommandSystem,
		Arguments: pm.MustArguments(
			pm.SystemCommandArguments{
				Name: "klogd",
				Args: []string{
					"-n",
				},
			},
		),
		Flags: pm.JobFlags{Protected: true},
	})

	mgr.System("dmesg", "-n", "1")
}

func (b *Bootstrap) First() {
	if !b.agent {
		if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &syscall.Rlimit{65536, 65536}); err != nil {
			log.Errorf("failed to setup max open files limit: %s", err)
		}

		if err := ioutil.WriteFile(path.Join("/proc", fmt.Sprint(os.Getpid()), "oom_score_adj"), []byte("-1000"), 0644); err != nil {
			log.Errorf("failed to adjust oom score")
		}
	}

	//register core extensions
	b.registerExtensions(settings.Settings.Extension)

	//register included extensions
	b.registerExtensions(b.i.Extension)

	b.syslogd()
}

//Bootstrap registers extensions and startup system services.
func (b *Bootstrap) Second() {
	progress := &screen.ProgressSection{
		Text: "Bootstrapping: Core Services",
	}
	screen.Push(progress)
	progress.Enter()

	//start up all init services ([init, net[ slice)
	b.startupServices(settings.AfterInit, settings.AfterNet)

	go b.screen()

	if !b.agent {
		progress.Text = "Bootstrapping: Networking"
		for {
			err := b.setupNetworking()
			if err == nil {
				break
			}

			log.Errorf("Failed to configure networking: %s", err)
			log.Infof("Retrying in 2 seconds")

			<-time.After(2 * time.Second)
			log.Infof("Retrying setting up network")
		}
	}

	progress.Text = "Bootstrapping: Network Services"

	//start up all net services ([net, boot[ slice)
	b.startupServices(settings.AfterNet, settings.AfterBoot)

	progress.Text = "Bootstrapping: Services"

	//start up all boot services ([boot, end] slice)
	b.startupServices(settings.AfterBoot, settings.ToTheEnd)

	progress.Text = "Bootstrapping: Done"
	progress.Leave()

	b.watchers()

	screen.Refresh()
}
