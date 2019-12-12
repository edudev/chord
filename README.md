# Installing Go and gRPC

## Go

Please refer here: https://golang.org/doc/install#install. Then set you GOPATH environment variable to a directory of your choice. Usually it's `${HOME}/go`. If you are on Debian/Ubuntu:

```sh
apt install golang
export GOPATH=${HOME}/go # it may be included in .profile .bashrc to make it persistent
```

## gRPC

Please refer here: https://grpc.io/docs/quickstart/go/. If you are on Debian/Ubuntu:

```sh
apt install protobuf-compiler libprotobuf-dev
```

# Downloading and building

```sh
CHORDDIR=${GOPATH}/src/github.com/edudev/chord
mkdir -p ${CHORDDIR}
git clone https://github.com/edudev/chord ${CHORDDIR}

# Getting and preparing dependencies
go get -d github.com/edudev/chord
go install github.com/golang/protobuf/protoc-gen-go
export PATH="${PATH}:${GOPATH}/bin"
protoc -I ${CHORDDIR}/server \
       ${CHORDDIR}/server/*.proto \
       --go_out=plugins=grpc:${CHORDDIR}/server

# Building project
go install github.com/edudev/chord
```

If you plan on runing chord with many nodes (40+) in a single process, add the following lines

```
*         hard    nofile      500000
*         soft    nofile      500000
root      hard    nofile      500000
root      soft    nofile      500000
```
as the end of `/etc/security/limits.conf`

# Running

## Chord Server

```sh
export PATH="${PATH}:${GOPATH}/bin"

# Simple, 64 nodes on localhost
chord

# N nodes on local host
chord -N <number of nodes>
chord -N 1 # runs chord with 1 local nodes

# Binding a hostname/address (with 64 nodes)
chord -addr <hostname|address>
chord -addr 127.0.0.1
chord -addr node001.ib.cluster

# Joining with 64 nodes to node001
chord -join <hostname:port|address:port>
chord -join node001:21210

# Optimisations and other options
chord -<optimisation name>
chord -fix_fingers # whether to intelligently fix fingers
chord -learn_nodes # whether to aggresively learn nodes for fingers
chord -log_hops # whether to log hop counts for predecessor search
```

First instance's port is always 21210. Second's is 21211 and increases one by one for each local node on a host.

## Stabilise

```sh
cd ${CHORDDIR}/cmd/stabilise

# Simple, 64 nodes, try to stabilise localhost
go run main.go

# To run stabilisation utility:
go run main.go -addr <bootstrap node:port> -N <ring size>
go run main.go -addr node001:21210 -N 1024 # stabilise the ring with 1024 nodes

# Optimisations and other options
go run main.go -<optimisation name>
go run main.go -check_fingers # also wait for fingertable to be correct
go run main.go -check_successors # also wait for successor list to be correct
```

`<ring size>` is total number of nodes. It is sum of multiple node instances running on individual hosts.

# Testing

## Memcached text protocol

```sh
$ nc -C localhost 11211
> get a
< END
> set a 0 0 1
> A
< STORED
> get a
< VALUE a 0 1
< A
< END
> delete a
< DELETED
> get a
< END
```

## Benchmarking on DAS-5

This setup assumes that you have `chord` and `stabilise` executables in `${HOME}/opt/bin`.

```sh
cd ${CHORDDIR}/test-scripts
./setup.sh # Only for first time
module load slurm # It may be already included in your .profile or .bashrc file

sbatch -N<das-5 nodes count> test-sbatch-chord.sh
sbatch -N4 test-sbatch-chord.sh # 2 servers for running chord, 2 servers for running benchmarks
```

You'll get and output looking like:

```
Submitted batch job <job-id>
Submitted batch job 01234567
```

You can watch the status with:

```sh
watch -n1 cat slurm-<job-id>.out
watch -n1 cat slurm-01234567.out
```

You can see the locations of files containing outputs of the benchmarking tools in output.
