<template>
  <div class="p-page p-page-upgrade" tabindex="1">
    <v-toolbar
      flat
      color="secondary"
      :density="$vuetify.display.smAndDown ? 'compact' : 'default'"
      class="p-page__navigation"
    >
      <v-toolbar-title>
        {{ $gettext(`Membership`) }}
        <v-icon :icon="rtl ? 'mdi-chevron-left' : 'mdi-chevron-right'"></v-icon>
        <span v-if="busy">
          {{ $gettext(`Busy, please wait…`) }}
        </span>
        <span v-else-if="success">
          {{ $gettext(`Successfully Connected`) }}
        </span>
        <span v-else-if="error">
          {{ $gettext(`Error`) }}
        </span>
        <span v-else>
          {{ $gettext(`Upgrade`) }}
        </span>
      </v-toolbar-title>

      <v-btn icon :href="links.compare" target="_blank" class="action-upgrade" :title="$gettext('Learn more')">
        <v-icon size="26" color="surface-variant">mdi-diamond-stone</v-icon>
      </v-btn>
    </v-toolbar>
    <div class="pa-6">
      <v-form ref="form" v-model="valid" autocomplete="off" validate-on="invalid-input" @submit.prevent>
        <div v-if="busy">
          <v-progress-linear :indeterminate="true" color="surface-variant"></v-progress-linear>
        </div>
        <div v-else-if="error">
          <v-alert color="primary" icon="mdi-connection" variant="outlined">
            {{ error ? error + "." : $gettext("Failed to connect account.") }}
          </v-alert>
          <div class="action-buttons">
            <v-btn color="primary" :block="$vuetify.display.xs" variant="outlined" :disabled="busy" @click.stop="reset">
              {{ $gettext(`Cancel`) }}
            </v-btn>
            <v-btn
              color="highlight"
              :block="$vuetify.display.xs"
              :href="links.contact"
              target="_blank"
              variant="flat"
              class="action-contact"
            >
              {{ $gettext(`Contact Us`) }}
            </v-btn>
          </div>
        </div>
        <div v-else-if="success">
          <v-alert color="primary" icon="mdi-check-decagram" variant="outlined">
            {{ $gettext(`Your account has been successfully connected.`) }}
            <span v-if="$config.values.restart">
              {{ $gettext(`Please restart your instance for the changes to take effect.`) }}
            </span>
          </v-alert>

          <div class="action-buttons">
            <v-btn
              href="https://my.photoprism.app/dashboard"
              target="_blank"
              color="primary"
              :block="$vuetify.display.xs"
              variant="outlined"
              class="action-manage"
              :disabled="busy"
            >
              {{ $gettext(`Manage Account`) }}
            </v-btn>
            <v-btn
              v-if="$config.values.restart && !$config.values.disable.restart"
              color="highlight"
              :block="$vuetify.display.xs"
              variant="flat"
              :disabled="busy"
              class="px-5 action-restart"
              @click.stop.p.prevent="onRestart"
            >
              {{ $gettext(`Restart`) }}
              <v-icon end>mdi-restart</v-icon>
            </v-btn>
            <v-btn
              v-if="$config.getTier() < 4"
              href="https://my.photoprism.app/dashboard/membership"
              target="_blank"
              color="highlight"
              :block="$vuetify.display.xs"
              variant="flat"
              class="px-5 action-upgrade"
              :disabled="busy"
            >
              {{ $gettext(`Upgrade Now`) }}
              <v-icon :icon="rtl ? 'mdi-chevron-left' : 'mdi-chevron-right'" size="20" end></v-icon>
            </v-btn>
          </div>
        </div>
        <div v-else>
          <div v-if="$config.getTier() < 4" class="pb-6 text-subtitle-2 text-break text-selectable">
            {{ $gettext(`Become a member today, support our mission and enjoy our member benefits!`) }}
            {{
              $gettext(
                `Your continued support helps us provide regular updates and remain independent, so we can fulfill our mission and protect your privacy.`
              )
            }}
          </div>

          <v-alert color="primary" variant="outlined">
            <p class="text-body-2 text-break text-selectable">
              <strong>{{
                $gettext(
                  `To upgrade, you can either enter an activation code or click "Register" to sign up on our website:`
                )
              }}</strong>
            </p>

            <v-text-field
              v-model="form.token"
              single-line
              hide-details
              autocomplete="off"
              :placeholder="$gettext('Activation Code')"
            ></v-text-field>

            <div class="action-buttons">
              <v-btn
                v-if="$config.getTier() >= 4"
                href="https://my.photoprism.app/dashboard"
                target="_blank"
                color="primary"
                :block="$vuetify.display.xs"
                variant="outlined"
                class="action-manage"
                :disabled="busy"
              >
                {{ $gettext(`Manage Account`) }}
              </v-btn>
              <v-btn
                v-else
                color="primary"
                :block="$vuetify.display.xs"
                variant="outlined"
                :disabled="busy"
                class="action-compare"
                @click.stop="compare"
              >
                {{ $gettext(`Compare Editions`) }}
              </v-btn>

              <v-btn
                v-if="!form.token.length"
                color="highlight"
                :block="$vuetify.display.xs"
                variant="flat"
                :disabled="busy"
                class="px-5 action-proceed"
                @click.stop="connect"
              >
                {{ $gettext(`Register`) }}
                <v-icon :icon="rtl ? 'mdi-chevron-left' : 'mdi-chevron-right'" size="20" end></v-icon>
              </v-btn>
              <v-btn
                v-else
                color="highlight"
                :block="$vuetify.display.xs"
                variant="flat"
                :disabled="busy || form.token.length !== tokenMask.length"
                class="px-5 action-activate"
                @click.stop="activate"
              >
                {{ $gettext(`Activate`) }}
                <v-icon :icon="rtl ? 'mdi-chevron-left' : 'mdi-chevron-right'" end></v-icon>
              </v-btn>
            </div>
          </v-alert>

          <div class="pt-6 text-caption text-break text-selectable">
            {{
              $gettext(
                `You are welcome to contact us at membership@photoprism.app for questions regarding your membership.`
              )
            }}
            {{
              $gettext(
                `By using the software and services we provide, you agree to our terms of service, privacy policy, and code of conduct.`
              )
            }}
          </div>
        </div>
      </v-form>
    </div>
    <p-about-footer></p-about-footer>
  </div>
