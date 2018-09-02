package containers

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/threefoldtech/0-core/apps/core0/helper/socat"
	"github.com/threefoldtech/0-core/base/pm"
	"gopkg.in/yaml.v2"
)

const (
	cmdFlistCreate = "corex.flist.create"
)

func zflist(args ...string) (*pm.JobResult, error) {
	log.Debugf("zflist %v", args)
	return pm.System("zflist", args...)
}

func containerPath(container *container, path string) string {
	return filepath.Join(container.Root, path)
}

type createArgs struct {
	Container uint16 `json:"container"`
	Flist     string `json:"flist"`   //path where to create the flist
	Storage   string `json:"storage"` // zdb://host:port to the data storage
	Src       string `json:"src"`     //path to the directory to create flist from
}

type router struct {
	Pools  map[string]map[string]string `yaml:"pools,flow"`
	Lookup []string                     `yaml:"lookup,flow"`
}

func (c createArgs) Validate() error {
	if c.Container <= 0 {
		return fmt.Errorf("invalid container id")
	}
	if c.Flist == "" {
		return fmt.Errorf("flist destination need to be specified")
	}
	if c.Storage == "" {
		return fmt.Errorf("flist data storage need to be specified")
	}
	if c.Src == "" {
		return fmt.Errorf("source directory need to be specified")
	}
	return nil
}

func (m *containerManager) flistCreate(cmd *pm.Command) (interface{}, error) {
	var args createArgs

	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	if err := args.Validate(); err != nil {
		return nil, err
	}

	m.conM.RLock()
	cont, ok := m.containers[args.Container]
	m.conM.RUnlock()

	if !ok {
		return nil, fmt.Errorf("container does not exist")
	}

	//pause container
	//TODO: avoid race if cont has just started and pid is not set yet!
	if cont.PID == 0 {
		return nil, fmt.Errorf("container is not fully started yet")
	}

	//pause container
	syscall.Kill(-cont.PID, syscall.SIGSTOP)
	defer syscall.Kill(-cont.PID, syscall.SIGCONT)

	archivePath := containerPath(cont, args.Flist)
	srcPath := containerPath(cont, args.Src)

	// create flist
	storage, err := socat.ResolveURL(args.Storage)
	if err != nil {
		return nil, fmt.Errorf("failed to process storage url: %s", err)
	}

	_, err = zflist("--archive", archivePath, "-args.Storageargs.Storage-create", srcPath, "--backend", storage)
	if err != nil {
		return nil, err
	}

	// add the router.yaml to the flist archive
	router := router{
		Pools: map[string]map[string]string{
			"private": map[string]string{"00:FF": fmt.Sprintf("zdb://%s", args.Storage)},
		},
		Lookup: []string{"private"},
	}
	routerb, err := yaml.Marshal(router)
	if err != nil {
		return nil, err
	}

	return archivePath, addRouterFile(archivePath, routerb)
}

// addRouterFile add a router.yaml file to the flist tar archive
func addRouterFile(flist string, router []byte) error {
	// extract flist archive
	f, err := os.Open(flist)
	if err != nil {
		log.Error("error opening flist: %v", err)
		return err
	}
	defer f.Close()

	untarFlist := flist + ".d"
	if err := Untar(untarFlist, f); err != nil {
		log.Errorf("fail to untar flist: %v", err)
		return err
	}
	defer func() {
		os.RemoveAll(untarFlist)
	}()

	// add the router.yaml file to the archive directory
	routerPath := filepath.Join(untarFlist, "router.yaml")
	if err := ioutil.WriteFile(routerPath, router, 0660); err != nil {
		return err
	}

	// re packages the flist directory with the new router.yaml
	output, err := os.OpenFile(flist, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0660)
	if err != nil {
		log.Errorf("fail to open flist (%s) for writing: %v", flist, err)
		return err
	}
	defer output.Close()

	if err := Tar(untarFlist, output); err != nil {
		log.Errorf("fail to create the tar archive: %v", err)
		return err
	}

	return nil
}

// Untar takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing any files
func Untar(dst string, r io.Reader) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	gzr, err := gzip.NewReader(r)
	if err != nil {
		log.Error("fail to create gzip reader : %v", err)
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			log.Error("error in loop")
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					log.Error("error mkdir")
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				log.Error("open file")
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				log.Error("error copy")
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
}

// Tar takes a source and variable writers and walks 'source' writing each file
// found to the tar writer
func Tar(src string, w io.Writer) error {

	// ensure the src actually exists before trying to tar it
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("Unable to tar files - %v", err.Error())
	}

	gzw := gzip.NewWriter(w)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	var baseDir string
	if sourceInfo.IsDir() {
		baseDir = filepath.Base(src)
	}

	// walk path
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {

		// return on any error
		if err != nil {
			return err
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		if info.IsDir() && path == src {
			header.Name = "."
		} else if baseDir != "" {
			header.Name = "." + strings.TrimPrefix(path, src)
		}

		if info.IsDir() {
			header.Name += "/"
		}

		// write the header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if header.Typeflag == tar.TypeReg {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("%s: open: %v", path, err)
			}
			defer file.Close()

			_, err = io.CopyN(tw, file, info.Size())
			if err != nil && err != io.EOF {
				return fmt.Errorf("%s: copying contents: %v", path, err)
			}
		}
		return nil
	})
}
