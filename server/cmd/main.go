package main

import (
	"context"
	"database/sql"
	"github.com/PuerkitoBio/goquery"
	"github.com/ZakorzhevskyAV/yt_gRPC_proxy/ytgrpcproxy"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

type ThumbnailServer struct {
	ytgrpcproxy.UnimplementedThumbnailReturnServer
}

func ErrOutput(data string, err error) (*ytgrpcproxy.ThumbnailData, error) {
	return &ytgrpcproxy.ThumbnailData{
		Data: []byte(data),
	}, err
}

func SuccessOutput(data []byte, err error) (*ytgrpcproxy.ThumbnailData, error) {
	return &ytgrpcproxy.ThumbnailData{
		Data: data,
	}, err
}

func DBOpen() (db *sql.DB, err error) {
	err = os.MkdirAll("../../db", 0666)
	if err != nil {
		log.Printf("Failed to create DB directory: %s", err)
		return nil, err
	}
	if entry, err := os.Stat("../../db/sqlite.db"); err != nil {
		file, err := os.Create("../../db/sqlite.db")
		if err != nil {
			log.Printf("Failed to create DB file: %s", err)
			return nil, err
		}
	} else if entry.IsDir() {
		log.Printf("DB file is already assigned to a directory")
		return nil, err
	}
	db, err = sql.Open("sqlite3", "../../db/sqlite.db")
	if err != nil {
		log.Printf("Failed to open DB file: %s", err)
		return nil, err
	}
	return db, nil
}

func (s ThumbnailServer) Get(ctx context.Context, address *ytgrpcproxy.ThumbnailAddress) (*ytgrpcproxy.ThumbnailData, error) {
	log.Printf("Request acquired")
	log.Printf(address.Address)
	response, err := http.Get(address.Address)
	if err != nil {
		log.Printf("Failed to get a response: %s", err)
		return ErrOutput("Failed to get a response", err)
	}
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Printf("Failed to parse HTML body: %s", err)
		return ErrOutput("Failed to parse HTML body", err)
	}
	selector := doc.Find("link")
	var ThumbnailLink string
	selector.Each(func(i int, selectornew *goquery.Selection) {
		val, _ := selectornew.Attr("itemprop")
		if val == "thumbnailUrl" {
			ThumbnailLink, _ = selectornew.Attr("href")
		}
	})
	imgresponse, err := http.Get(ThumbnailLink)
	if err != nil {
		log.Printf("Failed to get a response: %s", err)
		return ErrOutput("Failed to get a response", err)
	}
	img, err := io.ReadAll(imgresponse.Body)
	if err != nil {
		log.Printf("Failed to read response body to get the image: %s", err)
		return ErrOutput("Failed to read response body to get the image", err)
	}
	var db *sql.DB
	db, err = DBOpen()
	if err != nil {
		log.Printf("Failed to read response body to get the image: %s", err)
		return SuccessOutput(img, nil)
	}
	err = os.WriteFile("image.jpg", img, 0666)
	if err != nil {
		log.Printf("Failed to write an image into a file: %s", err)
	}
	return &ytgrpcproxy.ThumbnailData{
		Data: img,
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
