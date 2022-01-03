/*

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.app>

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

    PhotoPrism® is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.app/developer-guide/

*/

import Sockette from "sockette";
import Event from "pubsub-js";
import { config } from "../session";

const host = window.location.host;
const prot = "https:" === document.location.protocol ? "wss://" : "ws://";
const apiUri = window.__CONFIG__ ? window.__CONFIG__.apiUri : "/api/v1";
const url = prot + host + apiUri + "/ws";

const Socket = new Sockette(url, {
  timeout: 5e3,
  onopen: (e) => {
    console.log("websocket: connected");
    config.disconnected = false;
    document.body.classList.remove("disconnected");
    Event.publish("websocket.connected", e);
  },
  onmessage: (e) => {
    const m = JSON.parse(e.data);
    Event.publish(m.event, m.data);
  },
  onreconnect: () => console.log("websocket: reconnecting"),
  onmaximum: () => console.warn("websocket: hit max reconnect limit"),
  onclose: () => {
    console.warn("websocket: disconnected");
    config.disconnected = true;
    document.body.classList.add("disconnected");
  },
});

export default Socket;
