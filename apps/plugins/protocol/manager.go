package protocol

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/threefoldtech/0-core/base/plugin"

	"github.com/garyburd/redigo/redis"
	"github.com/threefoldtech/0-core/base/pm"
)

const (
	SinkQueue = "core:default"
	DBIndex   = 0

	//ReturnExpire in 300 seconds (5min)
	ReturnExpire = 300
)

type Manager struct {
	api  plugin.API
	pool *redis.Pool
}

//Result handler implementation
func (m *Manager) Result(cmd *pm.Command, result *pm.JobResult) {
	if err := m.Set(result); err != nil {
		log.Debugf("failed to forward result: %s", cmd.ID)
	}
}

func (m *Manager) Database() Database {
	return m
}

func (m *Manager) process() {
	for {
		var command pm.Command
		err := m.next(SinkQueue, &command)
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
		if m.Flagged(command.ID) {
			log.Errorf("received a command with a duplicate ID(%v), dropping", command.ID)
			continue
		}

		m.Flag(command.ID)
		log.Debugf("Starting command %s", &command)

		_, err = m.api.Run(&command)

		if pm.IsUnknownCommand(err) {
			result := pm.NewJobResult(&command)
			result.State = pm.StateUnknownCmd
			m.Set(result)
		} else if err != nil {
			log.Errorf("Unknown error while processing command (%s): %s", command, err)
		}
	}
}

//Set forwards job result
func (m *Manager) Set(result *pm.JobResult) error {
	m.UnFlag(result.ID)
	return m.setResult(result)
}

//Get gets a result of a job if it exists
func (m *Manager) Get(job string, timeout int) (*pm.JobResult, error) {
	if m.Flagged(job) {
		return m.getResult(job, timeout)
	}

	return nil, fmt.Errorf("unknown job id '%s' (may be it has expired)", job)
}

//Flag mark job as processing
func (m *Manager) Flag(id string) error {
	conn := m.pool.Get()
	defer conn.Close()

	key := fmt.Sprintf("result:%s:flag", id)
	_, err := conn.Do("RPUSH", key, "")
	return err
}

//UnFlag mark job as done
func (m *Manager) UnFlag(id string) error {
	conn := m.pool.Get()
	defer conn.Close()

	key := fmt.Sprintf("result:%s:flag", id)
	_, err := conn.Do("EXPIRE", key, ReturnExpire)
	return err
}

//Flagged check if job is marked
func (m *Manager) Flagged(id string) bool {
	conn := m.pool.Get()
	defer conn.Close()

	key := fmt.Sprintf("result:%s:flag", id)
	v, _ := redis.Int(conn.Do("EXISTS", key))
	return v == 1
}

//next gets the next available command
func (m *Manager) next(queue string, command *pm.Command) error {
	conn := m.pool.Get()
	defer conn.Close()

	payload, err := redis.ByteSlices(conn.Do("BLPOP", queue, 10))
	if err != nil {
		return err
	}

	if payload == nil || len(payload) < 2 {
		return redis.ErrNil
	}

	return json.Unmarshal(payload[1], command)
}

func (m *Manager) setResult(result *pm.JobResult) error {
	if result.ID == "" {
		return fmt.Errorf("result with no ID, not pushing results back")
	}

	queue := fmt.Sprintf("result:%s", result.ID)

	conn := m.pool.Get()
	defer conn.Close()

	if err := m.push(conn, queue, result); err != nil {
		return err
	}

	if _, err := conn.Do("EXPIRE", queue, ReturnExpire); err != nil {
		return err
	}

	return nil
}

func (m *Manager) push(conn redis.Conn, queue string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if _, err := conn.Do("RPUSH", queue, data); err != nil {
		return err
	}

	return nil
}

func (m *Manager) cycle(queue string, timeout int) ([]byte, error) {
	conn := m.pool.Get()
	defer conn.Close()

	return redis.Bytes(conn.Do("BRPOPLPUSH", queue, queue, timeout))
}

func (m *Manager) getResult(id string, timeout int) (*pm.JobResult, error) {
	queue := fmt.Sprintf("result:%s", id)
	payload, err := m.cycle(queue, timeout)
	if err != nil {
		return nil, err
	}

	var result pm.JobResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
