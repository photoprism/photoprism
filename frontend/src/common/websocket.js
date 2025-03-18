/*

Copyright (c) 2018 - 2025 PhotoPrism UG. All rights reserved.

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

import Sockette from "sockette";
import $event from "common/event";
import { $config } from "app/session";

const host = window.location.host;
const protocol = "https:" === document.location.protocol ? "wss://" : "ws://";
const apiUri = window.__CONFIG__ ? window.__CONFIG__.apiUri : "/api/v1";
const socketUrl = protocol + host + apiUri + "/ws";

const Socket = new Sockette(socketUrl, {
  timeout: 5e3,
  onopen: (e) => {
    console.log("websocket: connected");
    $config.disconnected.value = false;
    document.body.classList.remove("disconnected");
    $event.publish("websocket.connected", e);
  },
  onmessage: (e) => {
    const m = JSON.parse(e.data);
    $event.publish(m.event, m.data);
  },
  onreconnect: () => console.log("websocket: reconnecting"),
  onmaximum: () => console.warn("websocket: hit max reconnect limit"),
  onclose: () => {
    console.warn("websocket: disconnected");
    $config.disconnected.value = true;
    document.body.classList.add("disconnected");
  },
});

export default Socket;
