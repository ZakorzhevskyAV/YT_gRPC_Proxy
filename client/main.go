package main

import (
	"flag"
	"fmt"
	"github.com/ZakorzhevskyAV/yt_gRPC_proxy/ytgrpcproxy"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"os"
)

func main() {

	flag.Usage = func() {
		fmt.Printf("Usage of %s: write '%s <YT video link>' to get the thumbnail picture downloaded at the same directory as this command-line utility", os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	link := flag.Arg(1)

	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":8000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := ytgrpcproxy.NewThumbnailReturnClient(conn)

	response, err := c.Get(context.Background(), &ytgrpcproxy.ThumbnailAddress{Address: link})
	if err != nil {
		log.Fatalf("Error when trying to get YT thumbnail: %s", err)
	}
	fmt.Printf(string(response.Data))
	fmt.Printf("Response from server acquired")

}
