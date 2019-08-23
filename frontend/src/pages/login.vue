<template>
  <v-layout align-center justify-space-between id="login-page">
    <v-card class="mx-auto" id="login-card">
      <v-img
        class="white--text"
        height="150px"
        src="/static/img/logo.png"
      >
        <v-card-title class="align-center justify-center fill-height display-3">Photoprism</v-card-title>
      </v-img>
      <v-card-actions>
        <v-layout row wrap align-center justify-space-between id="login-form">
          <v-text-field
            @click:append="passwordVisable = !passwordVisable"
            @keyup.enter="login"
            :append-icon="passwordVisable ? 'visibility' : 'visibility_off'"
            :type="passwordVisable ? 'text' : 'password'"
            :error-messages="passwordErrors"
            error
            label="enter password"
            prepend-inner-icon="lock"
            required
            v-model="password"
            width="100%"
          ></v-text-field>
          <v-btn 
            @click="login"
            id="login-button"
          >login</v-btn>
        </v-layout>
      </v-card-actions>
    </v-card>
  </v-layout>
</template>

<script>
    export default {
        name: 'login',
        data() {
            return {
                password: '',
                passwordVisable: false,
                passwordErrors: [],
            };
        },
        methods: {
            async login () {
                this.passwordErrors = []
                this.$session.setKey(this.password)
                if (!await this.$session.isAuthed()) {
                    this.passwordErrors = ["invalid password entered"]
                    this.$session.deleteKey()
                    return
                }
                this.$router.replace(
                    this.$router.currentRoute.query.redirect || '/',
                )
            }
        }
    };
</script>
