import Event from "pubsub-js";

const Alert = {
    info: function (message) {
        Event.publish("alert.info", message);
    },
    warning: function (message) {
        Event.publish("alert.warning", message);
    },
    error: function (message) {
        Event.publish("alert.error", message);
    },
    success: function (message) {
        Event.publish("alert.success", message);
    },
};

export default Alert;