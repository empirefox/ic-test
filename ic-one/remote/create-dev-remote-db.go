package main

import (
	"flag"

	"github.com/dchest/uniuri"
	"github.com/empirefox/ic-client-one/ipcam"
	"github.com/empirefox/ic-client-one/storage"
	"github.com/golang/glog"
)

func init() {
	flag.Set("stderrthreshold", "INFO")
}

func main() {
	cpath := flag.String("cpath", "./room-dev-remote.db", "config file path")
	flag.Parse()
	c := storage.NewConf(*cpath)
	if err := c.Open(); err != nil {
		glog.Errorln(err)
	}
	defer c.Close()

	c.Put(storage.K_REC_DIR, []byte("ipcam-records-remote-dev"))

	c.PutIpcam(&ipcam.Ipcam{
		Id:     "Mock_" + uniuri.New(),
		Url:    "rtsp://savage:qingqing@192.168.1.8:83/h.246.sdp",
		Rec:    false,
		Off:    false,
		Online: true,
	})

	glog.Errorln("created!")
}
