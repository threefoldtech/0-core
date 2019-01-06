package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/pborman/uuid"
)

const (
	CommandsQueue = "core:default"

	ResultNoTimeout      = 0
	ResultDefaultTimeout = 10
)

var (
	schemaRegex = regexp.MustCompile(`^(unix|tcp)(\+ssl)?`)
)

type redisClient struct {
	pool *redis.Pool

	info InfoManager
}

func NewClientWithPool(pool *redis.Pool) Client {
	cl := &redisClient{pool: pool}

	cl.info = &infoMgr{cl}
	return cl
}

type options struct {
	Network string
	SSL     bool
	Address string
}

func parseAddress(address string) (options, error) {
	u, err := url.Parse(address)
	if err != nil {
		return options{}, err
	}
	scheme := u.Scheme
	if len(scheme) == 0 {
		scheme = "tcp"
	}

	parsed := schemaRegex.FindStringSubmatch(scheme)
	if len(parsed) == 0 {
		return options{}, fmt.Errorf("invalid address scheme '%s'", scheme)
	}
	switch parsed[1] {
	case "tcp":
		address = u.Host
	case "unix":
		address = u.Path
	}

	return options{
		Network: parsed[1],
		SSL:     parsed[2] == "+ssl",
		Address: address,
	}, nil
}

func NewPool(address, password string) (*redis.Pool, error) {
	o, err := parseAddress(address)
	if err != nil {
		return nil, err
	}

	var opts []redis.DialOption
	if o.SSL {
		opts = append(opts,
			redis.DialNetDial(func(network, address string) (net.Conn, error) {
				return tls.Dial(network, address, &tls.Config{
					InsecureSkipVerify: true,
				})
			}),
		)
	}

	return &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			// the redis protocol should probably be made sett-able
			c, err := redis.Dial(o.Network, o.Address, opts...)

			if err != nil {
				return nil, err
			}

			if len(password) > 0 {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			} else {
				// check with PING
				if _, err := c.Do("PING"); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		// custom connection test method
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if _, err := c.Do("PING"); err != nil {
				return err
			}
			return nil
		},
	}, nil
}

func NewClient(address, password string) (Client, error) {
	pool, err := NewPool(address, password)
	if err != nil {
		return nil, err
	}
	return NewClientWithPool(pool), nil
}

func (c *redisClient) Raw(name string, args A, opts ...Option) (JobId, error) {
	cmd := &Command{
		Command:   name,
		Arguments: args,
	}

	for _, opt := range opts {
		opt.apply(cmd)
	}

	if cmd.ID == "" {
		cmd.ID = uuid.New()
	}

	db := c.pool.Get()
	defer db.Close()

	data, err := json.Marshal(cmd)
	if err != nil {
		return JobId(""), err
	}

	if _, err := db.Do("RPUSH", CommandsQueue, string(data)); err != nil {
		return JobId(""), err
	}

	flag := fmt.Sprintf("result:%v:flag", cmd.ID)
	if _, err := db.Do("BRPOPLPUSH", flag, flag, ResultDefaultTimeout); err != nil {
		return JobId(cmd.ID), fmt.Errorf("failed to queue command '%v'", cmd.ID)
	}

	return JobId(cmd.ID), nil
}

func (c *redisClient) Exists(job JobId) bool {
	db := c.pool.Get()
	defer db.Close()

	flag := fmt.Sprintf("result:%v:flag", job)

	res, err := db.Do("RPOPLPUSH", flag, flag)
	return err == nil && res != nil
}

func (c *redisClient) result(job JobId, timeout ...int) (*Result, error) {
	if !c.Exists(job) {
		return nil, fmt.Errorf("job '%v' does not exist", job)
	}
	db := c.pool.Get()
	defer db.Close()

	var reply []byte
	var err error
	q := fmt.Sprintf("result:%v", job)

	if len(timeout) > 0 {
		reply, err = redis.Bytes(db.Do("BRPOPLPUSH", q, q, timeout[0]))
	} else {
		reply, err = redis.Bytes(db.Do("RPOPLPUSH", q, q))
	}

	if err != nil {
		return nil, err
	}

	var r Result
	if err := json.Unmarshal(reply, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

func (c *redisClient) Result(job JobId, timeout ...int) (*Result, error) {
	to := ResultDefaultTimeout
	if len(timeout) > 0 {
		to = timeout[0]
	}

	return c.result(job, to)
}

func (c *redisClient) ResultNonBlock(job JobId) (*Result, error) {
	return c.result(job)
}

func (c *redisClient) Info() InfoManager {
	return c.info
}
