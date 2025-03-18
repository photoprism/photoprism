// True if debug and/or trace logs should be recorded.
const debug = window.__CONFIG__?.debug;
const trace = window.__CONFIG__?.trace;

// Use global variable to configure pubsub.js, a dependency-free publish/subscribe event hub:
// https://sahadar.github.io/pubsub/#installation
window.pubsub = {
  separator: ".",
  recurrent: true,
  async: true,
  log: trace,
};

// Import pubsub.js, see https://www.npmjs.com/package/pubsub-js.
import * as PubSub from "pubsub-js";

// Use $event as a generic alias for publishing and subscribing to events.
const $event = PubSub;

// Log all events when running in trace log mode, and log config events in debug mode.
// Event names are displayed in blue so that they are easy to recognize.
if (trace) {
  $event.subscribeAll((ev, data) => {
    if (data) {
      console.debug(`%c${ev}`, "background: transparent; color: #9FA8DA; font-weight: normal;", data);
    } else {
      console.debug(`%c${ev}`, "background: transparent; color: #9FA8DA; font-weight: normal;");
    }
  });
} else if (debug) {
  $event.subscribe("config", (ev, data) => {
    if (data) {
      console.debug(`%c${ev}`, "background: transparent; color: #9FA8DA; font-weight: normal;", data);
    } else {
      console.debug(`%c${ev}`, "background: transparent; color: #9FA8DA; font-weight: normal;");
    }
  });
}

// Export $event to publish and subscribe to events.
export default $event;
