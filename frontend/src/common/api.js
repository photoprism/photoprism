import axios from 'axios';
import Event from 'pubsub-js';
import 'babel-polyfill';

const Api = axios.create({
    baseURL: '/api/v1',
    headers: {common: {
        'X-Session-Token': window.localStorage.getItem('session_token'),
    }},
});

Api.interceptors.request.use(function (config) {
    // Do something before request is sent
    Event.publish('ajax.start', config);
    return config;
}, function (error) {
    // Do something with request error
    return Promise.reject(error);
});

Api.interceptors.response.use(function (response) {
    Event.publish('ajax.end', response);

    return response;
}, function (error) {
    if(console && console.log) {
        console.log(error);
    }

    const data = error.response.data;

    Event.publish('ajax.end');
    Event.publish('alert.error', data.message ? data.message : data.error);

    if(data.code === 401) {
        window.location = '/';
    }

    return Promise.reject(error);
});

export default Api;
