package main

import (
	"context"
	"log"
	"time"

	v1 "micro-scheduler/api/helloworld/v1"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/hashicorp/consul/api"
)

func main() {
	config := api.DefaultConfig()
	config.Address = "120.27.227.150:8500"
	consulCli, err := api.NewClient(config)
	if err != nil {
		panic(err)
	}
	r := consul.New(consulCli)

	// new grpc client
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///micro-scheduler"),
		grpc.WithDiscovery(r),
		grpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	gClient := v1.NewGreeterClient(conn)

	// new http client
	hConn, err := http.NewClient(
		context.Background(),
		http.WithMiddleware(
			recovery.Recovery(),
		),
		http.WithEndpoint("discovery:///micro-scheduler"),
		http.WithDiscovery(r),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer hConn.Close()
	hClient := v1.NewGreeterHTTPClient(hConn)

	for {
		time.Sleep(time.Second)
		callGRPC(gClient)
		callHTTP(hClient)
	}
}

func callGRPC(client v1.GreeterClient) {
	reply, err := client.SayHello(context.Background(), &v1.HelloRequest{Name: "kratos"})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[grpc] SayHello %+v\n", reply)
}

func callHTTP(client v1.GreeterHTTPClient) {
	reply, err := client.SayHello(context.Background(), &v1.HelloRequest{Name: "kratos"})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[http] SayHello %s\n", reply.Message)
}
