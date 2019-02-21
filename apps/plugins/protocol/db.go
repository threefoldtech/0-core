package protocol

import "github.com/garyburd/redigo/redis"

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

//GetKey gets value from key
func (m *Manager) GetKey(key string) ([]byte, error) {
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

//SetKey sets a value to a key
func (m *Manager) SetKey(key string, value []byte) error {
	conn := m.pool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", key, value)
	return err
}

//DelKey delets keys
func (m *Manager) DelKeys(keys ...string) (int64, error) {
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
