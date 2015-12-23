package main

import (
	"flag"

	"github.com/empirefox/ic-client-one/storage"
	"github.com/golang/glog"
)

func init() {
	flag.Set("stderrthreshold", "INFO")
}

func main() {
	cpath := flag.String("c", "./room.db", "new file path")
	regToken := flag.String("t", "", "new reg token")
	server := flag.String("s", "gocamcom.daoapp.io", "ctrl server")
	flag.Parse()

	c := storage.NewConf(*cpath)
	if err := c.Open(); err != nil {
		glog.Errorln(err)
	}
	defer c.Close()

	if *regToken != "" {
		c.Put(storage.K_REG_TOKEN, []byte(*regToken))
	}
	glog.Errorln(string(c.Get(storage.K_REG_TOKEN)))

	c.Put(storage.K_SERVER, []byte(*server))

	for _, ic := range c.GetIpcams() {
		glog.Exitln(ic.Url)
	}

}
