package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsevents"
	"github.com/iohub/fsync/util"
)

var noteDescription = map[fsevents.EventFlags]string{
	fsevents.MustScanSubDirs:   "MustScanSubdirs",
	fsevents.UserDropped:       "UserDropped",
	fsevents.KernelDropped:     "KernelDropped",
	fsevents.EventIDsWrapped:   "EventIDsWrapped",
	fsevents.HistoryDone:       "HistoryDone",
	fsevents.RootChanged:       "RootChanged",
	fsevents.Mount:             "Mount",
	fsevents.Unmount:           "Unmount",
	fsevents.ItemCreated:       "Created",
	fsevents.ItemRemoved:       "Removed",
	fsevents.ItemInodeMetaMod:  "InodeMetaMod",
	fsevents.ItemRenamed:       "Renamed",
	fsevents.ItemModified:      "Modified",
	fsevents.ItemFinderInfoMod: "FinderInfoMod",
	fsevents.ItemChangeOwner:   "ChangeOwner",
	fsevents.ItemXattrMod:      "XAttrMod",
	fsevents.ItemIsFile:        "IsFile",
	fsevents.ItemIsDir:         "IsDir",
	fsevents.ItemIsSymlink:     "IsSymLink",
}

func WatchPath(path string, syncHost string) {
	absPath, _ := filepath.Abs(path)
	projPath := absPath + "/"
	dev, err := fsevents.DeviceForPath(path)
	if err != nil {
		log.Fatalf("Failed to retrieve device for path: %v", err)
	}

	es := &fsevents.EventStream{
		Paths:   []string{path},
		Latency: 500 * time.Millisecond,
		Device:  dev,
		Flags:   fsevents.FileEvents | fsevents.WatchRoot,
	}
	es.Start()
	ec := es.Events

	log.Println("Device UUID", fsevents.GetDeviceUUID(dev))

	go func() {
		for msg := range ec {
			for _, event := range msg {
				if fname, ok := isFileModified(event); ok {
					logEvent(event)
					subpath := strings.TrimPrefix(fname, projPath)
					util.PostFile(fname, fmt.Sprintf(util.UrlParamFormat, syncHost, subpath))
					time.Sleep(300 * time.Microsecond)
				}
			}
		}
	}()

	in := bufio.NewReader(os.Stdin)
	log.Print("Started, press enter to stop")
	in.ReadString('\n')
	es.Stop()
}

func isFileModified(ev fsevents.Event) (string, bool) {
	if ev.Flags&fsevents.ItemIsFile == fsevents.ItemIsFile &&
		ev.Flags&fsevents.ItemModified == fsevents.ItemModified {

		return "/" + ev.Path, true
	}
	return "", false
}

func logEvent(event fsevents.Event) {
	note := ""
	for bit, description := range noteDescription {
		if event.Flags&bit == bit {
			note += description + " "
		}
	}
	log.Printf("EventID: %d Path: %s Flags: %s", event.ID, event.Path, note)
}

func main() {
	paction := flag.String("action", "watch", "watch or sync")
	ppath := flag.String("path", ".", "path to watch or path to sync")
	phost := flag.String("host", "127.0.0.1:8080", "remote addr")
	flag.Parse()

	action := *paction
	switch action {
	case "watch":
		WatchPath(*ppath, *phost)
	case "sync":
		util.ForceSync(*ppath, *phost)
	default:
		panic("unkown action")
	}
}
