import Sockette from "sockette";
import Event from "pubsub-js";

const host = window.location.host;
const prot = ('https:' === document.location.protocol ? 'wss://' : 'ws://');

const Socket = new Sockette(prot + host + "/api/v1/ws", {
    timeout: 5e3,
    onopen: e => {
        // console.log('Websocket connected', e);
        Socket.send("hello world");
    },
    onmessage: e => {
        const m = JSON.parse(e.data);
        // console.log('Websocket data received', m);
        Event.publish(m.event, m.data);
    },
    onreconnect: e => console.log('Websocket reconnecting', e),
    onmaximum: e => console.warn('Websocket max reconnect limit', e),
    onclose: e => console.log('Websocket closed', e),
    onerror: e => console.log('Websocket error', e)
});

export default Socket;
