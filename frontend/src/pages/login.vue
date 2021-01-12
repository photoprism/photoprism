<template>
  <div class="p-page p-page-login">
    <v-form ref="form" dense autocomplete="off" class="p-form-login" accept-charset="UTF-8" @submit.prevent="login">
      <v-card flat tile class="ma-2 application">
        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 class="pa-2">
              <p class="subheading">
                <translate>Please enter your name and password:</translate>
              </p>
            </v-flex>
            <v-flex xs12 class="pa-2">

              <v-text-field
                  v-model="username"
                  :disabled="loading"
                  :label="$gettext('Name')"
                  color="accent"
                  flat solo required hide-details
                  type="text"
              ></v-text-field>
            </v-flex>
            <v-flex xs12 class="pa-2">
              <v-text-field
                  v-model="password"
                  :disabled="loading"
                  :label="$gettext('Password')"
                  color="accent"
                  flat solo required hide-details
                  :append-icon="showPassword ? 'visibility' : 'visibility_off'"
                  :type="showPassword ? 'text' : 'password'"
                  @click:append="showPassword = !showPassword"
                  @keyup.enter.native="login"
              ></v-text-field>
            </v-flex>
            <v-flex xs12 class="px-2 py-3">
              <v-btn color="secondary-dark"
                     class="white--text ml-0"
                     depressed
                     :disabled="loading || !this.password || !this.username"
                     @click.stop="login">
                <translate>Sign in</translate>
                <v-icon :right="!rtl" :left="rtl" dark>login</v-icon>
              </v-btn>
            </v-flex>
          </v-layout>
        </v-card-actions>
      </v-card>
    </v-form>

    <p-about-footer></p-about-footer>
  </div>
</template>

<script>
export default {
  name: 'Login',
  data() {
    return {
      loading: false,
      showPassword: false,
      username: "admin",
      password: "",
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
