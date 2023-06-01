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

func WriteImageToFile(data []byte, err error) {
	for i := 0; i < 1000; i++ {
		filename := fmt.Sprintf("image%d.jpg", i)
		if _, err = os.Stat(filename); err != nil {
			err = os.WriteFile(filename, data, 0666)
			if err != nil {
				log.Fatalf("Failed to write an image into a file: %s", err)
			}
			break
		} else {
			continue
		}
	}
}

func SendRequest(c ytgrpcproxy.ThumbnailReturnClient, link string) {
	response, err := c.Get(context.Background(), &ytgrpcproxy.ThumbnailAddress{Address: link})
	if err != nil {
		log.Fatalf("Error when trying to get YT thumbnail: %s", err)
	}
	if len(response.Data) == 0 {
		return
	}
	WriteImageToFile(response.Data, err)
	return
}

func main() {

	flag.Usage = func() {
		fmt.Printf("Usage of this CLI app: write '<app_name> <YT video link>' to get the YT video thumbnail picture downloaded at the same directory as this command-line utility.\n" +
			"Use -(-)async=on flag to execute requests asynchronously.\n" +
			"Arguments have to be sent after -(-)async flag.\n")
		flag.PrintDefaults()
	}
	asyncflag := flag.String("async", "off", "Use -(-)async=on flag to execute requests asynchronously.")

	flag.Parse()

	links := flag.Args()

	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":8000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %s", err)
	}
	defer conn.Close()

	c := ytgrpcproxy.NewThumbnailReturnClient(conn)
	var channel chan int
	channel = make(chan int)

	if *asyncflag == "on" {
		for i, link := range links {
			go func() {
				SendRequest(c, link)
				channel <- 1
			}()
			<-channel
			fmt.Printf("Request %d sent, response from server acquired\n", i)
		}
	} else {
		for i, link := range links {
			SendRequest(c, link)
			fmt.Printf("Request %d sent, response from server acquired\n", i)
		}
	}

}
