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
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/golang/glog"
)

func init() {
	flag.Set("stderrthreshold", "INFO")
}

func main() {
	pfile := flag.String("src", "./provider-local.json", "json format provider")
	server := flag.String("server", "http://127.0.0.1:9998/sys/oauth", "server to post")
	secret := flag.String("secret", "s", "secret for auth")
	flag.Parse()

	token := jwt.New(jwt.SigningMethodHS256)
	token.Header["kid"] = "system"
	token.Claims["exp"] = time.Now().Add(time.Second * 10).Unix()
	tokenString, err := token.SignedString([]byte(*secret))
	if err != nil {
		glog.Errorln(err)
		return
	}

	provider, err := ioutil.ReadFile(*pfile)
	if err != nil {
		glog.Errorln(err)
		return
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
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("POST", *server, bytes.NewReader(provider))
	if err != nil {
		glog.Errorln(err)
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+tokenString)

	response, err := client.Do(req)
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
