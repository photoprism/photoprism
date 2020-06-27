<template>
  <div class="p-page p-page-login">
    <v-toolbar flat color="secondary">
      <v-toolbar-title>
        {{ $config.get("siteTitle") }}: {{ $config.get("siteCaption") }}
      </v-toolbar-title>

      <v-spacer></v-spacer>
    </v-toolbar>

    <v-container class="pt-4">
      <p class="subheading">
        <span><translate>Please enter your name and password to proceed:</translate></span>
      </p>
      <v-form ref="form" autocomplete="off" class="p-form-login" @submit.prevent="login" dense>
        <v-text-field
                :disabled="loading"
                :label="labels.username"
                color="accent"
                v-model="username"
                flat solo required
                type="text"
        ></v-text-field>
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
        <v-btn color="secondary-dark"
               class="white--text ml-0"
               depressed
               :disabled="loading || !this.password || !this.username"
               @click.stop="login">
          <translate>Sign in</translate>
          <v-icon right dark>vpn_key</v-icon>
        </v-btn>
      </v-form>
    </v-container>
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
