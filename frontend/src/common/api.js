
import axios from "axios";
import Event from "pubsub-js";
import "@babel/polyfill";
import { router } from '../app.js'

const Api = axios.create({
    baseURL: "/api/v1",
});

Api.interceptors.request.use(function (config) {
    // Do something before request is sent
    Event.publish("ajax.start", config);
    return config;
}, function (error) {
    // Do something with request error
    return Promise.reject(error);
});

Api.interceptors.response.use(function (response) {
    Event.publish("ajax.end", response);

    return response;
}, function (error) {
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

    Event.publish("ajax.end");
    Event.publish("alert.error", errorMessage);

    if ((code === 401 || error.response.status === 401)
      && router.currentRoute.name !== "Login") {
      // the router "beforeAuth" didn't catch that (maybe there is an old
      // session key in the localStorage) so redirect to login ourselves
      router.push({ name: "Login" })
    }

    return Promise.reject(error);
});

export default Api;
