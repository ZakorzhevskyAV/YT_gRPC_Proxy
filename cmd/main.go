package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
	"yt_gRPC_proxy/ytgrpcproxy"
)

type ThumbnailServer struct {
	ytgrpcproxy.UnimplementedThumbnailReturnServer
}

func (s ThumbnailServer) Get(ctx context.Context, address *ytgrpcproxy.ThumbnailAddress) (*ytgrpcproxy.ThumbnailData, error) {
	return &ytgrpcproxy.ThumbnailData{
		Data: []byte("aaa"),
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatalf("cannot create listener: %s", err)
	}

	serverRegister := grpc.NewServer()
	service := &ThumbnailServer{}

	ytgrpcproxy.RegisterThumbnailReturnServer(serverRegister, service)
	err = serverRegister.Serve(lis)
	if err != nil {
		log.Fatalf("serving failed: %s", err)
	}

}
