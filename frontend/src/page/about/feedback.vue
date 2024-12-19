<template>
  <div class="p-page p-page-support">
    <v-toolbar flat color="secondary" :density="$vuetify.display.smAndDown ? 'compact' : 'default'">
      <v-toolbar-title>
        <translate>Contact Us</translate>
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn icon>
        <v-icon size="26" color="surface-variant">mdi-message-text</v-icon>
      </v-btn>
    </v-toolbar>
    <v-container v-if="sent" fluid class="pa-6">
      <h3 class="text-h6 font-weight-bold pt-6 pb-2 text-center">
        <translate>We appreciate your feedback!</translate>
      </h3>
      <p class="text-body-2 py-6 text-center">
        <translate>Due to the high volume of emails we receive, our team may be unable to get back to you immediately.</translate>
        <translate>We do our best to respond within five business days or less.</translate>
      </p>
      <p class="mt-6 text-center">
        <img src="https://cdn.photoprism.app/thank-you/colorful.png" width="100%" alt="THANK YOU" />
      </p>
    </v-container>
    <v-form v-else ref="form" v-model="valid" autocomplete="off" class="pa-4" validate-on="blur">
      <v-row>
        <v-col cols="12" class="pa-2">
          <v-select
            v-model="form.Category"
            :disabled="busy"
            :items="options.FeedbackCategories()"
            item-title="text"
            item-value="value"
            :label="$gettext('Category')"
            color="surface-variant"
            hide-details
            required
            autocomplete="off"
            class="input-category"
            :rules="[(v) => !!v || $gettext('Required')]"
          ></v-select>
        </v-col>

        <v-col cols="12" class="pa-2">
          <v-textarea v-model="form.Message" required auto-grow hide-details autocomplete="off" rows="10" :rules="[(v) => !!v || $gettext('Required')]" :label="$gettext('How can we help?')"></v-textarea>
        </v-col>

        <v-col cols="12" sm="6" class="pa-2">
          <v-text-field v-model="form.UserName" hide-details autocomplete="off" color="surface-variant" :label="$gettext('Name')" type="text"> </v-text-field>
        </v-col>

        <v-col cols="12" sm="6" class="pa-2">
          <v-text-field v-model="form.UserEmail" hide-details required autocapitalize="none" color="surface-variant" :rules="[(v) => !!v || $gettext('Required')]" :label="$gettext('E-Mail')" type="email"> </v-text-field>
        </v-col>

        <v-col cols="12" class="d-flex grow px-2 py-1">
          <v-btn color="highlight" class="ml-0" :disabled="!form.Category || !form.Message || !form.UserEmail" @click.stop="send">
            <translate>Send</translate>
            <v-icon end>mdi-send</v-icon>
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
      // TODO: fix form
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
