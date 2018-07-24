package main

import (
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func GinServe(projDir string, port string) {
	projPath, _ := filepath.Abs(projDir)
	router := gin.Default()
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.POST("/upload", func(c *gin.Context) {
		path := c.Query("path")
		file, err := c.FormFile("file")
		fmt.Printf("path:%s\n", path)
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
			return
		}
		if err := saveUploadedFile(file, path, projPath); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}
		c.String(http.StatusOK, fmt.Sprintf("File %s sync successfully.", file.Filename))
	})
	router.Run(port)
}

func mkdir(path string) error {
	cmd := exec.Command("mkdir", "-p", path)
	if err := cmd.Start(); err != nil {
		fmt.Printf("[ERROR] mkdir:%s", path)
		return err
	}
	return nil
}

func saveUploadedFile(file *multipart.FileHeader, originPath string, projPath string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	idx := strings.LastIndex(originPath, "/")
	subObj := originPath
	subDir := projPath
	if idx != -1 {
		subObj = originPath[0:idx]
		subDir = subDir + "/" + subObj
	}
	fmt.Printf("[INFO] subdir: %s\n", subDir)
	mkdir(subDir)
	fobj := projPath + "/" + originPath
	fmt.Printf("[INFO] save filename: %s\n", fobj)
	out, err := os.Create(fobj)
	if err != nil {
		return err
	}
	defer out.Close()
	io.Copy(out, src)
	return nil
}

func main() {
	ppath := flag.String("path", ".", "path to sync")
	pport := flag.String("port", "8080", "port")
	flag.Parse()

	GinServe(*ppath, ":"+*pport)
}
