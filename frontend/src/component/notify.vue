<template>
  <teleport to="body">
    <transition name="fade-transition">
      <div v-if="visible" id="p-notify" tabindex="-1">
        <div
          :class="'p-notify--' + message.color"
          class="v-snackbar v-snackbar--bottom v-snackbar--center p-notify"
          role="alert"
          tabindex="-1"
          @click.stop.prevent="showNext"
        >
          <div class="v-snackbar__wrapper v-snackbar--variant-flat">
            <span class="v-snackbar__underlay"></span>
            <div tabindex="-1" class="v-snackbar__content">
              <i
                v-if="message.icon"
                :class="['text-' + message.color, 'mdi-' + message.icon]"
                class="mdi v-icon notranslate p-notify__icon"
                aria-hidden="true"
              ></i>
              <div class="p-notify__text">
                {{ message.text }}
              </div>
              <i
                :class="'text-on-' + message.color"
                class="mdi-close mdi v-icon notranslate p-notify__close"
                aria-hidden="true"
              ></i>
            </div>
          </div>
        </div>
      </div>
    </transition>
  </teleport>
</template>
<script>
export default {
  name: "PNotify",
  data() {
    return {
      visible: false,
      snackbar: true,
      messages: [],
      message: {
        icon: "",
        color: "transparent",
        text: "",
        delay: this.defaultDelay,
      },
      lastText: "",
      lastId: 1,
      subscriptionId: "",
      defaultColor: "info",
      defaultDelay: 2000,
      warningDelay: 3000,
      errorDelay: 8000,
    };
  },
  created() {
    this.subscriptionId = this.$event.subscribe("notify", this.onNotify);
  },
  beforeUnmount() {
    this.messages = [];
    this.visible = false;
    this.$event.unsubscribe(this.subscriptionId);
  },
  methods: {
    onNotify(ev, data) {
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

    addSuccessMessage(message) {
      this.addMessage("success", "check-circle", message, this.defaultDelay);
    },

    addInfoMessage(message) {
      this.addMessage("info", "information-outline", message, this.defaultDelay);
    },

    addWarningMessage(message) {
      this.addMessage("warning", "alert", message, this.warningDelay);
    },

    addErrorMessage(message) {
      this.addMessage("error", "alert-circle-outline", message, this.errorDelay);
    },

    addMessage(color, icon, text, delay) {
      if (!text || text === this.lastText) {
        return;
      }

      this.lastId++;
      this.lastText = text;

      const m = {
        id: this.lastId,
        color,
        icon,
        text,
        delay,
      };

      this.messages.push(m);

      if (!this.visible) {
        this.showNext();
      }
    },
    onSnackbar(show) {
      if (show) {
        this.snackbar = true;
      } else {
        this.snackbar = false;
        this.showNext();
      }
    },
    showNext() {
      const message = this.messages.shift();

      if (message) {
        this.message = message;

        if (!this.message.icon) {
          this.message.icon = "";
        }

        if (!this.message.color) {
          this.message.color = this.defaultColor;
        }

        if (!this.message.delay || this.message.delay <= 0) {
          this.message.delay = this.defaultDelay;
        }

        if (!this.snackbar) {
          this.snackbar = true;
        }

        this.visible = true;

        setTimeout(() => {
          this.lastText = "";
          this.showNext();
        }, this.message.delay);
      } else {
        this.lastText = "";
        this.visible = false;
        this.message.text = "";
      }
    },
  },
};
</script>
