package main

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"syscall"
)

const (
	//HubHost host name of hub
	HubHost = "hub.grid.tf"

	//ValidFileName
	ValidFileName = ".valid"
)

type FList interface {
	Hash() (string, error)
	Open() (io.ReadCloser, error)
	Ext() string
}

//LocalFList represents a local flist
type LocalFList string

func (l LocalFList) Hash() (string, error) {
	_, err := os.Stat(string(l))
	if err != nil {
		return "", err
	}

	m := md5.New()
	io.WriteString(m, string(l))
	return fmt.Sprintf("%x", m.Sum(nil)), nil
}

func (l LocalFList) Open() (io.ReadCloser, error) {
	return os.Open(string(l))
}

func (l LocalFList) Ext() string {
	return path.Ext(string(l))
}

type RemoteFList string

func (l RemoteFList) Hash() (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s.md5", string(l)))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get flist hash: %s", resp.Status)
	}

	return string(data[:32]), nil
}

func (l RemoteFList) Open() (io.ReadCloser, error) {
	response, err := http.Get(string(l))
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download flist: %s", response.Status)
	}

	return response.Body, nil
}

func (l RemoteFList) Ext() string {
	return path.Ext(string(l))
}

type Meta struct {
	Base string
	Hash string
}

func getFList(src string) (FList, error) {
	u, err := url.Parse(src)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "file" || u.Scheme == "" {
		return LocalFList(u.Path), nil
	} else if u.Scheme == "http" || u.Scheme == "https" {
		if u.Hostname() != HubHost {
			return nil, fmt.Errorf("remote flist must be from (%s)", HubHost)
		}
		return RemoteFList(src), nil
	}

	return nil, fmt.Errorf("invalid flist url (%s)", src)
}

func metaValidate(base string, f *os.File) (bool, error) {
	stat, err := f.Stat()
	if err != nil {
		return false, err
	}

	if stat.Size() == 0 {
		//meta was not downloaded, so invalid meta
		return false, nil
	}

	var data map[string]int64
	dec := json.NewDecoder(f)
	if err := dec.Decode(&data); err != nil {
		//corrupt validation file, redownload
		return false, nil
	}

	if len(data) == 0 {
		//no files in validate file, redownload anyway
		return false, nil
	}

	files, err := ioutil.ReadDir(base)
	if err != nil {
		return false, err
	}

	for _, file := range files {
		if file.IsDir() || file.Name() == ValidFileName {
			continue
		}
		size, ok := data[file.Name()]
		if !ok || file.Size() != size {
			//no entry, or invalid size
			return false, nil
		}
	}

	return true, nil
}

func extract(base string, flist FList, v *os.File) error {
	src, err := flist.Open()
	if err != nil {
		return err
	}

	defer src.Close()
	var reader io.Reader

	switch flist.Ext() {
	case ".gz", ".tgz", ".flist":
		var zipReader *gzip.Reader
		zipReader, err = gzip.NewReader(src)
		if err != nil {
			return err
		}
		defer zipReader.Close()
		reader = zipReader
	case ".tbz2", ".bz2":
		reader = bzip2.NewReader(src)
	case ".tar":
		//use same reader
	default:
		return fmt.Errorf("unknown flist type: %s", flist.Ext())
	}

	archive := tar.NewReader(reader)
	valid := make(map[string]int64)

	for {
		header, err := archive.Next()
		if err != nil && err != io.EOF {
			return err
		} else if err == io.EOF {
			break
		}

		if header.FileInfo().IsDir() {
			continue
		}

		if err := os.MkdirAll(path.Join(base, path.Dir(header.Name)), 0755); err != nil {
			return err
		}

		file, err := os.Create(path.Join(base, header.Name))
		if err != nil {
			return err
		}

		if _, err := io.Copy(file, archive); err != nil {
			file.Close()
			return err
		}
		size, _ := file.Seek(0, 2)
		valid[path.Base(header.Name)] = size
		file.Close()
	}

	v.Truncate(0)
	enc := json.NewEncoder(v)
	enc.SetIndent("", "  ")
	return enc.Encode(valid)
}

func getMeta(src string) (*Meta, error) {
	flist, err := getFList(src)
	if err != nil {
		return nil, err
	}

	hash, err := flist.Hash()
	if err != nil {
		return nil, err
	}

	base := path.Join(CacheFListDir, hash)
	if err := os.MkdirAll(base, 0755); err != nil {
		return nil, err
	}

	vfileName := path.Join(base, ValidFileName)
	vfile, err := os.OpenFile(vfileName, os.O_CREATE|os.O_RDWR, os.ModePerm&os.FileMode(0755))
	if err != nil {
		return nil, err
	}
	defer vfile.Close()
	if err := syscall.Flock(int(vfile.Fd()), syscall.LOCK_EX); err != nil {
		return nil, err
	}

	defer syscall.Flock(int(vfile.Fd()), syscall.LOCK_UN)

	valid, err := metaValidate(base, vfile)
	if err != nil {
		return nil, err
	}
	meta := &Meta{
		Base: base,
		Hash: hash,
	}

	if valid {
		//downloaded meta is valid
		log.Debugf("flist %s is valid, no redownload", src)
		return meta, nil
	}

	log.Debugf("downloading flist: %s", src)

	return meta, extract(base, flist, vfile)
}
