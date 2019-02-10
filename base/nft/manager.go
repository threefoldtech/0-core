package nft

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	logging "github.com/op/go-logging"
	"github.com/threefoldtech/0-core/base/pm"
)

const (
	//NFTDebug if true, nft files will not be deleted for inspection
	NFTDebug = false
)

var (
	log  = logging.MustGetLogger("nft")
	lock sync.RWMutex
)

//ApplyFromFile applies nft rules from a file
func ApplyFromFile(cfg string) error {
	lock.Lock()
	defer lock.Unlock()

	_, err := pm.System("nft", "-f", cfg)
	return err
}

//Apply (merge) nft rules
func Apply(nft Nft) error {
	data, err := nft.MarshalText()
	if err != nil {
		return err
	}
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
		if !NFTDebug {
			os.RemoveAll(f.Name())
		}
	}()

	if _, err := f.Write(data); err != nil {
		return err
	}
	f.Close()
	log.Debugf("nft applying: %s", f.Name())
	return ApplyFromFile(f.Name())
}

//Drop drops a single rule given a handle
func Drop(family Family, table, chain string, handle int) error {
	lock.Lock()
	defer lock.Unlock()

	_, err := pm.System("nft", "delete", "rule", string(family), table, chain, "handle", fmt.Sprint(handle))
	return err
}

func Find(f ...Filter) ([]FilterRule, error) {
	lock.RLock()
	defer lock.RUnlock()

	job, err := pm.System("nft", "--json", "--handle", "--numeric", "--numeric", "list", "ruleset")
	if err != nil {
		return nil, err
	}

	return filter(job.Streams.Stdout(), f...)
}
