package containers

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestTar(t *testing.T) {
	path := createTar(t)
	defer func() {
		os.Remove(path)
	}()

	router := router{
		Pools: map[string]map[string]string{
			"private": map[string]string{"00:FF": "zdb://172.0.0.1:990"},
		},
		Lookup: []string{"private"},
	}
	routerb, err := yaml.Marshal(router)
	require.NoError(t, err)

	err = addRouterFile(path, routerb)
	require.NoError(t, err)

	f, err := os.Open(path)
	require.NoError(t, err)

	gzr, err := gzip.NewReader(f)
	require.NoError(t, err)

	tr := tar.NewReader(gzr)
	filesInTar := make([]string, 0, 2)
	for {
		header, err := tr.Next()

		// if no more files are found return
		if err == io.EOF {
			break
		}

		// return any other error
		if err != nil {
			t.Fatal(err)
		}

		// if the header is nil, just skip it (not sure how this happens)
		if header == nil {
			continue
		}

		if header.Typeflag == tar.TypeReg {
			filesInTar = append(filesInTar, header.Name)
		}
	}

	assert.Equal(t, []string{"./flistdb.sqlite3", "./router.yaml"}, filesInTar)
}

func createTar(t *testing.T) string {
	f, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer f.Close()

	gzw := gzip.NewWriter(f)
	defer gzw.Close()
	tw := tar.NewWriter(gzw)
	defer tw.Close()

	var files = []struct {
		Name, Body string
	}{
		{"./flistdb.sqlite3", "foo"},
	}
	for _, file := range files {
		hdr := &tar.Header{
			Typeflag: tar.TypeReg,
			Name:     file.Name,
			Size:     int64(len(file.Body)),
			Mode:     493,
		}
		tw.WriteHeader(hdr)
		require.NoError(t, err)

		if _, err := tw.Write([]byte(file.Body)); err != nil {

		}
	}
	return f.Name()
}
