/*

Copyright (c) 2018 - 2020 Michael Mayer <hello@photoprism.org>

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

    PhotoPrism™ is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.org if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.org/developer-guide/

*/

import Event from "pubsub-js";

const Notify = {
    info: function (message) {
        Event.publish("notify.info", {msg: message});
    },
    warn: function (message) {
        Event.publish("notify.warning", {msg: message});
    },
    error: function (message) {
        Event.publish("notify.error", {msg: message});
    },
    success: function (message) {
        Event.publish("notify.success", {msg: message});
    },
    logout: function (message) {
        Event.publish("notify.error", {msg: message});
        Event.publish("session.logout", {msg: message});
    },
    ajaxStart: function() {
        Event.publish("ajax.start");
    },
    ajaxEnd: function() {
        Event.publish("ajax.end");
    },
    blockUI: function() {
        const el = document.getElementById("busy-overlay");

        if(el) {
            el.style.display = "block";
        }
    },
    unblockUI: function() {
        const el = document.getElementById("busy-overlay");

        if(el) {
            el.style.display = "none";
        }
    },
    wait: function () {
        this.warn("Busy, please wait...");
    },
};

export default Notify;
