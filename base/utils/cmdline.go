package utils

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/google/shlex"
)

type KernelOptions map[string][]string

func (k KernelOptions) Is(key string) bool {
	_, ok := k[key]
	return ok
}

func (k KernelOptions) Get(key string) ([]string, bool) {
	v, ok := k[key]
	return v, ok
}

func (k KernelOptions) GetLast() map[string]interface{} {
	r := make(map[string]interface{})
	for key, values := range k {
		if len(values) > 0 {
			r[key] = values[len(values)-1]
		} else {
			r[key] = ""
		}
	}

	return r
}

func (k KernelOptions) String(keys ...string) string {
	var s []string
	for _, key := range keys {
		values, ok := k[key]
		if !ok {
			continue
		}

		if len(values) == 0 {
			s = append(s, key)
			continue
		}

		for _, v := range values {
			if len(v) != 0 {
				s = append(s, fmt.Sprintf("%s=%s", key, v))
			} else {
				s = append(s, key)
			}
		}
	}

	return strings.Join(s, ", ")
}

//Cmdline return kerenl arguments cmdline string making sure
//arguments exists only once on the command line
func (k KernelOptions) Cmdline() []byte {
	x := make(map[string]struct{})
	for k, vs := range k {
		if len(vs) == 0 {
			x[k] = struct{}{}
			continue
		}

		for _, v := range vs {
			if strings.ContainsAny(v, " \t") {
				x[fmt.Sprintf("%s=\"%s\"", k, v)] = struct{}{}
			} else {
				x[fmt.Sprintf("%s=%s", k, v)] = struct{}{}
			}
		}
	}

	var buf strings.Builder
	for o := range x {
		if buf.Len() > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(o)
	}

	return []byte(buf.String())
}

func parseKerenlOptions(content string) KernelOptions {
	options := KernelOptions{}
	cmdline, _ := shlex.Split(strings.TrimSpace(content))
	for _, option := range cmdline {
		kv := strings.SplitN(option, "=", 2)
		key := kv[0]

		if len(kv) == 2 {
			options[key] = append(options[key], kv[1])
		} else {
			options[key] = nil
		}
	}

	return options
}

//GetKernelOptions Get kernel cmdline arguments
func GetKernelOptions() KernelOptions {
	content, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		log.Warning("Failed to read /proc/cmdline", err)
		return KernelOptions{}
	}

	return parseKernelOptions(string(content))
}
