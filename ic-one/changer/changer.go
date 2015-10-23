package main

import (
	"flag"
	"strconv"

	"github.com/empirefox/ic-client-one/storage"
	"github.com/golang/glog"
)

func init() {
	flag.Set("stderrthreshold", "INFO")
}

func main() {
	cpath := flag.String("cpath", "./room-dev.db", "config file path")
	cam := flag.String("cam", "", "camera id")
	attr := flag.String("attr", "", "attribute of the camera")
	attrval := flag.String("val", "", "value of the attribute")
	flag.Parse()
	val := *attrval
	if val == "" {
		glog.Errorln("Wrong val:", val)
		return
	}
	c := storage.NewConf(*cpath)
	if err := c.Open(); err != nil {
		glog.Errorln(err)
	}
	defer c.Close()

	i, err := c.GetIpcam([]byte(*cam))
	if err != nil {
		glog.Errorln(err)
		return
	}
	switch *attr {
	case "id":
		i.Id = val
	case "url":
		i.Url = val
	case "rec":
		i.Rec = AtoB(val)
	case "audio":
		i.AudioOff = AtoB(val)
	case "off":
		i.Off = AtoB(val)
	case "online":
		i.Online = AtoB(val)
	default:
		glog.Errorln("Wrong attr:", *attr)
		return
	}

	err = c.PutIpcam(&i)
	if err != nil {
		glog.Errorln(err)
		return
	}
}

func AtoB(val string) bool {
	b, _ := strconv.ParseBool(val)
	return b
}
