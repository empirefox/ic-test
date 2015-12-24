/*
 *  Copyright (c) 2015 The WebRTC project authors. All Rights Reserved.
 *
 *  Use of this source code is governed by a BSD-style license
 *  that can be found in the LICENSE file in the root of the source
 *  tree.
 */

/*jshint latedef: nofunc */
'use strict';

var startButton = document.getElementById('startButton');
var callButton = document.getElementById('callButton');
var hangupButton = document.getElementById('hangupButton');
callButton.disabled = true;
hangupButton.disabled = true;
startButton.onclick = start;
callButton.onclick = call;
hangupButton.onclick = hangup;

var startTime;
var localVideo = document.getElementById('localVideo');
var remoteVideo = document.getElementById('remoteVideo');

localVideo.addEventListener('loadedmetadata', function() {
  trace('Local video videoWidth: ' + this.videoWidth +
    'px,  videoHeight: ' + this.videoHeight + 'px');
});

remoteVideo.addEventListener('loadedmetadata', function() {
  trace('Remote video videoWidth: ' + this.videoWidth +
    'px,  videoHeight: ' + this.videoHeight + 'px');
});

remoteVideo.onresize = function() {
  trace('Remote video size changed to ' +
    remoteVideo.videoWidth + 'x' + remoteVideo.videoHeight);
  // We'll use the first onsize callback as an indication that video has started
  // playing out.
  if (startTime) {
    var elapsedTime = window.performance.now() - startTime;
    trace('Setup time: ' + elapsedTime.toFixed(3) + 'ms');
    startTime = null;
  }
};

var pc1;
var ws;
var offerOptions = {
  offerToReceiveAudio: 0,
  offerToReceiveVideo: 1
};

function start() {
  trace('Connecting to ws server');
  startButton.disabled = true;

  ws = new WebSocket('ws://127.0.0.1:9999/many');
  ws.onopen = function() {
    trace('Connected to ws server');
    callButton.disabled = false;
  };
  ws.onclose = function() {
    trace("Ws Connection closed");
    startButton.disabled = false;
    ws = null;
  };
  ws.onerror = function(e) {
    trace('Ws Connection error:' + e.toString());
    startButton.disabled = false;
  };
  ws.onmessage = function(e) {
    trace("ONE ->MANY: " + e.data);
    var signal;
    try {
      signal = JSON.parse(e.data);
    } catch (e) {
      trace('Failed to parse WSS message: ' + e.data);
      return;
    }
    if (signal.error) {
      trace('Signaling server error message: ' + signal.error);
      return;
    }
    switch (signal.type) {
      case 'answer':
        gotAnwser(signal);
        break;
      case 'candidate':
        gotIceCandidate(signal);
        break;
      default:
        trace('WARNING: unexpected message: ' + e.data);
    }
  };
  ws.signaling = function(message) {
    trace("MANY->One : " + JSON.stringify(message));
    ws.send(JSON.stringify(message));
  };
}

function call() {
  callButton.disabled = true;
  hangupButton.disabled = false;
  trace('Starting call');
  startTime = window.performance.now();

  var peerConnectionConfig = {
    iceServers: [{
      urls: [
        "stun:stun3.l.google.com:19302", "stun:stun.ideasip.com", "stun:stun4.l.google.com:19302",
        "stun:stun2.l.google.com:19302", "stun:stun1.l.google.com:19302", "stun:stun.ekiga.net",
        "stun:stun.schlund.de", "stun:stun.voipstunt.com", "stun:stun.voiparound.com",
        "stun:stun.voipbuster.com", "stun:stun.voxgratia.org"
      ]
    }]
  };
  pc1 = new RTCPeerConnection(peerConnectionConfig);
  trace('Created local peer connection object pc1');

  pc1.onicecandidate = onIceCandidate;
  pc1.oniceconnectionstatechange = onIceStateChange;

  pc1.onaddstream = gotRemoteStream;

  trace('pc1 createOffer start');
  pc1.createOffer(onCreateOfferSuccess, onCreateSessionDescriptionError, offerOptions);
}

function onCreateSessionDescriptionError(error) {
  trace('Failed to create session description: ' + error.toString());
}

function onCreateOfferSuccess(desc) {
  trace('Offer from pc1\n' + desc.sdp);
  trace('pc1 setLocalDescription start');
  pc1.setLocalDescription(desc, onSetLocalSuccess, onSetSessionDescriptionError);
  if (ws) {
    ws.signaling({
      sdp: desc.sdp,
      type: desc.type
    });
  }
}

function onSetLocalSuccess() {
  trace('pc1 setLocalDescription complete');
}

function onSetRemoteSuccess() {
  trace('pc1 setRemoteDescription complete');
}

function onSetSessionDescriptionError(error) {
  trace('Failed to set session description: ' + error.toString());
}

function gotRemoteStream(e) {
  remoteVideo.srcObject = e.stream;
  trace('pc1 received remote stream');
}

function gotAnwser(signal) {
  trace('Answer from pc2:\n' + signal.sdp);
  trace('pc1 setRemoteDescription start');
  pc1.setRemoteDescription(new RTCSessionDescription(signal), onSetRemoteSuccess, onSetSessionDescriptionError);
}

function onIceCandidate(event) {
  if (event.candidate) {
    var message = {
      type: 'candidate',
      label: event.candidate.sdpMLineIndex,
      id: event.candidate.sdpMid,
      candidate: event.candidate.candidate
    };
    if (ws) {
      ws.signaling(message);
    }
    trace('pc1 ICE candidate: \n' + event.candidate);
  }
}

function gotIceCandidate(signal) {
  pc1.addIceCandidate(new RTCIceCandidate({
    sdpMLineIndex: signal.label,
    candidate: signal.candidate
  }), onAddIceCandidateSuccess, onAddIceCandidateError);
  trace('remote ICE candidate: \n' + signal);
}

function onAddIceCandidateSuccess() {
  trace('pc1 addIceCandidate success');
}

function onAddIceCandidateError(error) {
  trace('pc1 failed to add ICE Candidate: ' + error.toString());
}

function onIceStateChange(event) {
  if (pc1) {
    trace('pc1 ICE state: ' + pc1.iceConnectionState);
    console.log('ICE state change event: ', event);
  }
}

function hangup() {
  trace('Ending call');
  pc1.close();
  pc1 = null;
  if (ws) {
    ws.signaling({
      type: 'bye'
    });
    ws.close();
  }
  hangupButton.disabled = true;
  callButton.disabled = false;
}
