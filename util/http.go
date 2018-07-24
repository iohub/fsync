package util

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func PostFile(fname string, url string) (*http.Response, error) {
	buf := bytes.NewBufferString("")
	writer := multipart.NewWriter(buf)
	_, err := writer.CreateFormFile("file", fname)
	if err != nil {
		fmt.Println("error writing to buffer")
		return nil, err
	}
	fh, err := os.Open(fname)
	if err != nil {
		fmt.Println("error opening file")
		return nil, err
	}
	boundary := writer.Boundary()
	cbuf := bytes.NewBufferString(fmt.Sprintf("\r\n--%s--\r\n", boundary))
	reader := io.MultiReader(buf, fh, cbuf)
	fi, err := fh.Stat()
	if err != nil {
		fmt.Printf("Error Stating file: %s", fname)
		return nil, err
	}
	req, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "multipart/form-data; boundary="+boundary)
	req.ContentLength = fi.Size() + int64(buf.Len()) + int64(cbuf.Len())

	return http.DefaultClient.Do(req)
}
