package main

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/threefoldtech/0-core/apps/plugins/nft"
)

const (
	filterInput = `
	{
		"nftables": [
		  {
			"table": {
			  "family": "ip",
			  "name": "nat",
			  "handle": 0
			}
		  },
		  {
			"set": {
			  "family": "ip",
			  "table": "nat",
			  "name": "host",
			  "elem": [
				"10.20.1.1",
				"172.18.0.1",
				"172.19.0.1"
			  ],
			  "type": "ipv4_addr",
			  "handle": 0
			}
		  },
		  {
			"chain": {
			  "family": "ip",
			  "name": "pre",
			  "table": "nat",
			  "handle": 1,
			  "type": "nat",
			  "prio": 0,
			  "hook": "prerouting",
			  "policy": "accept"
			}
		  },
		  {
			"rule": {
			  "expr": [
				{
				  "match": {
					"left": {
					  "payload": {
						"name": "ip",
						"field": "daddr"
					  }
					},
					"right": "@host"
				  }
				},
				{
				  "match": {
					"left": {
					  "payload": {
						"name": "tcp",
						"field": "dport"
					  }
					},
					"right": 8000
				  }
				},
				{
				  "mangle": {
					"left": {
					  "meta": "mark"
					},
					"right": 123
				  }
				},
				{
				  "dnat": {
					"addr": "172.18.0.100"
				  }
				}
			  ],
			  "family": "ip",
			  "table": "nat",
			  "chain": "pre",
			  "handle": 7
			}
		  },
		  {
			"rule": {
			  "expr": [
				{
				  "match": {
					"left": {
					  "payload": {
						"name": "ip",
						"field": "daddr"
					  }
					},
					"right": "@host"
				  }
				},
				{
				  "match": {
					"left": {
					  "payload": {
						"name": "tcp",
						"field": "dport"
					  }
					},
					"right": 8001
				  }
				},
				{
				  "mangle": {
					"left": {
					  "meta": "mark"
					},
					"right": 124
				  }
				},
				{
				  "dnat": {
					"addr": "172.18.0.100",
					"port": 7000
				  }
				}
			  ],
			  "family": "ip",
			  "table": "nat",
			  "chain": "pre",
			  "handle": 8
			}
		  },
		  {
			"chain": {
			  "family": "ip",
			  "name": "post",
			  "table": "nat",
			  "handle": 2,
			  "type": "nat",
			  "prio": 0,
			  "hook": "postrouting",
			  "policy": "accept"
			}
		  },
		  {
			"rule": {
			  "expr": [
				{
				  "match": {
					"left": {
					  "payload": {
						"name": "ip",
						"field": "saddr"
					  }
					},
					"right": {
					  "prefix": {
						"addr": "172.18.0.0",
						"len": 16
					  }
					}
				  }
				},
				{
				  "masquerade": null
				}
			  ],
			  "family": "ip",
			  "table": "nat",
			  "chain": "post",
			  "handle": 4
			}
		  },
		  {
			"rule": {
			  "expr": [
				{
				  "match": {
					"left": {
					  "payload": {
						"name": "ip",
						"field": "saddr"
					  }
					},
					"right": {
					  "prefix": {
						"addr": "172.19.0.0",
						"len": 16
					  }
					}
				  }
				},
				{
				  "masquerade": null
				}
			  ],
			  "family": "ip",
			  "table": "nat",
			  "chain": "post",
			  "handle": 5
			}
		  },
		  {
			"table": {
			  "family": "inet",
			  "name": "filter",
			  "handle": 0
			}
		  },
		  {
			"chain": {
			  "family": "inet",
			  "name": "output",
			  "table": "filter",
			  "handle": 1,
			  "type": "filter",
			  "prio": 0,
			  "hook": "output",
			  "policy": "accept"
			}
		  },
		  {
			"chain": {
			  "family": "inet",
			  "name": "input",
			  "table": "filter",
			  "handle": 2,
			  "type": "filter",
			  "prio": 0,
			  "hook": "input",
			  "policy": "drop"
			}
		  },
		  {
			"rule": {
			  "expr": [
				{
				  "match": {
					"left": {
					  "ct": {
						"key": "state"
					  }
					},
					"right": {
					  "set": [
						"established",
						"related"
					  ]
					}
				  }
				},
				{
				  "accept": null
				}
			  ],
			  "family": "inet",
			  "table": "filter",
			  "chain": "input",
			  "handle": 5
			}
		  },
		  {
			"rule": {
			  "expr": [
				{
				  "match": {
					"left": {
					  "meta": "iifname"
					},
					"right": "lo"
				  }
				},
				{
				  "accept": null
				}
			  ],
			  "family": "inet",
			  "table": "filter",
			  "chain": "input",
			  "handle": 6
			}
		  },
		  {
			"rule": {
			  "expr": [
				{
				  "match": {
					"left": {
					  "meta": "iifname"
					},
					"right": "vxbackend"
				  }
				},
				{
				  "accept": null
				}
			  ],
			  "family": "inet",
			  "table": "filter",
			  "chain": "input",
			  "handle": 7
			}
		  },
		  {
			"rule": {
			  "expr": [
				{
				  "match": {
					"left": {
					  "payload": {
						"name": "ip",
						"field": "protocol"
					  }
					},
					"right": 1
				  }
				},
				{
				  "accept": null
				}
			  ],
			  "family": "inet",
			  "table": "filter",
			  "chain": "input",
			  "handle": 8
			}
		  },
		  {
			"rule": {
			  "expr": [
				{
				  "match": {
					"left": {
					  "meta": "iif"
					},
					"right": "core0"
				  }
				},
				{
				  "match": {
					"left": {
					  "payload": {
						"name": "udp",
						"field": "dport"
					  }
					},
					"right": {
					  "set": [
						53,
						67,
						68
					  ]
					}
				  }
				},
				{
				  "accept": null
				}
			  ],
			  "family": "inet",
			  "table": "filter",
			  "chain": "input",
			  "handle": 9
			}
		  },
		  {
			"rule": {
			  "expr": [
				{
				  "match": {
					"left": {
					  "meta": "iif"
					},
					"right": "kvm0"
				  }
				},
				{
				  "match": {
					"left": {
					  "payload": {
						"name": "udp",
						"field": "dport"
					  }
					},
					"right": {
					  "set": [
						53,
						67,
						68
					  ]
					}
				  }
				},
				{
				  "accept": null
				}
			  ],
			  "family": "inet",
			  "table": "filter",
			  "chain": "input",
			  "handle": 10
			}
		  },
		  {
			"rule": {
			  "expr": [
				{
				  "match": {
					"left": {
					  "payload": {
						"name": "tcp",
						"field": "dport"
					  }
					},
					"right": 6379
				  }
				},
				{
				  "accept": null
				}
			  ],
			  "family": "inet",
			  "table": "filter",
			  "chain": "input",
			  "handle": 11
			}
		  },
		  {
			"rule": {
			  "expr": [
				{
				  "match": {
					"left": {
					  "payload": {
						"name": "tcp",
						"field": "dport"
					  }
					},
					"right": 22
				  }
				},
				{
				  "accept": null
				}
			  ],
			  "family": "inet",
			  "table": "filter",
			  "chain": "input",
			  "handle": 12
			}
		  },
		  {
			"chain": {
			  "family": "inet",
			  "name": "pre",
			  "table": "filter",
			  "handle": 3,
			  "type": "filter",
			  "prio": 0,
			  "hook": "prerouting",
			  "policy": "accept"
			}
		  },
		  {
			"chain": {
			  "family": "inet",
			  "name": "forward",
			  "table": "filter",
			  "handle": 4,
			  "type": "filter",
			  "prio": 0,
			  "hook": "forward",
			  "policy": "accept"
			}
		  }
		]
	  }
	`
)

