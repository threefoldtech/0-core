# Flists

This file contains link to some example flist you can use with the core0.

- Ubuntu 16.04 : https://stor.jumpscale.org/public/flists/flist-ubuntu1604.db.tar.gz

## How to create a flist:

To create a flist you need to have [Jumpscale](https://github.com/Jumpscale/jumpscale_core8#how-to-install-from-master)  

Creation of the flist db:
```python
# open a connection to a rocksdb database
kvs = j.servers.kvs.getRocksDBStore(name='flist', namespace=None, dbpath="/tmp/flist-example.db")
# create a flist object and pass the reference of the rocksdb to it.
f = j.tools.flist.getFlist(rootpath='/', kvs=kvs)
# add the path you want to include in your flist, can be multiple call to f.add
f.add('/opt')
# upload your flist to an ardb data store.
f.upload("remote-ardb-server", 16379)
```

Package you flist db:
```shell
cd /tmp/flist-example.db
tar -cf ../flist-example.db.tar *
cd .. && gzip flist-example.db.tar
```

The result is `flist-example.db.tar.gz` and this is the file you need to pass to the client during a container creation
