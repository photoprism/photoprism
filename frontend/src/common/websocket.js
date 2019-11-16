import Sockette from "sockette";
import Event from "pubsub-js";

const host = window.location.host;
const Socket = new Sockette("ws://" + host + "/api/v1/ws", {
    timeout: 5e3,
    onopen: e => {
        console.log('Connected!', e);
        Socket.send("hello world");
    },
    onmessage: e => {
        const m = JSON.parse(e.data);
        console.log('Received:', m);
        Event.publish(m.event, m.data);
    },
    onreconnect: e => console.log('Reconnecting...', e),
    onmaximum: e => console.log('Stop Attempting!', e),
    onclose: e => console.log('Closed!', e),
    onerror: e => console.log('Error:', e)
});

export default Socket;
