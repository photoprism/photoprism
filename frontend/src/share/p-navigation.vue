<template>
    <div id="p-navigation">
        <v-toolbar dark scroll-off-screen color="navigation darken-1" class="hidden-md-and-up p-navigation-small"
                   @click.stop="showNavigation()">
            <v-toolbar-side-icon class="p-navigation-show"></v-toolbar-side-icon>

            <v-toolbar-title class="p-navigation-title">{{ page.title }}</v-toolbar-title>
        </v-toolbar>
        <v-navigation-drawer
                v-model="drawer"
                mini-variant
                :width="270"
                :mobile-break-point="960"
                class="p-navigation-sidebar navigation"
                fixed dark app
        >
            <v-toolbar flat>
                <v-list class="navigation-home">
                    <v-list-tile class="p-navigation-logo">
                        <v-list-tile-avatar class="clickable" @click.stop.prevent="openDocs">
                            <img src="/static/img/logo.png" alt="Logo">
                        </v-list-tile-avatar>
                        <v-list-tile-content>
                            <v-list-tile-title class="title">
                                PhotoPrism
                            </v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>
                </v-list>
            </v-toolbar>

            <v-list class="pt-3" v-if="auth">
                <v-list-tile :to="{ name: 'albums', params: { token: token } }" :exact="true" @click="" class="p-navigation-albums">
                    <v-list-tile-action>
                        <v-icon>share</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>
                            <translate key="Albums">Albums</translate>
                        </v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>
            </v-list>
        </v-navigation-drawer>
    </div>
</template>

<script>
    import Album from "../model/album";
    import Event from "pubsub-js";

    export default {
        name: "p-navigation",
        data() {
            return {
                drawer: null,
                mini: true,
                session: this.$session,
                public: this.$config.get("public"),
                readonly: this.$config.get("readonly"),
                config: this.$config.values,
                page: this.$config.page,
                upload: {
                    subscription: null,
                    dialog: false,
                },
                edit: {
                    subscription: null,
                    dialog: false,
                    album: null,
                    selection: [],
                    index: 0,
                },
                token: this.$route.params.token,
                uid: this.$route.params.uid,
            };
        },
        computed: {
            auth() {
                return this.session.auth || this.public
            },
        },
        methods: {
            feature(name) {
                return this.$config.values.settings.features[name];
            },
            openDocs() {
                window.open("https://docs.photoprism.org/", "_blank");
            },
            showNavigation() {
                this.drawer = true;
                this.mini = false;
            },
            logout() {
                this.$session.logout();
            },
        },
    };
</script>
