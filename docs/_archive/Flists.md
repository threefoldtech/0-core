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

## How to merge two flist into a single one:

Why merging two flist ? Imagine you two flist container two of your application and you want to deploy a container inside which you'll run both application. To do so, you need a flist with the files of your two application, thus you need to merge them together.

The flist librairy in Jumpscale provide you with the tool to do exactly that. Here is an example script:

```python
# uncomress flist db
j.sal.fs.targzUncompress('/tmp/app_a.db.tar.gz','/tmp/app_a.db')
# open the connection to the rocksdb of the application A
kvs = j.servers.kvs.getRocksDBStore(name='app_A', namespace=None, dbpath="/tmp/app_a.db")
# create the flist object
flist_app_a = j.tools.flist.getFlist(rootpath='/', kvs=kvs)

# uncomress flist db
j.sal.fs.targzUncompress('/tmp/app_b.db.tar.gz','/tmp/app_b.db')
# open the connection to the rocksdb of the application B
kvs = j.servers.kvs.getRocksDBStore(name='app_B', namespace=None, dbpath="/tmp/app_b.db")
# create the flist object
fardb = j.tools.flist.getFlist(rootpath='/', kvs=kvs)

# open the connection to the destination rocksdb where we are goign to store the merged flist
kvs = j.servers.kvs.getRocksDBStore(name='flist', namespace=None, dbpath="/tmp/merge.db")
fdest = j.tools.flist.getFlist(rootpath='/', kvs=kvs)

# instanciate a FlistMerger object
merger = FlistMerger()
# Add both of your input flist as source
merger.add_source(fjs)
merger.add_source(fardb)
# Set the destination flist object
merger.add_destination(fdest)
# Do the merge
merger.merge()
# check the result
fdest.pprint()

# package the merged flist rocksdb
j.sal.fs.targzCompress('/tmp/merge.db', '/tmp/merge.db.tar.gz')
# remove uncompressed rocksdb
j.sal.fs.removeDirTree('/tmp/merge.db')
```
