# Examples

## Single Server

1. start a server
```shell
go run single/main.go -http :8080
```
2. start a client

```shell
go run client/main.go -server localhost:8080 -id 1
```
3. start another client
```shell
go run client/main.go -server localhost:8080 -id 2
```
4. And now you can send messages between two clients by typing `[id] [conent]` in client command line terminal. For example, type in first client terminal:
```shell
2 hello
```
and you'll see in second client terminal:
```shell
>>: {"to":"2","from":"1","message":"hello"}
```

## Clusters

> Note: To run the cluster example, you should prepare a redis server that serving on 'localhost:6379', or modify the source code as you wish.

1. start a server:
```shell
go run cluster/main.go -http :8081 -grpc :9001
```

2. start another server:
```shell
go run cluster/main.go -http :8082 -grpc :9002
```

3. start a client connected to first server:
```shell
go run client/main.go -server localhost:8081 -id 1
```
4. start another client connected to second server:
```shell
go run client/main.go -server localhost:8082 -id 2
```

5. And now you can send messages between two clients by typing `[id] [conent]` in client command line terminal. For example, type in first client terminal:
```shell
2 hello
```
and you'll see in second client terminal:
```shell
>>: {"to":"2","from":"1","message":"hello"}
```

