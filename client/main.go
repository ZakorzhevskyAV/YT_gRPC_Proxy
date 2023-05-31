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

	link := flag.Arg(0)

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
	var filename string
	for i := 0; i < 1000; i++ {
		filename = fmt.Sprintf("image%d.jpg", i)
		if _, err = os.Stat(filename); err != nil {
			err = os.WriteFile(filename, response.Data, 0666)
			if err != nil {
				log.Fatalf("Failed to write an image into a file: %s", err)
			}
			break
		} else {
			continue
		}
	}
	fmt.Printf("Response from server acquired")

}
