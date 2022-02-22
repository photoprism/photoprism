<template>
  <v-container fluid fill-height class="auth-login wallpaper">
    <v-layout align-center justify-center>
      <v-flex xs12 sm8 md4>
        <v-form ref="form" dense autocomplete="off" class="auth-login-form" accept-charset="UTF-8"
                @submit.prevent="login">
          <v-card class="elevation-12 auth-login-box blur-7">
            <v-card-text class="pa-3">
              <div class="logo text-xs-center">
                <img :src="$config.appIcon()" :alt="config.name">
              </div>
              <v-spacer></v-spacer>
              <v-text-field
                  v-model="username"
                  required hide-details solo flat light autofocus
                  type="text"
                  :disabled="loading"
                  :label="$gettext('Name')"
                  browser-autocomplete="off"
                  background-color="grey lighten-5"
                  class="input-name"
                  color="#05dde1"
                  :placeholder="$gettext('Name')"
                  prepend-icon="person"
              ></v-text-field>
              <v-text-field
                  v-model="password"
                  required hide-details solo flat light
                  :type="showPassword ? 'text' : 'password'"
                  :disabled="loading"
                  :label="$gettext('Password')"
                  browser-autocomplete="off"
                  background-color="grey lighten-5"
                  :placeholder="$gettext('Password')"
                  class="input-password mt-1"
                  :append-icon="showPassword ? 'visibility' : 'visibility_off'"
                  prepend-icon="lock"
                  color="#05dde1"
                  @click:append="showPassword = !showPassword"
                  @keyup.enter.native="login"
              ></v-text-field>
              <v-spacer></v-spacer>
              <div class="pa-3 text-xs-center">
                <!-- a href="#" target="_blank" class="text-link px-2"><translate>Forgot password?</translate></a -->
                <v-btn color="#00adb0" round :disabled="loading || !password || !username"
                       class="white--text action-confirm px-3" @click.stop="login">
                  <translate>Sign in</translate>
                  <v-icon :right="!rtl" :left="rtl" dark>arrow_forward</v-icon>
                </v-btn>
              </div>
            </v-card-text>
          </v-card>
        </v-form>
      </v-flex>
    </v-layout>
    <footer>
      <p class="auth-site float-left white--text body-2">
        <strong class="white--text">{{ config.siteTitle }}</strong> – {{ config.siteCaption }}
      </p>
      <p v-if="config.imprint" class="auth-imprint float-right white--text body-2">
        <a v-if="config.imprintUrl" :href="config.imprintUrl" target="_blank" class="text-link">{{ config.imprint }}</a>
        <span v-else>{{ config.imprint }}</span>
      </p>
    </footer>
  </v-container>
</template>

<script>

export default {
  name: "PPageAuthLogin",
  data() {
    const c = this.$config.values;

    return {
      loading: false,
      showPassword: false,
      username: "",
      password: "",
      sponsor: this.$config.isSponsor(),
      config: this.$config.values,
      siteDescription: c.siteDescription ? c.siteDescription : c.siteCaption,
      nextUrl: this.$route.params.nextUrl ? this.$route.params.nextUrl : "/",
      rtl: this.$rtl,
    };
  },
  created() {
    this.$scrollbar.hide();
  },
  destroyed() {
    this.$scrollbar.show();
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
