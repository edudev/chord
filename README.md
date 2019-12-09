# Downloading dependencies

```
go get
```

# Local environment setup
If you plan on runing chord with many nodes (40+) in a single process,
add the following lines
```
*         hard    nofile      500000
*         soft    nofile      500000
root      hard    nofile      500000
root      soft    nofile      500000
```
as the end of `/etc/security/limits.conf`

# Building

```
protoc -I server/ server/*.proto --go_out=plugins=grpc:server
go build
```

# Running

```
./chord
```

# Testing

```
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
