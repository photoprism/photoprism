import Event from "pubsub-js";

class Log {
    constructor() {
        this.cap = 100;
        this.created = new Date;
        this.logs = [
            /* EXAMPLE LOG MESSAGE
            {
                "msg": "waiting for events",
                "level": "debug",
                "time": this.created.toISOString(),
            },
            */
        ];

        this.logId = 0;

        Event.subscribe("log", this.onLog.bind(this));
    }

    onLog(ev, data) {
        data.id = this.logId++;

        this.logs.unshift(data);

        if(this.logs.length > this.cap) {
            this.logs.splice(this.cap);
        }
    }
}

const log = new Log;

export default log;
