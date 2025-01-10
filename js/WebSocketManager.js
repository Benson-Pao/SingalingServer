// WebSocketManager.js
class WebSocketManager {
    constructor(domain, token) {
        this.ws = null;
        this.wsurl = `ws://${domain}/ws`;
        this.token = token;
        this.pendingCandidates = [];
        this.eventHandlers = {
            offer: [],
            answer: [],
            candidate: []
        };
    }

    connect() {
        this.ws = new WebSocket(`${this.wsurl}?token=${this.token}`);
        this.ws.onopen = () => {
            console.log("WebSocket connected");
        };
        this.ws.onmessage = this.onMessage.bind(this);
        this.ws.onerror = (error) => console.error("WebSocket Error:", error);
        this.ws.onclose = () => console.log("WebSocket closed");
    }

    onMessage(message) {
        const data = JSON.parse(message.data);
        console.log("Received WebSocket message:", data);
        if (this.eventHandlers[data.type]) {
            this.eventHandlers[data.type].forEach(handler => handler(data));
        }
    }

    sendSignal(type, data) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({ type, token: this.token, ...data }));
        } else {
            console.error("WebSocket is not open, cannot send message");
        }
    }

    addEventListener(type, handler) {
        if (this.eventHandlers[type]) {
            this.eventHandlers[type].push(handler);
        }
    }
}
export default WebSocketManager;
