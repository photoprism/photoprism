<template>
  <v-container id="auth-consent" theme="login" fluid fill-height class="auth-login auth-consent wallpaper background-welcome pa-6" :style="wallpaper()">
    <v-theme-provider theme="login">
      <v-row id="auth-layout" class="auth-layout">
        <v-col cols="12" sm="9" md="6" lg="5" xl="3">
          <v-form ref="form" class="auth-login-form" accept-charset="UTF-8" @submit.prevent="onApprove">
            <v-card id="auth-login-box" class="elevation-12 auth-login-box pa-1 blur-7">
              <v-card-text>
                <p-auth-header></p-auth-header>
                <v-row class="align-start" density="compact">
                  <v-col v-if="loading" cols="12" class="py-8 d-flex justify-center">
                    <v-progress-circular indeterminate color="highlight"></v-progress-circular>
                  </v-col>
                  <v-col v-else-if="!request" cols="12" class="text-subtitle-1 text-center py-4 consent-not-found">
                    {{ $gettext(`This authorization request has expired or is invalid. Please return to the app and try again.`) }}
                  </v-col>
                  <template v-else>
                    <v-col cols="12" class="text-subtitle-1 text-center pb-0 consent-title">
                      <template v-if="request.available">
                        {{ $gettext(`%{client} wants to access your account %{account} on %{site}.`, consentParams) }}
                      </template>
                      <template v-else>
                        {{ $gettext(`%{client} cannot be connected to your account %{account} on %{site}.`, consentParams) }}
                      </template>
                    </v-col>
                    <v-col cols="12" class="text-body-2 text-center text-secondary pt-1 pb-2 consent-origin">
                      {{ request.client.origin }}
                    </v-col>
                    <v-col v-if="request.available" cols="12" class="text-body-2 text-center text-secondary pt-0 consent-hint">
                      {{ $gettext(`Only allow access if you started connecting this app yourself.`) }}
                    </v-col>
                    <v-col v-if="!request.available" cols="12" class="text-body-2 text-center consent-unavailable">
                      {{ $gettext(`Your account does not have the permissions this app requires.`) }}
                    </v-col>
                    <v-col v-else cols="12" class="pb-0 consent-scopes">
                      <v-checkbox
                        v-for="scope in request.scopes"
                        :key="scope.name"
                        v-model="selected"
                        :value="scope.name"
                        :label="scopeLabel(scope)"
                        :disabled="!scope.available || scope.name === requiredScope"
                        :class="`consent-scope consent-scope-${scope.name}`"
                        color="highlight"
                        density="compact"
                        hide-details
                      ></v-checkbox>
                    </v-col>
                  </template>
                  <v-col v-if="!loading" cols="12" class="auth-actions pt-3">
                    <div class="action-buttons auth-buttons d-flex justify-center">
                      <v-btn
                        v-if="request"
                        :disabled="busy"
                        :block="$vuetify.display.xs"
                        color="highlight"
                        variant="outlined"
                        class="action-deny"
                        @click.stop.prevent="onDeny"
                      >
                        {{ request.available ? $gettext(`Deny`) : $gettext(`Back to App`) }}
                      </v-btn>
                      <v-btn
                        v-if="request && request.available"
                        :disabled="busy || !canApprove"
                        :block="$vuetify.display.xs"
                        color="highlight"
                        variant="flat"
                        class="action-approve"
                        @click.stop.prevent="onApprove"
                      >
                        {{ $gettext(`Allow`) }}
                      </v-btn>
                    </div>
                  </v-col>
                </v-row>
              </v-card-text>
            </v-card>
          </v-form>
        </v-col>
      </v-row>
      <p-auth-footer></p-auth-footer>
    </v-theme-provider>
  </v-container>
</template>

<script>
import $api from "common/api";
import PAuthHeader from "component/auth/header.vue";
import PAuthFooter from "component/auth/footer.vue";

// redirectAllowed reports whether a redirect URL returned by the server may be followed: a web URL or the
// private-use URI scheme of a native app, never a script or data URL.
export function redirectAllowed(url) {
  if (!url || typeof url !== "string") {
    return false;
  }

  try {
    const scheme = new URL(url).protocol.toLowerCase();
    return scheme === "https:" || scheme === "http:" || /^[a-z][a-z0-9+-]*(\.[a-z0-9+-]+)+:$/.test(scheme);
  } catch {
    return false;
  }
}

