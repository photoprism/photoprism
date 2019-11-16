import Event from "pubsub-js";

const Alert = {
    info: function (message) {
        Event.publish("alert.info", {msg: message});
    },
    warning: function (message) {
        Event.publish("alert.warning", {msg: message});
    },
    error: function (message) {
        Event.publish("alert.error", {msg: message});
    },
    success: function (message) {
        Event.publish("alert.success", {msg: message});
    },
};

export default Alert;
