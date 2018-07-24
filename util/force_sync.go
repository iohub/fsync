package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func walkAndSync(rootPath string, syncHost string) {
	walker := func(fpath string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			fmt.Printf("[INFO] sync: %s\n", fpath)
			subpath := strings.TrimPrefix(fpath, rootPath)
			PostFile(fpath, fmt.Sprintf(UrlParamFormat, syncHost, subpath))
		}
		return nil
	}

	filepath.Walk(rootPath, walker)
}

func ForceSync(rootPath string, host string) {
	walkAndSync(rootPath, host)
}
