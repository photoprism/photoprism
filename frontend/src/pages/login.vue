<template>
  <div class="p-page p-page-login">
    <v-toolbar flat color="secondary">
      <v-toolbar-title>
        {{ $config.get("siteCaption") }}
      </v-toolbar-title>
    </v-toolbar>

    <v-form dense ref="form" autocomplete="off" class="p-form-login" @submit.prevent="login" accept-charset="UTF-8">
      <v-card flat tile class="ma-2 application">
        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 class="pa-2">
              <p class="subheading">
                <translate>Please enter your name and password to proceed:</translate>
              </p>
            </v-flex>
            <v-flex xs12 class="pa-2">

              <v-text-field
                      :disabled="loading"
                      :label="labels.username"
                      color="accent"
                      v-model="username"
                      flat solo required
                      type="text"
              ></v-text-field>
            </v-flex>
            <v-flex xs12 class="pa-2">
              <v-text-field
                      :disabled="loading"
                      :label="labels.password"
                      color="accent"
                      v-model="password"
                      flat solo required
                      :append-icon="showPassword ? 'visibility' : 'visibility_off'"
                      :type="showPassword ? 'text' : 'password'"
                      @click:append="showPassword = !showPassword"
                      @keyup.enter.native="login"
              ></v-text-field>
            </v-flex>
            <v-flex xs12 class="pa-2">
              <v-btn color="secondary-dark"
                     class="white--text ml-0"
                     depressed
                     :disabled="loading || !this.password || !this.username"
                     @click.stop="login">
                <translate>Sign in</translate>
                <v-icon right dark>vpn_key</v-icon>
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
        name: 'login',
        data() {
            return {
                loading: false,
                showPassword: false,
                username: "admin",
                password: "",
                nextUrl: this.$route.params.nextUrl ? this.$route.params.nextUrl : "/",
                labels: {
                    username: this.$gettext("Name"),
                    password: this.$gettext("Password"),
                }
            };
        },
        methods: {
            login() {
                if (!this.username || !this.password) {
                    return
                }

                this.loading = true;
                this.$session.login(this.username, this.password).then(
                    () => {
                        this.loading = false;
                        this.$router.push(this.nextUrl);
                    }
                ).catch(() => this.loading = false)
            },
        }
    };
</script>
