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

import Axios from "axios";
import Notify from "common/notify";
import {$gettext} from "./vm";

const testConfig = {"jsHash":"48019917", "cssHash":"2b327230", "version": "test"};
const config = window.__CONFIG__ ? window.__CONFIG__ : testConfig;

const Api = Axios.create({
    baseURL: "/api/v1",
    headers: {common: {
        "X-Session-ID": window.localStorage.getItem("session_id"),
        "X-Client-Hash": config.jsHash,
        "X-Client-Version": config.version,
    }},
});

Api.interceptors.request.use(function (config) {
    // Do something before request is sent
    Notify.ajaxStart();
    return config;
}, function (error) {
    // Do something with request error
    return Promise.reject(error);
});

Api.interceptors.response.use(function (response) {
    Notify.ajaxEnd();

    if(typeof response.data == "string") {
        Notify.error($gettext("Request failed - invalid response"));
        console.warn("WARNING: Server returned HTML instead of JSON - API not implemented?");
    }

    return response;
}, function (error) {
    Notify.ajaxEnd();

    if (Axios.isCancel(error)) {
        return Promise.reject(error);
    }

    if(console && console.log) {
        console.log(error);
    }

    let errorMessage = $gettext("An error occurred - are you offline?");
    let code = error.code;

    if(error.response && error.response.data) {
        let data = error.response.data;
        code = data.code;
        errorMessage = data.message ? data.message : data.error;
    }

    if (code === 401) {
        Notify.logout(errorMessage);
    } else {
        Notify.error(errorMessage);
    }

    return Promise.reject(error);
});

export default Api;
