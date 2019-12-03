import Sockette from "sockette";
import Event from "pubsub-js";
import randomString from "crypto-random-string";

export const token = randomString({length: 16});
const host = window.location.host;
const prot = ('https:' === document.location.protocol ? 'wss://' : 'ws://');
const url = prot + host + "/api/v1/ws";
const clientInfo = {
    "token": token,
    "hash": window.clientConfig.jsHash,
    "version": window.clientConfig.version,
};

const Socket = new Sockette(url, {
    timeout: 5e3,
    onopen: e => {
        console.log('websocket: connected', e);
        Socket.send(JSON.stringify(clientInfo));
    },
    onmessage: e => {
        const m = JSON.parse(e.data);
        Event.publish(m.event, m.data);
    },
    onreconnect: e => console.log('websocket: reconnecting', e),
    onmaximum: e => console.warn('websocket: hit max reconnect limit', e),
    onclose: e => console.log('websocket: closed', e),
    onerror: e => console.log('websocket: error', e)
});

export default Socket;
