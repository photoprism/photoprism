<template>
    <div class="p-page p-page-login">
        <v-toolbar flat color="blue-grey lighten-4">
            <v-toolbar-title>Login</v-toolbar-title>

            <v-spacer></v-spacer>
        </v-toolbar>

        <v-container class="pt-5">
            <p>Please enter the admin password to proceed:</p>
            <v-form ref="form" autocomplete="off" class="p-form-login" dense>
                <v-text-field
                        label="Password"
                        color="grey"
                        v-model="password"
                        solo
                        flat
                        :append-icon="showPassword ? 'visibility' : 'visibility_off'"
                        :type="showPassword ? 'text' : 'password'"
                        @click:append="showPassword = !showPassword"
                ></v-text-field>
                <v-btn color="blue-grey"
                       class="white--text ml-0"
                       depressed
                       @click.stop="login">
                    Sign in
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
