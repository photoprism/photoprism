<template>
  <div class="p-page p-page-support">
    <v-toolbar flat color="secondary" :dense="$vuetify.breakpoint.smAndDown">
      <v-toolbar-title>
        <translate>Contact Us</translate>
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn icon>
        <v-icon size="26" color="secondary-dark">chat</v-icon>
      </v-btn>
    </v-toolbar>
    <v-container v-if="sent" fluid class="pa-6">
      <h3 class="title font-weight-bold pt-6 pb-2 text-xs-center">
        <translate>We appreciate your feedback!</translate>
      </h3>
      <p class="body-2 py-6 text-xs-center">
        <translate>Due to the high volume of emails we receive, our team may be unable to get back to you immediately.</translate>
        <translate>We do our best to respond within five business days or less.</translate>
      </p>
      <p class="mt-6 text-xs-center">
        <img src="https://cdn.photoprism.app/thank-you/colorful.png" width="100%" alt="THANK YOU" />
      </p>
    </v-container>
    <v-form v-else ref="form" v-model="valid" autocomplete="off" class="pa-4" lazy-validation>
      <v-row>
        <v-col cols="12" class="pa-2">
          <v-select
            v-model="form.Category"
            :disabled="busy"
            :items="options.FeedbackCategories()"
            :label="$gettext('Category')"
            color="secondary-dark"
            background-color="secondary-light"
            flat
            solo
            hide-details
            required
            autocomplete="off"
            class="input-category"
            :rules="[(v) => !!v || $gettext('Required')]"
          ></v-select>
        </v-col>

        <v-col cols="12" class="pa-2">
          <v-textarea v-model="form.Message" required auto-grow flat solo hide-details autocomplete="off" rows="10" :rules="[(v) => !!v || $gettext('Required')]" :label="$gettext('How can we help?')"></v-textarea>
        </v-col>

        <v-col cols="12" sm="6" class="pa-2">
          <v-text-field v-model="form.UserName" flat solo hide-details autocomplete="off" color="secondary-dark" background-color="secondary-light" :label="$gettext('Name')" type="text"> </v-text-field>
        </v-col>

        <v-col cols="12" sm="6" class="pa-2">
          <v-text-field v-model="form.UserEmail" flat solo hide-details required autocapitalize="none" color="secondary-dark" :rules="[(v) => !!v || $gettext('Required')]" background-color="secondary-light" :label="$gettext('E-Mail')" type="email"> </v-text-field>
        </v-col>

        <v-col cols="12" class="d-flex grow px-2 py-1">
          <v-btn color="primary-button" class="white--text ml-0 flex-grow-1" depressed :disabled="!form.Category || !form.Message || !form.UserEmail" @click.stop="send">
            <translate>Send</translate>
            <v-icon :right="!rtl" :left="rtl" dark>send</v-icon>
          </v-btn>
        </v-col>
      </v-row>
    </v-form>

    <p-about-footer></p-about-footer>
  </div>
</template>

<script>
import * as options from "options/options";
import Api from "common/api";

export default {
  name: "PPageSupport",
  data() {
    return {
      sent: false,
      busy: false,
      valid: false,
      options: options,
      form: {
        Category: "",
        Message: "",
        UserName: "",
        UserEmail: "",
        UserAgent: navigator.userAgent,
        UserLocales: navigator.language,
      },
      rtl: this.$rtl,
    };
  },
  methods: {
    send() {
      if (this.$refs.form.validate()) {
        Api.post("feedback", this.form).then(() => {
          this.$notify.success(this.$gettext("Message sent"));
          this.sent = true;
        });
      } else {
        this.$notify.error(this.$gettext("All fields are required"));
      }
    },
  },
};
</script>