</template>

<script>
import * as options from "options/options";
import $api from "common/api";
import { restart } from "common/server";
import links from "common/links";

import PAboutFooter from "component/about/footer.vue";

export default {
  name: "PPageConnect",
  components: {
    PAboutFooter,
  },
  data() {
    const token = this.$route.params.token ? this.$route.params.token : "";
    const membership = this.$config.getMembership();

    return {
      links,
      success: false,
      busy: false,
      valid: false,
      error: "",
      options: options,
      isPublic: this.$config.isPublic(),
      isAdmin: this.$session.isAdmin(),
      isDemo: this.$config.isDemo(),
      isSponsor: this.$config.isSponsor(),
      tier: this.$config.getTier(),
      membership: membership,
      showInfo: !token && membership === "ce",
      rtl: this.$isRtl,
      tokenMask: "nnnn-nnnn-nnnn",
      form: {
        token,
      },
    };
  },
  created() {
    this.$config.load().then(() => {
      if (this.$config.isPublic() || !this.$session.isSuperAdmin()) {
        this.$router.push({ name: "home" });
      }
    });
  },
  mounted() {
    this.$view.enter(this);
  },
  unmounted() {
    this.$view.leave(this);
  },
  methods: {
    onRestart() {
      restart(this.$router.resolve({ name: "about" }).href);
    },
    reset() {
      this.success = false;
      this.busy = false;
      this.error = "";
    },
    compare() {
      window.open(links.compare, "_blank").focus();
    },
    connect() {
      this.$view.redirect(links.connect + encodeURIComponent(window.location));
    },
    activate() {
      if (!this.form.token || this.form.token.length !== this.tokenMask.length) {
        return;
      }

      const values = { Token: this.form.token };

      if (values.Token.length >= 4) {
        this.busy = true;
        this.$notify.blockUI();
        $api
          .put("connect/hub", values)
          .then(() => {
            this.$notify.success(this.$gettext("Connected"));
            this.success = true;
            this.busy = false;
            this.$config.update();
          })
          .catch((error) => {
            this.busy = false;
            if (error.response && error.response.data) {
              let data = error.response.data;
              this.error = data.message ? data.message : data.error;
            }

            if (!this.error) {
              this.error = this.$gettext("Invalid parameters");
            }
          })
          .finally(() => {
            this.$notify.unblockUI();
          });
      } else {
        this.$notify.error(this.$gettext("Invalid parameters"));
        this.$router.push({ name: "upgrade" });
      }
    },
    getMembership() {
      const m = this.$config.getMembership();
      switch (m) {
        case "":
        case "ce":
          return "Community";
        case "cloud":
          return "Cloud";
        case "essentials":
          return "Essentials";
        default:
          return "Plus";
      }
    },
  },
};
</script>
