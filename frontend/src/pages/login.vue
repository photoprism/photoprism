<template>
  <div class="p-page p-page-login">
    <v-toolbar flat color="secondary" dense class="mb-3" :height="42">
      <v-toolbar-title class="subheading">
        {{ siteDescription }}
      </v-toolbar-title>
    </v-toolbar>
    <v-form ref="form" dense autocomplete="off" class="p-form-login" accept-charset="UTF-8" @submit.prevent="login">
      <v-card flat tile class="ma-2 application">
        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 class="pa-2">
              <v-text-field
                  v-model="username"
                  required hide-details
                  type="text"
                  :disabled="loading"
                  :label="$gettext('Name')"
                  browser-autocomplete="off"
                  color="secondary-dark"
                  class="input-name"
                  placeholder="••••••••"
              ></v-text-field>
            </v-flex>
            <v-flex xs12 class="pa-2">
              <v-text-field
                  v-model="password"
                  required hide-details
                  :type="showPassword ? 'text' : 'password'"
                  :disabled="loading"
                  :label="$gettext('Password')"
                  browser-autocomplete="off"
                  color="secondary-dark"
                  placeholder="••••••••"
                  class="input-password"
                  :append-icon="showPassword ? 'visibility' : 'visibility_off'"
                  @click:append="showPassword = !showPassword"
                  @keyup.enter.native="login"
              ></v-text-field>
            </v-flex>
            <v-flex xs12 class="px-2 py-3">
              <v-btn depressed :disabled="loading || !password || !username"
                     class="white--text ml-0 action-confirm"
                     color="primary-button" @click.stop="login">
                <translate>Sign in</translate>
                <v-icon :right="!rtl" :left="rtl" dark>login</v-icon>
              </v-btn>
            </v-flex>
          </v-layout>
        </v-card-actions>
      </v-card>
    </v-form>

    <p-about-footer v-if="!sponsor"></p-about-footer>
  </div>
</template>

<script>
export default {
  name: "PPageLogin",
  data() {
    const c = this.$config.values;

    return {
      loading: false,
      showPassword: false,
      username: "admin",
      password: "",
      sponsor: this.$config.values.sponsor,
      siteDescription: c.siteDescription ? c.siteDescription : c.siteCaption,
      nextUrl: this.$route.params.nextUrl ? this.$route.params.nextUrl : "/",
      rtl: this.$rtl,
    };
  },
  methods: {
    login() {
      if (!this.username || !this.password) {
        return;
      }

      this.loading = true;
      this.$session.login(this.username, this.password).then(
        () => {
          this.loading = false;
          this.$router.push(this.nextUrl);
        }
      ).catch(() => this.loading = false);
    },
  }
};
</script>
