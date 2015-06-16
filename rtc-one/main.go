package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"

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

type PeerMsg struct {
	Candidate string `json:"candidate,omitempty"`
	Mid       string `json:"sdpMid,omitempty"`
	Line      int    `json:"sdpMLineIndex,omitempty"`

	Type string `json:"type,omitempty"`
	Sdp  string `json:"sdp,omitempty"`
}

func readMsgs(ws *websocket.Conn, pc rtc.PeerConn) {
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
				pc.AddCandidate(msg.Candidate, msg.Mid, msg.Line)
			default:
				glog.Errorln("got unknow json message:", string(b))
			}
		}
	}
}

func startWs() {
	ws, _, err := dailer.Dial("ws://192.168.1.222:9999/one", nil)
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

	send := make(chan []byte, 64)

	// offer comes
	pc := conductor.CreatePeer("rtsp://127.0.0.1:1235/test1.sdp", send)
	pc.CreateAnswer(offer.Sdp)
	glog.Infoln("CreateAnswer ok")
	go readMsgs(ws, pc)

	for {
		select {
		case msg, ok := <-send:
			if !ok {
				return
			}
			if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}
}

func addICE() {
	conductor.AddIceServer("stun:stun.l.google.com:19302", "", "")
	conductor.AddIceServer("stun:stun.anyfirewall.com:3478", "", "")
	conductor.AddIceServer("turn:turn.bistri.com:80", "homeo", "homeo")
	conductor.AddIceServer("turn:turn.anyfirewall.com:443?transport=tcp", "webrtc", "webrtc")
}

func main() {
	flag.Parse()
	conductor = rtc.NewConductor()
	addICE()
	conductor.Registry("rtsp://127.0.0.1:1235/test1.sdp")
	startWs()
}
