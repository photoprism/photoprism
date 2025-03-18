<template>
  <v-dialog
    :model-value="visible"
    persistent
    max-width="610"
    class="p-dialog modal-dialog p-settings-apps"
    @keydown.esc="close"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-form
      ref="form"
      validate-on="invalid-input"
      class="form-password"
      accept-charset="UTF-8"
      tabindex="1"
      @submit.prevent
    >
      <v-card>
        <v-card-title class="d-flex justify-start align-center ga-3">
          <v-icon v-if="action === 'add'" size="28" color="primary">mdi-plus</v-icon>
          <v-icon v-else-if="action === 'copy'" size="28" color="primary">mdi-shield-lock</v-icon>
          <v-icon v-else size="28" color="primary">mdi-cellphone-link</v-icon>
          <h6 class="text-h6">{{ $gettext(`Apps and Devices`) }}</h6>
        </v-card-title>
        <!-- Confirm -->
        <template v-if="confirmAction !== ''">
          <v-card-text class="dense">
            <v-row align="start" dense>
              <v-col cols="12" class="text-body-2">
                {{ $gettext(`Enter your password to confirm the action and continue:`) }}
              </v-col>
              <v-col cols="12">
                <v-text-field
                  v-model="password"
                  :disabled="busy"
                  name="password"
                  :type="showPassword ? 'text' : 'password'"
                  :label="$gettext('Password')"
                  hide-details
                  autofocus
                  autocorrect="off"
                  autocapitalize="none"
                  autocomplete="current-password"
                  class="input-password text-monospace text-selectable"
                  :append-inner-icon="showPassword ? 'mdi-eye-off' : 'mdi-eye'"
                  prepend-inner-icon="mdi-lock"
                  @click:append-inner="showPassword = !showPassword"
                  @keyup.enter="onConfirm"
                ></v-text-field>
              </v-col>
            </v-row>
          </v-card-text>
          <v-card-actions class="action-buttons">
            <v-btn variant="flat" color="secondary-light" class="action-back" @click.stop="onBack">
              {{ $gettext(`Back`) }}
            </v-btn>
            <v-btn
              variant="flat"
              color="highlight"
              :disabled="!password || password.length < 4"
              class="action-confirm"
              @click.stop="onConfirm"
            >
              {{ $gettext(`Continue`) }}
            </v-btn>
          </v-card-actions>
        </template>
        <!-- Copy -->
        <template v-else-if="action === 'copy'">
          <v-card-text class="dense">
            <v-row align="start" dense>
              <v-col cols="12" class="text-body-2">
                {{
                  $gettext(
                    `Please copy the following randomly generated app password and keep it in a safe place, as you will not be able to see it again:`
                  )
                }}
              </v-col>
              <v-col cols="12">
                <v-text-field
                  v-model="appPassword"
                  type="text"
                  hide-details
                  readonly
                  autocorrect="off"
                  autocapitalize="none"
                  autocomplete="off"
                  append-inner-icon="mdi-content-copy"
                  class="input-app-password text-selectable cursor-copy"
                  @click.stop="onCopyAppPassword"
                >
                </v-text-field>
              </v-col>
            </v-row>
          </v-card-text>
          <v-card-actions class="action-buttons">
            <v-btn variant="flat" color="button" class="action-close" @click.stop="close">
              {{ $gettext(`Close`) }}
            </v-btn>
            <v-btn
              v-if="appPasswordCopied"
              variant="flat"
              color="highlight"
              :disabled="busy"
              class="action-done"
              @click.stop="onDone"
            >
              {{ $gettext(`Done`) }}
            </v-btn>
            <v-btn v-else variant="flat" color="highlight" class="action-copy" @click.stop="onCopyAppPassword">
              {{ $gettext(`Copy`) }}
            </v-btn>
          </v-card-actions>
        </template>
        <!-- Add -->
        <template v-else-if="action === 'add'">
          <v-card-text class="dense">
            <v-row align="start" dense>
              <v-col cols="12" class="text-body-2">
                {{
                  $gettext(
                    `To generate a new app-specific password, please enter the name and authorization scope of the application and select an expiration date:`
                  )
                }}
              </v-col>
              <v-col cols="12">
                <v-text-field
                  v-model="app.client_name"
                  :disabled="busy"
                  name="client_name"
                  type="text"
                  :label="$gettext('Name')"
                  autofocus
                  hide-details
                  autocorrect="off"
                  autocapitalize="none"
                  autocomplete="off"
                  class="input-name text-selectable"
                ></v-text-field>
              </v-col>
              <v-col cols="12" sm="6">
                <v-select
                  v-model="app.scope"
                  hide-details
                  :disabled="busy"
                  item-title="text"
                  item-value="value"
                  :items="auth.ScopeOptions()"
                  :label="$gettext('Scope')"
                  :menu-props="{ maxHeight: 346 }"
                  class="input-scope"
                ></v-select>
              </v-col>
              <v-col cols="12" sm="6">
                <v-select
                  v-model="app.expires_in"
                  :disabled="busy"
                  :label="$gettext('Expires')"
                  autocomplete="off"
                  hide-details
                  class="input-expires"
                  item-title="text"
                  item-value="value"
                  :items="options.Expires()"
                ></v-select>
              </v-col>
            </v-row>
          </v-card-text>
          <v-card-actions class="action-buttons">
            <v-btn variant="flat" color="button" class="action-cancel" @click.stop="onCancel">
              {{ $gettext(`Cancel`) }}
            </v-btn>
            <v-btn
              variant="flat"
              color="highlight"
              :disabled="app.client_name === '' || app.scope === ''"
              class="action-generate"
              @click.stop="onGenerate"
            >
              {{ $gettext(`Generate`) }}
            </v-btn>
          </v-card-actions>
        </template>
        <!-- Apps -->
        <template v-else>
          <v-card-text>
            <v-row align="start" no-gutters>
              <v-col cols="12">
                <v-data-table
                  v-model="selected"
                  :headers="listColumns"
                  :items="results"
                  :items-per-page-options="[]"
                  hide-default-footer
                  item-key="ID"
                  :no-data-text="$gettext('Nothing was found.')"
                  density="compact"
                  class="elevation-0 user-results list-view"
                >
                  <template #item="props">
                    <tr :data-name="props.item.ClientName">
                      <td class="text-selectable text-break text-start">
                        {{ props.item.ClientName }}
                      </td>
                      <td class="text-start text-break hidden-xs" nowrap>
                        {{ scopeInfo(props.item.AuthScope) }}
                      </td>
                      <td class="text-start text-break" nowrap>
                        {{ formatDateTime(props.item.LastActive) }}
                      </td>
                      <td class="text-start hidden-sm-and-down" nowrap>
                        {{ formatDate(props.item.Expires) }}
                      </td>
                      <td>
                        <div class="table-actions">
                          <v-btn
                            icon="mdi-delete"
                            color="surface-variant"
                            density="compact"
                            variant="plain"
                            :ripple="false"
                            class="action-remove action-secondary"
                            @click.stop.prevent="onRevoke(props.item)"
                          ></v-btn>
                        </div>
                      </td>
                    </tr>
                  </template>
                </v-data-table>
              </v-col>
            </v-row>
          </v-card-text>
          <v-card-actions class="action-buttons">
            <v-btn variant="flat" color="button" class="action-close" @click.stop="close">
              {{ $gettext(`Close`) }}
            </v-btn>
            <v-btn variant="flat" color="highlight" class="action-add" @click.stop="onAdd">
              {{ $gettext(`Add`) }}
            </v-btn>
          </v-card-actions>
        </template>
      </v-card>
    </v-form>
    <p-confirm-dialog
      :visible="revoke.dialog"
      icon="mdi-delete-outline"
      @close="revoke.dialog = false"
      @confirm="onRevoked"
    ></p-confirm-dialog>
  </v-dialog>
