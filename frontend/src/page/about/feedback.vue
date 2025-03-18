<template>
  <div class="p-page p-page-support" tabindex="1">
    <v-toolbar
      flat
      :density="$vuetify.display.smAndDown ? 'compact' : 'default'"
      color="secondary"
      class="page-toolbar p-page__navigation"
    >
      <v-toolbar-title>
        {{ $gettext(`Contact Us`) }}
      </v-toolbar-title>

      <v-btn icon>
        <v-icon size="26" color="surface-variant">mdi-message-text</v-icon>
      </v-btn>
    </v-toolbar>
    <div v-if="sent" class="pa-6">
      <h3 class="text-h6 font-weight-bold pt-6 pb-2 text-center">
        {{ $gettext(`We appreciate your feedback!`) }}
      </h3>
      <p class="text-body-2 py-6 text-center">
        {{
          $gettext(
            `Due to the high volume of emails we receive, our team may be unable to get back to you immediately.`
          )
        }}
        {{ $gettext(`We do our best to respond within five business days or less.`) }}
      </p>
    </div>
    <v-form v-else ref="form" v-model="valid" autocomplete="off" class="pa-4" validate-on="invalid-input">
      <v-row dense>
        <v-col cols="12">
          <v-select
            v-model="form.Category"
            :label="$gettext('Category')"
            :rules="rules.text(true, 1, 32, $gettext('Category'))"
            :items="options.FeedbackCategories()"
            :disabled="busy"
            validate-on="invalid-input"
            item-title="text"
            item-value="value"
            color="surface-variant"
            hide-details
            autocomplete="off"
            class="input-category"
          ></v-select>
        </v-col>

        <v-col cols="12">
          <v-textarea
            v-model="form.Message"
            :rules="rules.text(true, 1, 1000, $gettext('Message'))"
            :placeholder="$gettext('How can we help?')"
            validate-on="invalid-input"
            auto-grow
            hide-details
            autocomplete="off"
            rows="10"
          ></v-textarea>
        </v-col>

        <v-col cols="12" sm="6">
          <v-text-field
            v-model="form.UserName"
            :rules="rules.text(true, 1, 100, $gettext('Name'))"
            :label="$gettext('Name')"
            validate-on="blur"
            hide-details
            autocomplete="off"
            color="surface-variant"
            type="text"
          >
          </v-text-field>
        </v-col>

        <v-col cols="12" sm="6">
          <v-text-field
            v-model="form.UserEmail"
            :label="$gettext('E-Mail')"
            :rules="rules.email(true)"
            validate-on="blur"
            hide-details
            autocapitalize="none"
            color="surface-variant"
            type="email"
          >
          </v-text-field>
        </v-col>

        <v-col cols="12" class="d-flex grow">
          <v-btn :disabled="!valid" color="highlight" class="ml-0" @click.stop="send">
            {{ $gettext(`Submit`) }}
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
import $api from "common/api";
import { rules } from "common/form";

import PAboutFooter from "component/about/footer.vue";

export default {
  name: "PPageSupport",
  components: {
    PAboutFooter,
  },
  data() {
    return {
      rules,
      sent: false,
      busy: false,
      valid: false,
      options: options,
      form: {
        Category: "feedback",
        Message: "",
        UserName: "",
        UserEmail: "",
        UserAgent: navigator.userAgent,
        UserLocales: navigator.language,
      },
      rtl: this.$isRtl,
    };
  },
  mounted() {
    this.$view.enter(this);
  },
  unmounted() {
    this.$view.leave(this);
  },
  methods: {
    send() {
      if (this.$refs.form.validate()) {
        $api.post("feedback", this.form).then(() => {
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
