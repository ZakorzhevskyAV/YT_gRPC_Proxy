package tests

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"github.com/ZakorzhevskyAV/yt_gRPC_proxy/server/cmd"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestDBOpen(t *testing.T) {
	dbdirpath := "./testdb/"

	_ = os.RemoveAll(dbdirpath)

	_, err := main.DBOpen(dbdirpath)
	if err != nil {
		t.Errorf("Failed to create and open DB file, error: %s", err)
	}
	_ = os.RemoveAll(dbdirpath)

	_ = os.MkdirAll(dbdirpath+"sqlite.db", 0666)
	_, err = main.DBOpen(dbdirpath)
	if err == nil {
		t.Errorf("Supposed DB file name is a directory name yet no error, error")
	}
	_ = os.RemoveAll(dbdirpath)

	_, err = os.Create(dbdirpath + "sqlite.db")
	_, err = main.DBOpen(dbdirpath)
	if err != nil {
		t.Errorf("Failed to open DB file, error: %s", err)
	}
	_ = os.RemoveAll(dbdirpath)
}

func TestDBInsertSelect(t *testing.T) {
	dbdirpath := "./testdb/"

	_ = os.RemoveAll(dbdirpath)

	db, err := main.DBOpen(dbdirpath)
	if err != nil {
		t.Errorf("Failed to create and open DB file, error: %s", err)
	}
	defer db.Close()

	var link string
	var data []byte

	link = "teststring"
	data = []byte("aaa")

	err = main.DBInsert(db, link, data)
	if err != nil {
		t.Errorf("Failed to insert data into DB, error: %s", err)
	}

	selecteddata, err := main.DBSelect(db, link)
	if err != nil {
		t.Errorf("Failed to select data into DB, error: %s", err)
	}
	if bytes.Compare(data, selecteddata) != 0 {
		t.Errorf("Data inserted and data selected are different, error: %s", err)
	}

	_ = os.RemoveAll(dbdirpath)
}

func TestDownloadingAndParsing(t *testing.T) {
	response, err := http.Get("https://www.youtube.com/watch?v=Gk-z2ykXfJo")
	if err != nil {
		t.Errorf("Failed to get a response from YT video page, error: %s", err)
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		t.Errorf("Failed to prepare YT video page for parsing, error: %s", err)
	}
	defer response.Body.Close()

	selector := doc.Find("link")
	var ThumbnailLink string
	var attr bool
	selector.Each(func(i int, selectornew *goquery.Selection) {
		val, _ := selectornew.Attr("itemprop")
		if val == "thumbnailUrl" {
			ThumbnailLink, attr = selectornew.Attr("href")
		}
		if !attr || ThumbnailLink == "" {
			t.Errorf("Failed to find needed HTML element with needed attribute")
		}
	})

	imgresponse, err := http.Get(ThumbnailLink)
	if err != nil {
		t.Errorf("Failed to get a response from YT video thumbnail page, error: %s", err)
	}

	//var img []byte
	_, err = io.ReadAll(imgresponse.Body)
	if err != nil {
		t.Errorf("Failed to read response body to get the image, error: %s", err)
	}
	defer imgresponse.Body.Close()
}
