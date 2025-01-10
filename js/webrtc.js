// WebRTC.js
class WebRTC {
    constructor() {

        this.peerConnection = new RTCPeerConnection({
            iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
        });
        this.isRemoteDescriptionSet = false;
        this.pendingCandidates = [];


    }

    async setRemoteDescription(description) {
        try {
            await this.peerConnection.setRemoteDescription(new RTCSessionDescription(description));
        } catch (error) {
            console.error('Error setting remote description:', error);
        }
    }

    addCandidate(candidate) {
        this.pendingCandidates.push(candidate);
    }

    createAnswer() {
        return this.peerConnection.createAnswer();
    }

    setLocalDescription(answer) {
        return this.peerConnection.setLocalDescription(answer);
    }

}

export default WebRTC;
