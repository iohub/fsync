package util

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/iohub/fsync/fwatcher"
)

func walkAndSync(dir string, syncHost string) {
	walker := func(fpath string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			fmt.Printf("[INFO] sync: %s\n", fpath)
			fwatcher.PostFile(fpath, dir, syncHost)
		}
		return nil
	}

	filepath.Walk(dir, walker)
}

func ForceSync(rootPath string, host string) {
	walkAndSync(rootPath, host)
}
