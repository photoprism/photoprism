<template>
  <v-dialog
    :model-value="visible"
    persistent
    max-width="500"
    class="p-dialog p-service-add"
    @keydown.esc="close"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-form ref="form" validate-on="invalid-input" accept-charset="UTF-8" tabindex="1" @submit.prevent>
      <v-card>
        <v-card-title class="d-flex justify-start align-center ga-3">
          <v-icon size="28" color="primary">mdi-swap-horizontal</v-icon>
          <h6 class="text-h6">
            {{ $gettext(`Add Account`) }}
          </h6>
        </v-card-title>
        <v-card-text class="dense">
          <v-row align="center" dense>
            <v-col cols="12">
              <v-text-field
                v-model="model.AccURL"
                hide-details
                autofocus
                :label="$gettext('Service URL')"
                placeholder="https://www.example.com/"
                autocorrect="off"
                autocomplete="off"
                autocapitalize="none"
              ></v-text-field>
            </v-col>
            <v-col cols="12" sm="6">
              <v-text-field
                v-model="model.AccUser"
                hide-details
                :label="$gettext('Username')"
                :placeholder="$gettext('optional')"
                autocorrect="off"
                autocomplete="off"
                autocapitalize="none"
              ></v-text-field>
            </v-col>
            <v-col cols="12" sm="6">
              <v-text-field
                v-model="model.AccPass"
                hide-details
                autocorrect="off"
                autocapitalize="none"
                autocomplete="new-password"
                :label="$gettext('Password')"
                :placeholder="$gettext('optional')"
                :append-inner-icon="showPassword ? 'mdi-eye-off' : 'mdi-eye'"
                :type="showPassword ? 'text' : 'password'"
                @click:append-inner="showPassword = !showPassword"
              ></v-text-field>
            </v-col>
            <v-col cols="12" class="text-start text-caption">
              {{
                $gettext(
                  `Note: Only WebDAV servers, like Nextcloud or PhotoPrism, can be configured as remote service for backup and file upload.`
                )
              }}
              {{ $gettext(`Support for additional services, like Google Drive, will be added over time.`) }}
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions class="action-buttons">
          <v-btn variant="flat" color="button" class="action-cancel action-close" @click.stop="close">
            <span>{{ label.cancel }}</span>
          </v-btn>
          <v-btn variant="flat" color="highlight" class="action-confirm" @click.stop="confirm">
            <span>{{ label.confirm }}</span>
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-form>
  </v-dialog>
</template>
<script>
import Service from "model/service";
import * as options from "options/options";

export default {
  name: "PServiceAdd",
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      options: options,
      showPassword: false,
      loading: false,
      model: new Service(),
      label: {
        cancel: this.$gettext("Cancel"),
        confirm: this.$gettext("Connect"),
      },
    };
  },
  watch: {
    visible: function (show) {
      if (show) {
        this.loading = false;
        this.showPassword = false;
        this.model = new Service();
      }
    },
  },
  methods: {
    afterEnter() {
      this.$view.enter(this);
    },
    afterLeave() {
      this.$view.leave(this);
    },
    close() {
      this.$emit("close");
    },
    confirm() {
      if (this.loading) {
        return;
      }

      this.loading = true;

      this.model
        .save()
        .then((a) => {
          this.$notify.success(this.$gettext("Account created"));
          this.$emit("confirm", a.UID);
        })
        .finally(() => {
          this.loading = false;
        });
    },
  },
};
</script>
