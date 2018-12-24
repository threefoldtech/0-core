package main

import (
	"fmt"
	"time"

	"github.com/threefoldtech/0-core/base/plugin"

	"github.com/garyburd/redigo/redis"
	"github.com/threefoldtech/0-core/base/pm"
)

const (
	SinkQueue = "core:default"
	DBIndex   = 0
)

type Manager struct {
	api  plugin.API
	db   *Database
	pool *redis.Pool
}

//RPush pushes values to the right
func (m *Manager) RPush(key string, args ...[]byte) (int64, error) {
	conn := m.pool.Get()
	defer conn.Close()
	input := make([]interface{}, 0, len(args)+1)
	input = append(input, key)
	for _, arg := range args {
		input = append(input, arg)
	}

	return redis.Int64(conn.Do("RPUSH", input...))
}

//LTrim trims a list
func (m *Manager) LTrim(key string, start, stop int64) error {
	conn := m.pool.Get()
	defer conn.Close()

	_, err := conn.Do("LTRIM", key, start, stop)
	return err
}

//Get gets value from key
func (m *Manager) Get(key string) ([]byte, error) {
	conn := m.pool.Get()
	defer conn.Close()
	result, err := redis.Bytes(conn.Do("GET", key))
	//we do the next, because this is how ledis used
	//to behave
	if err == redis.ErrNil {
		return nil, nil
	}
	return result, err
}

//Set sets a value to a key
func (m *Manager) Set(key string, value []byte) error {
	conn := m.pool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", key, value)
	return err
}

//Del delets keys
func (m *Manager) Del(keys ...string) (int64, error) {
	conn := m.pool.Get()
	defer conn.Close()
	input := make([]interface{}, 0, len(keys))
	for _, key := range keys {
		input = append(input, key)
	}

	return redis.Int64(conn.Do("DEL", input...))
}

//LExpire sets TTL on a list
func (m *Manager) LExpire(key string, duration int64) (int64, error) {
	conn := m.pool.Get()
	defer conn.Close()
	return redis.Int64(conn.Do("EXPIRE", key, duration))
}

//Result handler implementation
func (m *Manager) Result(cmd *pm.Command, result *pm.JobResult) {
	if err := m.Forward(result); err != nil {
		log.Debugf("failed to forward result: %s", cmd.ID)
	}
}

func (m *Manager) process() {
	for {
		var command pm.Command
		err := m.db.GetNext(SinkQueue, &command)
		if err == redis.ErrNil {
			continue
		} else if err != nil {
			log.Errorf("Failed to get next command from (%s): %s", SinkQueue, err)
			<-time.After(200 * time.Millisecond)
			continue
		}

		if command.ID == "" {
			log.Warningf("receiving a command with no ID, dropping")
			continue
		}
		if m.db.Flagged(command.ID) {
			log.Errorf("received a command with a duplicate ID(%v), dropping", command.ID)
			continue
		}

		m.db.Flag(command.ID)
		log.Debugf("Starting command %s", &command)

		_, err = m.api.Run(&command)

		if err == pm.UnknownCommandErr {
			result := pm.NewJobResult(&command)
			result.State = pm.StateUnknownCmd
			m.Forward(result)
		} else if err != nil {
			log.Errorf("Unknown error while processing command (%s): %s", command, err)
		}
	}
}

//Forward forwards job result
func (m *Manager) Forward(result *pm.JobResult) error {
	m.db.UnFlag(result.ID)
	return m.db.Respond(result)
}

//Flag marks a job ID as running
func (m *Manager) Flag(id string) error {
	return m.db.Flag(id)
}

//GetResult gets a result of a job if it exists
func (m *Manager) GetResult(job string, timeout int) (*pm.JobResult, error) {
	if m.db.Flagged(job) {
		return m.db.GetResponse(job, timeout)
	}

	return nil, fmt.Errorf("unknown job id '%s' (may be it has expired)", job)
}
