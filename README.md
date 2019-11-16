# Downloading dependencies
```
go get
```


# Building
```
protoc -I server/ server/ring.proto --go_out=plugins=grpc:server
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

