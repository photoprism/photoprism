<template>
  <div class="p-page p-page-connect">
    <v-toolbar flat color="secondary" :dense="$vuetify.breakpoint.smAndDown">
      <v-toolbar-title>
        <a href="https://link.photoprism.app/membership" target="_blank">Membership</a> <v-icon>navigate_next</v-icon>
        <span v-if="busy">
          <translate>Busy, please wait…</translate>
        </span>
        <span v-else-if="success">
          <translate>Verified</translate>
        </span>
        <span v-else>
          <translate>Connect</translate>
        </span>
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-shield" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
        <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
        <path d="M12 3a12 12 0 0 0 8.5 3a12 12 0 0 1 -8.5 15a12 12 0 0 1 -8.5 -15a12 12 0 0 0 8.5 -3"></path>
      </svg>
    </v-toolbar>
    <v-container fluid fill-height row wrap class="pa-4">
      <v-layout align-center justify-center fill-height>
        <v-flex v-if="busy" xs12 d-flex class="text-sm-center py-3">
          <v-progress-linear color="secondary-dark" :indeterminate="true"></v-progress-linear>
        </v-flex>
        <v-flex v-else-if="error" xs12 d-flex class="text-sm-left mb-3">
          <v-alert
              :value="true"
              color="error"
              icon="gpp_bad"
              class="mt-3"
              outline
          >
            {{ error }}
          </v-alert>
        </v-flex>
        <v-flex v-else-if="success" xs12 d-flex class="text-sm-left mb-3">
          <v-alert
              :value="true"
              color="success"
              icon="verified_user"
              class="mt-3"
              outline
          >
            <translate>Successfully Connected</translate>
          </v-alert>
        </v-flex>
        <v-flex v-else xs12 d-flex class="text-sm-left mb-3">
          <v-alert
              :value="true"
              color="warning"
              icon="gpp_maybe"
              class="mt-3 ra-4"
              outline
          >
            <translate>Request failed - invalid response</translate>
          </v-alert>
        </v-flex>
      </v-layout>
    </v-container>
    <p-about-footer></p-about-footer>
  </div>
</template>

<script>
import * as options from "options/options";
import Api from "common/api";

export default {
  name: 'PPageConnect',
  data() {
    return {
      success: false,
      busy: false,
      valid: false,
      error: "",
      options: options,
      rtl: this.$rtl,
    };
  },
  created() {
    this.$config.load().then(() => {
      this.send();
    });
  },
  methods: {
    send() {
      const name = this.$route.params.name.replace(/[^a-z0-9]/gi, '');
      const values = {Token: this.$route.params.token};

      if (name !== '' && values.Token.length >= 4) {
        this.busy = true;
        this.$notify.blockUI();
        Api.put("connect/"+name, values).then(() => {
          this.$notify.success(this.$gettext("Connected"));
          this.success = true;
          this.busy = false;
          this.$config.update();
        }).catch((error) => {
          this.busy = false;
          if (error.response && error.response.data) {
            let data = error.response.data;
            this.error = data.message ? data.message : data.error;
          }

          if(!this.error) {
            this.error = this.$gettext("Invalid parameters");
          }
        }).finally(() => {
          this.$notify.unblockUI();
        });
      } else {
        this.$notify.error(this.$gettext("Invalid parameters"));
        this.$router.push({name: "settings"});
      }

    },
  },
};
</script>
