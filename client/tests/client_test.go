package tests

import (
	"bytes"
	"fmt"
	"github.com/ZakorzhevskyAV/yt_gRPC_proxy/client"
	"github.com/ZakorzhevskyAV/yt_gRPC_proxy/ytgrpcproxy"
	"google.golang.org/grpc"
	"os"
	"testing"
)

func TestWriteImageToFile(t *testing.T) {
	var err error

	filename := fmt.Sprintf("test.jpg")

	_ = os.Remove(filename)

	main.WriteImageToFile([]byte("aaa"), err)
	if _, err = os.Stat(filename); err != nil {
		t.Errorf("File %s weren't created, error: %s", filename, err)
	} else {
		t.Logf("Data length written to file is %d", len([]byte("aaa")))
		if data, err := os.ReadFile(filename); err != nil {
			t.Errorf("Failed to read bytes from file %s, error: %s", filename, err)
		} else if bytes.Compare(data, []byte("aaa")) != 0 {
			t.Errorf("File data and data to be written are different, error")
		}
	}

	_ = os.Remove(filename)
}

// Требует запущенного gRPC сервера для выполнения
func TestSendRequest(t *testing.T) {
	var conn *grpc.ClientConn
	var link string

	conn, err := grpc.Dial(":8000", grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %s", err)
	}
	defer conn.Close()

	c := ytgrpcproxy.NewThumbnailReturnClient(conn)
	link = "https://www.youtube.com/watch?v=Gk-z2ykXfJo"

	main.SendRequest(c, link)

	filename := "image1.jpg"

	_ = os.Remove(filename)

	if _, err = os.Stat(filename); err != nil {
		t.Errorf("File %s weren't created, error: %s", filename, err)
	} else {
		if data, err := os.ReadFile(filename); err != nil {
			t.Errorf("Failed to read bytes from file %s, error: %s", filename, err)
		} else if len(data) == 0 {
			t.Errorf("Image file is empty, error")
		}
	}

	_ = os.Remove(filename)
}