func TestFilter(t *testing.T) {
	rules, err := Filter(filterInput, &nft.IntMatchFilter{
		Name:  "tcp",
		Field: "dport",
		Value: 8000,
	})

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Len(t, rules, 1); !ok {
		t.Error()
	}

	rules, err = Filter(filterInput, &nft.MarkFilter{
		Mark: 123,
	})

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Len(t, rules, 1); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, 7, rules[0].Handle); !ok {
		t.Error()
	}

	//test OR
	rules, err = Filter(filterInput, &nft.MarkFilter{
		Mark: 123,
	}, &nft.IntMatchFilter{
		Name:  "tcp",
		Field: "dport",
		Value: 8001,
	})

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Len(t, rules, 2); !ok {
		t.Error()
	}

	//test And
	rules, err = Filter(
		filterInput,
		nft.And{&nft.MarkFilter{
			Mark: 123,
		}, &nft.IntMatchFilter{
			Name:  "tcp",
			Field: "dport",
			Value: 8000,
		}})

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Len(t, rules, 1); !ok {
		t.Error()
	}
}

func TestFilterNetwork(t *testing.T) {
	_, value, _ := net.ParseCIDR("172.19.0.0/16")
	rules, err := Filter(filterInput, &nft.NetworkMatchFilter{
		Name:  "ip",
		Field: "saddr",
		Value: value,
	})

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Len(t, rules, 1); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, 5, rules[0].Handle); !ok {
		t.Error()
	}
}

func TestFilterMeta(t *testing.T) {
	rules, err := Filter(filterInput, &nft.MetaMatchFilter{
		Name:  "iif",
		Value: "core0",
	})

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Len(t, rules, 1); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, 9, rules[0].Handle); !ok {
		t.Error()
	}
}
