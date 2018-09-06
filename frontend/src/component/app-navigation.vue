<template>
    <v-navigation-drawer
            v-model="drawer"
            :mini-variant="mini"
            fixed
            dark
            app
    >
        <v-toolbar flat>
            <v-list>
                <v-list-tile>
                    <v-list-tile-avatar>
                        <img src="/assets/img/logo.png">
                    </v-list-tile-avatar>
                    <v-list-tile-content>
                        <v-list-tile-title class="title">
                            PhotoPrism
                        </v-list-tile-title>
                    </v-list-tile-content>
                    <v-list-tile-action>
                        <v-btn icon @click.stop="mini = !mini">
                            <v-icon>chevron_left</v-icon>
                        </v-btn>
                    </v-list-tile-action>
                </v-list-tile>
            </v-list>
        </v-toolbar>

        <!-- b-navbar-nav class="ml-auto">

            <b-nav-form action="/photos">
                <b-form-input size="sm" class="mr-sm-2" type="text" name="q" :value="q" placeholder="Search"/>
                <b-button size="sm" class="my-2 my-sm-0" type="submit">Search</b-button>
            </b-nav-form>

        </b-navbar-nav -->

        <v-list class="pt-3">
            <v-list-tile v-if="mini" @click.stop="mini = !mini">
                <v-list-tile-action>
                    <v-icon>chevron_right</v-icon>
                </v-list-tile-action>
            </v-list-tile>

            <v-list-tile
                    v-for="item in items"
                    :key="item.title"
                    :to="{ name: item.route }"
                    @click=""
            >
                <v-list-tile-action>
                    <v-icon>{{ item.icon }}</v-icon>
                </v-list-tile-action>

                <v-list-tile-content>
                    <v-list-tile-title>{{ item.title }}</v-list-tile-title>
                </v-list-tile-content>
            </v-list-tile>
        </v-list>
    </v-navigation-drawer>
</template>

<script>
    export default {
        data() {
            const query = this.$route.query;
            const q = query.hasOwnProperty('q') ? query['q'] : '';

            return {
                items: [
                    { title: 'Photos', route: 'photos', icon: 'photo_library' },
                    { title: 'Albums', route: 'albums', icon: 'folder' },
                    { title: 'Import', route: 'import', icon: 'add_circle' },
                    { title: 'Export', route: 'export', icon: 'save' },
                    { title: 'Settings', route: 'settings', icon: 'settings' },
                ],
                drawer: null,
                mini: false,
                q: q,
            };
        },
        methods: {
        }
    };
</script>

<style scoped>
</style>
