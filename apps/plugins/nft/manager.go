package nft

import (
	"fmt"
	"io/ioutil"
	"os"

	logging "github.com/op/go-logging"
)

const (
	//NFTDebug if true, nft files will not be deleted for inspection
	NFTDebug = false
)

var (
	log = logging.MustGetLogger("nft")
)

//ApplyFromFile applies nft rules from a file
func (m *manager) ApplyFromFile(cfg string) error {
	_, err := m.api.System("nft", "-f", cfg)
	return err
}

//Apply (merge) nft rules
func (m *manager) Apply(nft Nft) error {
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
	return m.ApplyFromFile(f.Name())
}

//Drop drops a single rule given a handle
func (m *manager) Drop(family Family, table, chain string, handle int) error {
	_, err := m.api.System("nft", "delete", "rule", string(family), table, chain, "handle", fmt.Sprint(handle))
	return err
}

func (m *manager) Find(filter ...Filter) ([]FilterRule, error) {
	job, err := m.api.System("nft", "--json", "--handle", "--numeric", "--numeric", "list", "ruleset")
	if err != nil {
		return nil, err
	}

	return runFilter(job.Streams.Stdout(), filter...)
}
