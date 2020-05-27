import Axios from "axios";
import Notify from "common/notify";

const testConfig = {"jsHash":"48019917", "cssHash":"2b327230", "version": "test"};
const config = window.__CONFIG__ ? window.__CONFIG__ : testConfig;

const Api = Axios.create({
    baseURL: "/api/v1",
    headers: {common: {
        "X-Session-Token": window.localStorage.getItem("session_token"),
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
        Notify.error("Request failed - invalid response");
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

    let errorMessage = "An error occurred - are you offline?";
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
