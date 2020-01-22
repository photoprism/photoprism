<template>
    <div class="p-page p-page-login">
        <v-toolbar flat color="secondary">
            <v-toolbar-title><translate>Authentication required</translate></v-toolbar-title>

            <v-spacer></v-spacer>
        </v-toolbar>

        <v-container class="pt-5">
            <p class="subheading">
                <span><translate>Please enter your password to proceed:</translate></span>
            </p>
            <v-form ref="form" autocomplete="off" class="p-form-login" @submit.prevent="login" dense>
                <v-text-field
                        :label="labels.password"
                        color="accent"
                        v-model="password"
                        solo
                        flat
                        :append-icon="showPassword ? 'visibility' : 'visibility_off'"
                        :type="showPassword ? 'text' : 'password'"
                        @click:append="showPassword = !showPassword"
                ></v-text-field>
                <v-btn color="secondary-dark"
                       class="white--text ml-0"
                       depressed
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
                showPassword: false,
                password: '',
                nextUrl: this.$route.params.nextUrl ? this.$route.params.nextUrl : "/",
                labels: {
                    password: this.$gettext("Password"),
                }
            };
        },
        methods: {
            login() {
                this.$session.login('admin', this.password).then(
                    () => {
                        this.$router.push(this.nextUrl);
                    }
                );
            },
        }
    };
</script>
