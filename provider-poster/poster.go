package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/golang/glog"
)

func init() {
	flag.Set("stderrthreshold", "INFO")
}

func main() {
	pfile := flag.String("src", "./provider-local.json", "json format provider")
	server := flag.String("server", "http://127.0.0.1:9999/sys/oauth", "server to post")
	flag.Parse()

	provider, err := ioutil.ReadFile(*pfile)
	if err != nil {
		glog.Errorln(err)
	}

	fmt.Printf("please make sure you want to post to\n%s\nwith:%s\nyes/no (yes):\n", *server, provider)
	reader := bufio.NewReader(os.Stdin)
	b, _, _ := reader.ReadLine()
	switch string(b) {
	case "", "yes":
	default:
		return
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	response, err := client.Post(*server, "application/json; charset=utf-8", bytes.NewReader(provider))
	if err != nil {
		glog.Errorln(err)
		return
	}
	glog.Infoln(response.StatusCode)
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		glog.Errorln(err)
		return
	}
	glog.Infoln(string(contents))
}
