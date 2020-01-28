# go-shuwdown

Golang app shutdown hooks

# example

```go
func StartServer() {
    lis, err := net.Listen("tcp", "127.0.0.1")
    if err != nil {
    	panic(err)
    }
   
    server := grpc.NewServer()
    shutdown.Add(func() {
    	server.GracefulStop()
    })
    server.Serve(lis)
}

func main() {
    go StartServer1()
    go StartServer2()
    ...

    shutdown.C()
}
```