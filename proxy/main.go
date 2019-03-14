package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"
)

const (
	port = ":50052"
)

type (
	codec struct {
		parentCodec encoding.Codec
	}

	frame struct {
		payload []byte
	}
)

func newCodec() grpc.Codec {
	return &codec{
		parentCodec: encoding.GetCodec(proto.Name),
	}
}

func (c *codec) Marshal(v interface{}) ([]byte, error) {
	out, ok := v.(*frame)
	if !ok {
		return c.parentCodec.Marshal(v)
	}
	return out.payload, nil

}

func (c *codec) Unmarshal(data []byte, v interface{}) error {
	dst, ok := v.(*frame)
	if !ok {
		return c.parentCodec.Unmarshal(data, v)
	}
	dst.payload = data
	return nil
}

func (c *codec) Name() string {
	return "proxy"
}

func (c *codec) String() string {
	return c.Name()
}

func ProxyHandler(conn *grpc.ClientConn) grpc.StreamHandler {
	return func(_ interface{}, serverStream grpc.ServerStream) error {
		method, ok := grpc.MethodFromServerStream(serverStream)
		if !ok {
			return fmt.Errorf("unknown method")
		}
		fmt.Printf("%v\n", method)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		clientStream, err := conn.NewStream(ctx, &grpc.StreamDesc{ServerStreams: true, ClientStreams: true}, method)
		if err != nil {
			return err
		}

		for {
			m := &frame{}
			if err = serverStream.RecvMsg(m); err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			if err := clientStream.SendMsg(m); err != nil {
				return err
			}
		}

		for {
			m := &frame{}
			if err := clientStream.RecvMsg(m); err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			if err := serverStream.SendMsg(m); err != nil {
				return err
			}
		}

		return nil
	}
}

func main() {
	customCodec := newCodec()

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.CallCustomCodec(customCodec)))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.CustomCodec(customCodec),
		grpc.UnknownServiceHandler(ProxyHandler(conn)),
	)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}