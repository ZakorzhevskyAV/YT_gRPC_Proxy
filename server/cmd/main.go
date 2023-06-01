package main

import (
	"context"
	"database/sql"
	"fmt"
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

func DBOpen(dbdirpath string) (db *sql.DB, err error) {
	err = os.MkdirAll(dbdirpath, 0666)
	if err != nil {
		log.Printf("Failed to create DB directory: %s", err)
		return nil, err
	}
	if entry, err := os.Stat(dbdirpath + "sqlite.db"); err != nil {
		err = os.WriteFile(dbdirpath+"sqlite.db", nil, 0666)
		if err != nil {
			log.Printf("Failed to create DB file: %s", err)
			return nil, err
		}
	} else if entry.IsDir() {
		log.Printf("DB file is already assigned to a directory")
		return nil, err
	}

	db, _ = sql.Open("sqlite3", dbdirpath+"sqlite.db")
	err = db.Ping()
	if err != nil {
		log.Printf("Failed to open DB file: %s", err)
		return nil, err
	}
	return db, nil
}

func DBInsert(db *sql.DB, link string, img []byte) (err error) {
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS images (image_link TEXT PRIMARY KEY, image_data BLOB)`)
	if err != nil {
		log.Printf("Failed to create table if not exists in DB file: %s", err)
		return err
	}
	stmt, err := db.Prepare("INSERT INTO images VALUES(?, ?)")
	if err != nil {
		log.Printf("Failed to prepare an insert statement for DB query: %s", err)
		return err
	}
	_, err = stmt.Exec(link, img)
	if err != nil {
		log.Printf("Failed to execute an insert statement for DB query: %s", err)
		return err
	}
	return nil
}

func DBSelect(db *sql.DB, link string) (img []byte, err error) {
	stmt, err := db.Prepare("SELECT image_data FROM images WHERE image_link=?")
	if err != nil {
		log.Printf("Failed to prepare a select statement for DB query: %s", err)
		return nil, err
	}
	rows, err := stmt.Query(link)
	if err != nil {
		log.Printf("Failed to execute a select statement for DB query: %s", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&img)
		if err != nil {
			log.Printf("Failed to scan rows from a select statement for DB query: %s", err)
			return nil, err
		}
	}
	err = rows.Err()
	if err != nil {
		log.Printf("Errors encountered during iteration over the rows: %s", err)
		return nil, err
	}
	return img, nil
}

func (s ThumbnailServer) Get(ctx context.Context, address *ytgrpcproxy.ThumbnailAddress) (*ytgrpcproxy.ThumbnailData, error) {
	log.Printf("Request acquired")
	log.Printf(address.Address)

	var img []byte
	var db *sql.DB
	var dbdirpath string
	var err error

	log.Printf("Opening DB to get thumbnail")
	dbdirpath = "./db/"
	db, err = DBOpen(dbdirpath)
	if err == nil {
		img, err = DBSelect(db, address.Address)
		if err != nil {
			log.Printf("Failed to get an image from DB: %s", err)
		} else if len(img) == 0 {
			log.Printf("No requested image in DB")
		} else {
			return SuccessOutput(img, nil)
		}
	} else {
		log.Printf("Failed to open DB file, trying to get the image from the web: %s", err)
	}
	defer db.Close()
	log.Printf("No DB file")

	log.Printf("Getting YT video page")
	response, err := http.Get(address.Address)
	if err != nil {
		log.Printf("Failed to get a response: %s", err)
		return ErrOutput("Failed to get a response", err)
	}
	log.Printf("Got YT video page")

	log.Printf("Preparing to parse YT video page")
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Printf("Failed to prepare HTML body for oarsing: %s", err)
		return ErrOutput("Failed to parse HTML body", err)
	}
	log.Printf("Prepared to parse YT video page")
	defer response.Body.Close()

	log.Printf("Parsing YT video page")
	selector := doc.Find("link")
	var ThumbnailLink string
	var attr bool
	selector.Each(func(i int, selectornew *goquery.Selection) {
		val, _ := selectornew.Attr("itemprop")
		if val == "thumbnailUrl" {
			ThumbnailLink, attr = selectornew.Attr("href")
		}
	})
	if !attr || ThumbnailLink == "" {
		log.Printf("Failed to find needed HTML element with needed attribute")
		return ErrOutput("Failed to find needed HTML element with needed attribute", fmt.Errorf("Failed to find needed HTML element with needed attribute"))
	}
	log.Printf("Parsed YT video page")

	log.Printf("Getting thumbnail")
	imgresponse, err := http.Get(ThumbnailLink)
	if err != nil {
		log.Printf("Failed to get a response: %s", err)
		return ErrOutput("Failed to get a response", err)
	}
	log.Printf("Got thumbnail")

	log.Printf("Converting thumbnail to bytes")
	img, err = io.ReadAll(imgresponse.Body)
	if err != nil {
		log.Printf("Failed to read response body to get the image: %s", err)
		return ErrOutput("Failed to read response body to get the image", err)
	}
	defer imgresponse.Body.Close()
	log.Printf("Converted thumbnail to bytes")

	log.Printf("Opening DB file")
	db, err = DBOpen(dbdirpath)
	if err != nil {
		log.Printf("Failed to open DB file: %s", err)
		return SuccessOutput(img, nil)
	}
	defer db.Close()
	log.Printf("Opened DB file")

	log.Printf("Inserting a row into DB file")
	err = DBInsert(db, address.Address, img)
	if err != nil {
		log.Printf("Failed to insert data into DB: %s", err)
		return SuccessOutput(img, nil)
	}
	log.Printf("Inserted a row into DB file")

	return SuccessOutput(img, nil)
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
