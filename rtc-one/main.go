package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"os"

	"github.com/empirefox/ic-client-one-wrap"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

func init() {
	flag.Set("stderrthreshold", "INFO")
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

var dailer = websocket.Dialer{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

var conductor rtc.Conductor

// Stream #0:0: Video: mjpeg, yuvj422p(pc, bt470bg/unknown/unknown), 640x480 [SAR 1:1 DAR 4:3], 10 tbr, 90k tbn, 90k tbc
// Stream #0:1: Audio: pcm_mulaw, 8000 Hz, 1 channels, s16, 64 kb/s
var streamUrl1 = "rtsp://savage:qingqing@192.168.1.8:83/h.246.sdp"
var streamUrl2 = "rtsp://127.0.0.1:1235/test1.sdp"
var streamUrl3 = "rtsp://218.204.223.237:554/live/1/0547424F573B085C/gsfp90ef4k0a6iap.sdp"
var streamUrl = streamUrl1

func main() {
	flag.Parse()
	conductor = rtc.NewConductor(new(NullStatusObserver))
	addICE()
	_, ok := conductor.Registry("testcam", streamUrl, "/home/savage/111", false, true)
	if !ok {
		panic("cannot registry camera")
	}

	send := make(chan []byte, 64)
	quit := make(chan bool, 1)
	readLineToQuit(quit)
	startWs(send, quit)
	glog.Infoln("peer deleted!")
	readLineToQuit(quit)
	<-quit
	conductor.Release()
	glog.Infoln("Quit!")
}

type NullStatusObserver struct{}

func (NullStatusObserver) OnGangStatus(id string, status uint) {
	glog.Errorf("cam[%s] status now: %d\n", id, status)
}

type PeerMsg struct {
	Camera string `json:"camera,omitempty"`
	Type   string `json:"type,omitempty"`

	Candidate string `json:"candidate,omitempty"`
	Id        string `json:"id,omitempty"`
	Label     int    `json:"label,omitempty"`

	Sdp string `json:"sdp,omitempty"`
}

func readMsgs(ws *websocket.Conn, pc rtc.PeerConn, quit chan bool) {
	for {
		_, b, err := ws.ReadMessage()
		if err != nil {
			glog.Errorln(err)
			return
		}
		var msg PeerMsg
		if json.Unmarshal(b, &msg) == nil {
			switch msg.Type {
			case "offer":
				// offer comes
				glog.Infoln("offer comes after running")
			case "candidate":
				// cadidate comes
				glog.Infoln("add candidate")
				pc.AddCandidate(msg.Candidate, msg.Id, msg.Label)
			case "bye":
				quit <- true
			default:
				glog.Errorln("got unknow json message:", string(b))
			}
		}
	}
}

func startWs(send chan []byte, quit chan bool) {
	ws, _, err := dailer.Dial("ws://127.0.0.1:9999/one", nil)
	if err != nil {
		glog.Errorln(err)
		return
	}
	defer ws.Close()

	glog.Infoln("ws connected")
	_, b, err := ws.ReadMessage()
	if err != nil {
		glog.Errorln(err)
		return
	}
	var offer PeerMsg
	if json.Unmarshal(b, &offer) != nil || offer.Type != "offer" {
		glog.Errorln("must be offer, but:", offer)
		return
	}

	// offer comes
	pc := conductor.CreatePeer("testcam", func(msg []byte) { send <- msg })
	defer conductor.DeletePeer(pc)
	pc.CreateAnswer(offer.Sdp)
	glog.Infoln("CreateAnswer ok")
	go readMsgs(ws, pc, quit)

	for {
		select {
		case msg, ok := <-send:
			if !ok {
				return
			}
			if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-quit:
			return
		}
	}
}

func addICE() {
	conductor.AddIceServer("stun:stun.l.google.com:19302", "", "")
	conductor.AddIceServer("stun:stun.anyfirewall.com:3478", "", "")
	conductor.AddIceServer("turn:turn.bistri.com:80", "homeo", "homeo")
	conductor.AddIceServer("turn:turn.anyfirewall.com:443?transport=tcp", "webrtc", "webrtc")
}

func readLineToQuit(quit chan bool) {
	reader := bufio.NewReader(os.Stdin)
	go func() {
		reader.ReadLine()
		quit <- true
	}()
}
