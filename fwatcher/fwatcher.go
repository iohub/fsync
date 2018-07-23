package fwatcher

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsevents"
)

var (
	hostURL = `http://%s/upload?path=%s`
	curlObj = `"file=@%s"`
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
					PostFile(fname, projPath, syncHost)
				}
			}
		}
	}()

	in := bufio.NewReader(os.Stdin)
	log.Print("Started, press enter to stop")
	in.ReadString('\n')
	es.Stop()
}

func PostFile(fname string, projPath string, syncHost string) (*http.Response, error) {
	subpath := strings.TrimPrefix(fname, projPath)
	url := fmt.Sprintf(hostURL, syncHost, subpath)
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
