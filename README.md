# A distributed key-value store using the Chord protocol
This repository contains the implementation of a distributed key-value store based on the Chord protocol. The Chord paper followed for this project can be found [here](https://pdos.csail.mit.edu/papers/chord:sigcomm01/chord_sigcomm.pdf "Chord paper").

# Getting started
## Prerequisites
Golang, gRPC and Protocol buffers are needed to run this project. The following instructions will get you a copy of the project up and running on your local machine for development and testing purposes. 

### Go
For running this project smoothly, you need `golang version >= 1.10`. If you are on `Debian buster` or `Ubuntu bionic`, run the following commands:

```sh
apt install golang
export GOPATH=${HOME}/go
```
For other distributions, please refer to golang installation page [here](https://golang.org/doc/install "Golang installation page"). You would also need to set `GOROOT` and `GOPATH` environment variable.

The `GOPATH` environment variable specifies the location of your workspace. If you want to use a custom location as your workspace, you can set the `GOPATH environment variable` by exporting in `.bashrc`, `.zshrc` etc. This [page](https://github.com/golang/go/wiki/SettingGOPATH "Setting GOPATH") explains how to set this variable on various platforms.

### gRPC and Protocol Buffers

If you are on `Debian` or `Ubuntu`, run the following command:

```sh
apt install protobuf-compiler libprotobuf-dev
```
For more information, please visit this [link](https://grpc.io/docs/quickstart/go/ "gRPC installation")

## Basic Setup: Downloading and building

For downloading this repository and building it, run the following commands in given order:

1. Download the repository
    ```sh
    CHORDDIR=${GOPATH}/src/github.com/edudev/chord
    mkdir -p ${CHORDDIR}
    git clone https://github.com/edudev/chord ${CHORDDIR}
    ```

2. Retrieve and prepare dependencies

    ```sh
    go get -d github.com/edudev/chord
    go install github.com/golang/protobuf/protoc-gen-go
    export PATH="${PATH}:${GOPATH}/bin"
    protoc -I ${CHORDDIR}/server \
           ${CHORDDIR}/server/*.proto \
           --go_out=plugins=grpc:${CHORDDIR}/server
    ```
    
3. Build the project
    ```sh
    go install github.com/edudev/chord
    ```

**NOTE:** If you plan on runing chord with many nodes (40+) in a single process, add the following lines at the end of `/etc/security/limits.conf`. **Rebooting of computer is required to refresh the limits.**
```
*         hard    nofile      500000
*         soft    nofile      500000
root      hard    nofile      500000
root      soft    nofile      500000
```

# Usage
Once the installation and building of the project is over, you are ready to go :) `[pun intended]`

## Chord Server
To run the chord server, run the follwing command to run the project without going to the repository folder:

```sh
export PATH="${PATH}:${GOPATH}/bin"
```

- To run the project with 64 chord nodes (default) on localhost, run:
    ```sh
    chord
    ```

- If you want to set the number of nodes, run the `chord` command with `-N` flag.
    ```sh
    chord -N <number of nodes>
    chord -N 1 # runs chord with 1 local nodes
    ```

- If you want to bind chord program to localhost, run with flag `-addr 127.0.0.1`. For custom hostname/address, binding can be done by using following templates:
    ```sh
    # Binding a hostname/address with 64 nodes (default)
    chord -addr 127.0.0.1
    chord -addr <hostname|address>
    chord -addr node001.ib.cluster
    ```
- If you want to join server nodes from one machine to another machine, run the following commands:
    ```sh
    # Joining with 64 nodes (default) to node001
    chord -join <hostname:port|address:port>
    chord -addr 127.0.0.2 -join 127.0.0.1:21210  
    chord -addr node002 -join node001:21210
    ```

- For running optimisations for fixing fingers, learning nodes and counting log hops, run the following commands:
    ```sh
    # Optimisations and other options
    chord -<optimisation name>
    chord -fix_fingers # whether to intelligently fix fingers
    chord -learn_nodes # whether to aggresively learn nodes for fingers
    chord -log_hops # whether to log hop counts for predecessor search
    ```

First virtual node's port is always `21210`. Second's is `21211` and it increases one by one for each node on the ring.

## Stabilisation
For running the stabilisation part on the ring, chord instance should be already running.

If it is not running, either run:
```sh
chord
```

or run following command at project directory `${CHORDDIR}`:
```sh
go run main.go
```

Now change into `stabilise` directory and run:
```sh
cd ${CHORDDIR}/cmd/stabilise
```

- To stabilise 64 nodes (default) on localhost, run:
    ```sh
    go run main.go
    ```

- To run stabilisation utilities on different machine:
    ```sh
    # To run stabilisation utility:
    go run main.go -addr <bootstrap node:port> -N <ring size>
    go run main.go -addr node001:21210 -N 1024 # stabilise the ring with 1024 nodes
    ```

- For other optimisations:
    ```
    # Optimisations and other options
    go run main.go -<optimisation name>
    go run main.go -check_fingers # also wait for fingertable to be correct
    go run main.go -check_successors # also wait for successor list to be correct
    ```

`<ring size>` is total number of nodes. It is sum of the multiple virtual node instances running on individual hosts.

# Testing
## Memcached text protocol
For testing purposes, if you want to access a chord server, run a client instance with either `netcat` or `telnet` while running chord instance on another terminal:

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

This setup assumes that you have `chord` and `stabilise` executables in `${HOME}/opt/bin` on DAS-5.

```sh
cd ${CHORDDIR}/test-scripts
./setup.sh # Only for first time
module load slurm # It may be already included in your .profile or .bashrc file

sbatch -N<das-5 nodes count> test-sbatch-chord.sh
sbatch -N4 test-sbatch-chord.sh # 2 servers for running chord, 2 servers for running benchmarks
```

You'll get and output like:

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

# Contributors
This project is made by a team of 5 members for "Distributed Systems" course at Vrije Universiteit Amsterdam. Their names in alphabetical order are:

1. D. Casenove
2. E. Dudev
3. J.M. Hohnerlein
4. R.S. Keskin
5. S. Kapoor 
