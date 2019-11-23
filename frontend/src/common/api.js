import "@babel/polyfill/noConflict";
import Axios from "axios";
import Notify from "common/notify";

const Api = Axios.create({
    baseURL: "/api/v1",
    headers: {common: {
        "X-Session-Token": window.localStorage.getItem("session_token"),
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