// PPageConsent lets the signed-in user approve or deny an app's request to access their account.
export default {
  name: "PPageConsent",
  components: {
    PAuthHeader,
    PAuthFooter,
  },
  data() {
    return {
      loading: true,
      busy: false,
      request: null,
      selected: [],
      requiredScope: "mcp",
      wallpaperUri: this.$config.values.wallpaperUri,
    };
  },
  computed: {
    requestId() {
      return typeof this.$route?.query?.request === "string" ? this.$route.query.request : "";
    },
    clientName() {
      return this.request?.client?.name || this.request?.client?.origin || this.$gettext("Unknown App");
    },
    accountName() {
      return this.request?.account?.displayName || this.request?.account?.name || "";
    },
    siteHost() {
      try {
        return new URL(this.request?.siteUrl).host;
      } catch {
        return this.request?.siteUrl || "";
      }
    },
    consentParams() {
      return { client: this.clientName, account: this.accountName, site: this.siteHost };
    },
    canApprove() {
      return this.selected.includes(this.requiredScope);
    },
  },
  created() {
    this.load();
  },
  mounted() {
    this.$view.enter(this);
  },
  unmounted() {
    this.$view.leave(this);
  },
  methods: {
    wallpaper() {
      if (this.wallpaperUri) {
        return `background-image: url(${this.wallpaperUri});`;
      }

      return "";
    },
    // scopeLabel returns the plain-language description of a requested scope.
    scopeLabel(scope) {
      const labels = {
        mcp: this.$gettext("Use PhotoPrism tools in this app"),
        photos: this.$gettext("Search and view your photos"),
        videos: this.$gettext("Search and view your videos"),
        albums: this.$gettext("View your albums"),
        favorites: this.$gettext("View your favorites"),
        moments: this.$gettext("View your moments"),
        calendar: this.$gettext("View your calendar"),
        folders: this.$gettext("View your folders"),
        places: this.$gettext("View your places"),
        labels: this.$gettext("View your labels"),
        people: this.$gettext("View people"),
        cameras: this.$gettext("View cameras"),
        lenses: this.$gettext("View lenses"),
        files: this.$gettext("View your files"),
      };

      const label = labels[scope.name] || scope.name;

      return scope.available ? label : `${label} (${this.$gettext("not available")})`;
    },
    // load fetches the pending authorization request and preselects the usable scopes.
    load() {
      if (!this.requestId) {
        this.loading = false;
        return;
      }

      $api
        .post("oauth/consent", { request: this.requestId })
        .then((resp) => {
          this.request = resp.data;
          this.selected = (resp.data?.scopes || []).filter((s) => s.available).map((s) => s.name);
        })
        .catch(() => {
          this.request = null;
        })
        .finally(() => {
          this.loading = false;
        });
    },
    // follow navigates to the app's redirect URL returned by the server.
    follow(resp) {
      const url = resp?.data?.redirectUri;

      if (redirectAllowed(url)) {
        this.$view.redirect(url, 0, false);
      } else {
        this.busy = false;
        this.$notify.error(this.$gettext("Something went wrong, try again"));
      }
    },
    // onError shows that the request has expired, or a generic error if it could not be decided.
    onError(err) {
      this.busy = false;

      if (err?.response?.status === 404) {
        this.request = null;
      } else {
        this.$notify.error(this.$gettext("Something went wrong, try again"));
      }
    },
    onApprove() {
      if (this.busy || !this.request?.available || !this.canApprove) {
        return;
      }

      this.busy = true;

      $api
        .post("oauth/consent/approve", { request: this.requestId, scopes: this.selected })
        .then((resp) => this.follow(resp))
        .catch((err) => this.onError(err));
    },
    onDeny() {
      if (this.busy || !this.request) {
        return;
      }

      this.busy = true;

      $api
        .post("oauth/consent/deny", { request: this.requestId })
        .then((resp) => this.follow(resp))
        .catch((err) => this.onError(err));
    },
  },
};
</script>
