package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/threefoldtech/0-core/base/pm"
	client "github.com/threefoldtech/0-core/client/go-client"
	"gopkg.in/yaml.v2"
)

var (
	outputs = []io.Writer{
		os.Stdout,
		os.Stderr,
	}
)

type M map[string]interface{}

type Response pm.JobResult

func (r *Response) Print() {
	data, err := yaml.Marshal(r)
	if err != nil {
		log.Fatalf("failed to format results: %s", err)
	}

	fmt.Println(string(data))
}

func PrintStreams(r *client.Result) {
	for i, s := range r.Streams {
		if len(s) > 0 {
			fmt.Fprintf(outputs[i], s)
		}
	}
}

func PrintResultOrDie(r *client.Result, err error) {
	if err != nil {
		log.Fatal(err)
	}

	PrintStreams(r)

	if r.State != client.StateSuccess {
		log.Errorf("State: %s", r.State)
		if r.Data != "" {
			log.Errorf("%s", r.Data)
		}

		if r.Critical != "" {
			log.Errorf("%s", r.Critical)
		}

		os.Exit(1)
	}
}

func PrintOrDie(obj interface{}, err error) {
	if err != nil {
		log.Fatal(err)
	}

	if out, err := yaml.Marshal(obj); err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(string(out))
	}
}

func (r *Response) PrintYaml() {
	var output interface{}
	if err := json.Unmarshal([]byte(r.Data), &output); err != nil {
		log.Fatal(err)
	}

	if out, err := yaml.Marshal(output); err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(string(out))
	}
}
