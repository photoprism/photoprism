<template>
  <v-snackbar
    id="p-notify"
    v-model="visible"
    :color="color"
    :timeout="0"
    :class="textColor"
    :bottom="true"
  >
    <span :dir="!rtl ? 'let' : 'rtl'">{{ text }}</span>
    <v-btn :class="textColor + ' pr-0'" icon flat @click="close">
      <v-icon :class="textColor">close</v-icon>
    </v-btn>
  </v-snackbar>
</template>
<script>
import Event from "pubsub-js";

export default {
  name: "PNotify",
  data() {
    return {
      text: "",
      color: "primary",
      textColor: "",
      visible: false,
      messages: [],
      lastMessageId: 1,
      lastMessage: "",
      subscriptionId: "",
      rtl: this.$rtl,
    };
  },
  created() {
    this.subscriptionId = Event.subscribe("notify", this.onNotify);
  },
  destroyed() {
    Event.unsubscribe(this.subscriptionId);
  },
  methods: {
    onNotify: function (ev, data) {
      const type = ev.split(".")[1];

      // Get the message.
      let m = data.message;

      // Skip empty messages.
      if (!m || !m.length) {
        console.warn("notify: empty message");
        return;
      }

      // Log notifications in test mode.
      if (this.$config.test) {
        console.log(type + ": " + m.toLowerCase());
        return;
      }

      // First letter of the message should be uppercase.
      m = m.replace(/^./, m[0].toUpperCase());

      switch (type) {
        case "warning":
          this.addWarningMessage(m);
          break;
        case "error":
          this.addErrorMessage(m);
          break;
        case "success":
          this.addSuccessMessage(m);
          break;
        case "info":
          this.addInfoMessage(m);
          break;
        default:
          alert(m);
      }
    },

    addWarningMessage: function (message) {
      this.addMessage("warning", "black--text", message, 3000);
    },

    addErrorMessage: function (message) {
      this.addMessage("error", "white--text", message, 8000);
    },

    addSuccessMessage: function (message) {
      this.addMessage("success", "white--text", message, 2000);
    },

    addInfoMessage: function (message) {
      this.addMessage("info", "white--text", message, 2000);
    },

    addMessage: function (color, textColor, message, delay) {
      if (message === this.lastMessage) return;

      this.lastMessageId++;
      this.lastMessage = message;

      const m = {
        id: this.lastMessageId,
        color: color,
        textColor: textColor,
        delay: delay,
        message: message,
      };

      this.messages.push(m);

      if (!this.visible) {
        this.show();
      }
    },

    close: function () {
      this.visible = false;
      this.show();
    },
    show: function () {
      const message = this.messages.shift();

      if (message) {
        this.text = message.message;
        this.color = message.color;
        this.textColor = message.textColor;
        this.visible = true;

        setTimeout(() => {
          this.lastMessage = "";
          this.show();
        }, message.delay);
      } else {
        this.visible = false;
        this.text = "";
      }
    },
  },
};
</script>