</template>
<script>
import User from "model/user";
import * as auth from "options/auth";
import * as options from "options/options";
import { DateTime } from "luxon";
import memoizeOne from "memoize-one";
import PConfirmDialog from "component/confirm/dialog.vue";

export default {
  name: "PSettingsApps",
  components: {
    PConfirmDialog,
  },
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    model: {
      type: Object,
      default: () => new User(null),
    },
  },
  data() {
    return {
      auth,
      options,
      busy: false,
      isDemo: this.$config.get("demo"),
      isPublic: this.$config.get("public"),
      password: "",
      showPassword: false,
      minLength: this.$config.get("passwordLength"),
      maxLength: 72,
      rtl: this.$isRtl,
      action: "",
      confirmAction: "",
      user: this.$session.getUser(),
      results: [],
      selected: [],
      app: {
        client_name: "",
        scope: "*",
        expires_in: 0,
      },
      revoke: {
        token: "",
        dialog: false,
      },
      appPassword: "",
      appPasswordCopied: false,
      listColumns: [
        { title: this.$gettext("Name"), key: "ID", sortable: false, align: "left" },
        {
          title: this.$gettext("Scope"),
          headerProps: {
            class: "hidden-xs",
          },
          key: "AuthScope",
          sortable: false,
          align: "left",
        },
        {
          title: this.$gettext("Last Used"),
          headerProps: {
            class: "text-no-wrap",
          },
          key: "LastActive",
          sortable: false,
          align: "left",
        },
        {
          title: this.$gettext("Expires"),
          headerProps: {
            class: "hidden-sm-and-down",
          },
          key: "Expires",
          sortable: false,
          align: "left",
          mobile: false,
        },
        { title: "", key: "", sortable: false, align: "right" },
      ],
    };
  },
  watch: {
    visible: function (show) {
      if (show) {
        this.reset();
        this.find();
      }
    },
  },
  created() {
    if (this.isPublic && !this.isDemo) {
      this.$emit("close");
    }
  },
  methods: {
    afterEnter() {
      this.$view.enter(this);
    },
    afterLeave() {
      this.$view.leave(this);
    },
    onCopyAppPassword() {
      this.$util.copyText(this.appPassword);
      this.appPasswordCopied = true;
    },
    formatDate(d) {
      if (!d) {
        return "–";
      }

      if (!Number.isInteger(d)) {
        return DateTime.fromISO(d).toLocaleString(DateTime.DATE_SHORT);
      } else if (d <= 0) {
        return "–";
      }

      return DateTime.fromSeconds(d).toLocaleString(DateTime.DATE_SHORT);
    },
    formatDateTime(d) {
      if (!d) {
        return "–";
      }

      if (!Number.isInteger(d)) {
        return DateTime.fromISO(d).toLocaleString(DateTime.DATETIME_SHORT);
      } else if (d <= 0) {
        return "–";
      }

      return DateTime.fromSeconds(d).toLocaleString(DateTime.DATETIME_SHORT);
    },
    scopeInfo(s) {
      let info = memoizeOne(auth.Scopes)()[s];
      if (info) {
        return info;
      }
      return s;
    },
    reset(action) {
      if (!action) {
        action = "apps";
      }

      this.app = {
        client_name: "",
        scope: "*",
        expires_in: 0,
      };

      this.action = action;
      this.confirmAction = "";
      this.appPasswordCopied = false;
      this.revoke.token = "";
      this.revoke.dialog = false;
    },
    onConfirm() {
      if (this.busy) {
        return;
      }

      switch (this.confirmAction) {
        case "onGenerate":
          this.onGenerate();
      }
    },
    onDone() {
      if (this.busy) {
        return;
      }

      this.appPassword = "";
      this.reset();
      this.find();
    },
    onCancel() {
      if (this.busy) {
        return;
      }

      this.reset();
    },
    onBack() {
      if (this.busy) {
        return;
      }

      this.confirmAction = "";
    },
    onAdd() {
      if (this.busy) {
        return;
      }

      this.action = "add";
      this.confirmAction = "";
    },
    onRevoke(app) {
      if (this.busy) {
        return;
      }

      this.revoke.token = app.ID;
      this.revoke.dialog = true;
    },
    onRevoked() {
      if (this.busy || !this.revoke.token) {
        return;
      }

      this.busy = true;
      this.$session
        .deleteApp(this.revoke.token)
        .then(() => {
          this.$notify.info(this.$gettext("Successfully deleted"));
          this.revoke.token = "";
          this.find();
          this.revoke.dialog = false;
          this.busy = false;
        })
        .catch(() => {
          this.busy = false;
        });
    },
    onGenerate() {
      if (this.busy) {
        return;
      }

      if (this.confirmAction === "" && this.$session.provider !== "oidc") {
        this.confirmAction = "onGenerate";
        return;
      }

      this.busy = true;
      this.$session
        .createApp(this.app.client_name, this.app.scope, this.app.expires_in, this.password)
        .then((app) => {
          this.appPassword = app.access_token;
          this.reset("copy");
        })
        .catch(() => {
          this.action = "add";
          this.confirmAction = "";
        })
        .finally(() => {
          this.busy = false;
        });
    },
    find() {
      this.$notify.blockUI();
      this.model
        .findApps()
        .then((resp) => {
          this.results = resp;
        })
        .finally(() => {
          this.$notify.unblockUI();
        });
    },
    close() {
      if (this.busy) {
        return;
      }

      this.appPassword = "";

      this.$emit("close");
    },
  },
};
</script>
