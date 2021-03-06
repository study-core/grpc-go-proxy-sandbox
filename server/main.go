package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/utahta/grpc-go-proxy-sandbox/helloworld"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
//type server struct{}
type server struct {
	helloworld.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	if t, ok := ctx.Deadline(); ok {
		fmt.Printf("in DEADLINE: %v\n", t)
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		fmt.Printf("Incoming: abe %v\n", md.Get("abe")[0])
	}

	log.Printf("Received: %v\n", in.Name)
	return &helloworld.HelloReply{Message: "Hello " + in.Name}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	helloworld.RegisterGreeterServer(s, &server{})  // zhu ce fu wu

	fmt.Printf("Start Server :%s ... \n", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
