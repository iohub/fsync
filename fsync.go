package main

import (
	"flag"
	"fmt"

	"github.com/iohub/fsync/fwatcher"
	"github.com/iohub/fsync/ginsync"
	"github.com/iohub/fsync/util"
)

func main() {
	paction := flag.String("action", "fwatcher", "fwatcher or fsync")
	ppath := flag.String("path", ".", "path to watch or path to sync")
	paddr := flag.String("addr", "127.0.0.1", "addr to listen or remote addr")
	pport := flag.String("port", "8080", "port")
	flag.Parse()

	action := *paction
	switch action {
	case "fwatcher":
		host := fmt.Sprintf("%s:%s", *paddr, *pport)
		fwatcher.WatchPath(*ppath, host)
	case "freceiver":
		ginsync.GinServe(*ppath, ":"+*pport)
	case "fsync":
		host := fmt.Sprintf("%s:%s", *paddr, *pport)
		util.ForceSync(*ppath, host)
	default:
		panic("unkown action")
	}
}
