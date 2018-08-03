<template>
    <div class="app-navigation">
        <b-navbar toggleable="md" type="dark" variant="dark">

            <b-navbar-toggle target="nav_collapse"></b-navbar-toggle>

            <b-navbar-brand href="/"><img src="/assets/img/logo.png" alt="PhotoPrism"></b-navbar-brand>

            <b-collapse is-nav id="nav_collapse">

                <b-navbar-nav>
                    <b-nav-item to="/photos" class="nav-item-photos">Photos</b-nav-item>
                    <b-nav-item to="/albums" class="nav-item-albums">Albums</b-nav-item>
                    <b-nav-item to="/import" class="nav-item-import">Import</b-nav-item>
                    <b-nav-item to="/export" class="nav-item-export">Export</b-nav-item>
                    <b-nav-item to="/settings" class="nav-item-settings">Settings</b-nav-item>
                </b-navbar-nav>

                <!-- Right aligned nav items -->
                <b-navbar-nav class="ml-auto">

                    <b-nav-form>
                        <b-form-input size="sm" class="mr-sm-2" type="text" placeholder="Search"/>
                        <b-button size="sm" class="my-2 my-sm-0" type="submit">Search</b-button>
                    </b-nav-form>

                </b-navbar-nav>

            </b-collapse>
        </b-navbar>
    </div>
</template>

<script>
    export default {
        data() {
            return {};
        },
        methods: {
            toggleLeftSidenav() {
                this.$refs.leftSidenav.toggle();
            },

            open(ref) {
            },

            close(ref) {
            },

            login() {
                this.$refs.loginDialog.open();
            },

            register() {
                this.$refs.registerDialog.open();
            },

            logout() {
                this.$session.logout();
            },
        }
    };
</script>

<style scoped>
</style>
