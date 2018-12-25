package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	cache "github.com/patrickmn/go-cache"
	"github.com/threefoldtech/0-core/base/pm"
)

const (
	StatisticsQueueKey = "statistics:%d"
	StateKey           = "state:%s:%s"
	KeyIdSep           = ":"
	IDTag              = "id"
)

var (
	Periods = []int64{300, 3600} //5 min, 1 hour
)

/*
StatsBuffer implements a buffering and flushing mechanism to buffer statsd messages
that are collected via the process manager. Flush happens when buffer is full or a certain time passes since last flush.

The StatsBuffer.Handler should be registers as StatsFlushHandler on the process manager object.
*/

type Tags []pm.Tag

func (t Tags) Len() int {
	return len(t)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (t Tags) Less(i, j int) bool {
	return t[i].Key < t[j].Key
}

// Swap swaps the elements with indexes i and j.
func (t Tags) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

type Stats struct {
	Operation Operation `json:"operation"`
	Key       string    `json:"key"`
	Value     float64   `json:"value"`
	Tags      []pm.Tag  `json:"tags"`
}

type Point struct {
	*Sample
	Key  string            `json:"key"`
	Tags map[string]string `json:"tags,omitempty"`
}

func (r *Manager) query(ctx pm.Context) (interface{}, error) {
	var filter struct {
		Key  string            `json:"key"`
		Tags map[string]string `json:"tags"`
	}

	cmd := ctx.Command()

	if err := json.Unmarshal(*cmd.Arguments, &filter); err != nil {
		return nil, err
	}

	result := make(map[string]*State)

	for key := range r.cache.Items() {
		parts := strings.SplitN(key, KeyIdSep, 3) //formated as `StateKey`
		metric := parts[1]
		if len(filter.Key) != 0 {
			if filter.Key != metric {
				continue
			}
		}

		data, err := r.protocol.Database().GetKey(key)
		if err != nil {
			log.Errorf("failed to get state for metric: %s", key)
			continue
		}

		state, err := LoadState(data)
		if err != nil {
			log.Errorf("failed to load stat for %s", key)
		}

		//filter on tags
		m := true
		for k, v := range filter.Tags {
			m = false
			for _, t := range state.Tags {
				if t.Key == k && t.Value == v {
					m = true
					break
				}
			}
			if !m {
				break
			}
		}

		if !m {
			continue
		}

		//get ID if set
		for _, t := range state.Tags {
			if t.Key == IDTag {
				metric = fmt.Sprintf("%s/%s", metric, t.Value)
				break
			}
		}

		result[metric] = state
	}

	return result, nil
}

func (r *Manager) hash(tags []pm.Tag) string {
	sort.Sort(Tags(tags))
	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%v", tags))))
}

//Stats is a states handler implementation
func (r *Manager) Stats(op string, key string, value float64, id string, tags ...pm.Tag) {
	if len(id) != 0 {
		tags = append(tags, pm.Tag{IDTag, id})
	}

	hash := r.hash(tags)
	internal := fmt.Sprintf(StateKey, key, hash)

	//touch key in cache so we know we are tracking this key
	r.cache.Set(internal, nil, cache.DefaultExpiration)

	data, err := r.protocol.Database().GetKey(internal)
	if err != nil {
		log.Errorf("failed to get value for %s: %s", key, err)
		return
	}

	var state *State
	if data == nil {
		state = NewState(Operation(op), Periods...)
	} else if state, err = LoadState(data); err != nil {
		log.Errorf("failed to load state object for %s: %s", key, err)
		return
	}

	if len(tags) != 0 {
		state.Tags = tags
	}

	for period, sample := range state.Feed(value) {
		if sample.Start == 0 {
			//undefined sample
			continue
		}

		queue := fmt.Sprintf(StatisticsQueueKey, period)
		p := Point{
			Sample: sample,
			Key:    key,
			Tags:   make(map[string]string),
		}

		for _, tag := range state.Tags {
			p.Tags[tag.Key] = tag.Value
		}

		if data, err := json.Marshal(&p); err == nil {
			r.protocol.Database().RPush(queue, data)
		} else {
			log.Errorf("statistics point marshal error: %s", err)
		}
	}

	data, err = json.Marshal(state)
	if err != nil {
		log.Errorf("failed to marshal state object for %s: %s", key, err)
		return
	}

	if err := r.protocol.Database().SetKey(internal, data); err != nil {
		log.Errorf("failed to save state object for %s: %s", key, err)
	}
}
