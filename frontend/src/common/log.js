import Event from "pubsub-js";

class Log {
    constructor() {
        this.logs = [];
        this.logId = 0;

        Event.subscribe('log', this.onLog.bind(this));
    }

    onLog(ev, data) {
        data.id = this.logId++;
        this.logs.unshift(data);
    }
}

const log = new Log;

export default log;
