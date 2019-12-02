<template>
    <div id="p-navigation">
        <v-toolbar dark scroll-off-screen color="navigation" class="hidden-lg-and-up p-navigation-small"
                   @click.stop="showNavigation()">
            <v-toolbar-side-icon class="p-navigation-show"></v-toolbar-side-icon>

            <v-toolbar-title class="p-navigation-title">{{ $router.currentRoute.meta.area }}</v-toolbar-title>

            <v-spacer></v-spacer>
        </v-toolbar>
        <v-navigation-drawer
                v-model="drawer"
                :mini-variant="mini"
                class="p-navigation-sidebar navigation"
                width="270"
                fixed dark app
        >
            <v-toolbar flat>
                <v-list class="navigation-home">
                    <v-list-tile class="p-navigation-logo">
                        <v-list-tile-avatar>
                            <img src="/static/img/logo.png">
                        </v-list-tile-avatar>
                        <v-list-tile-content>
                            <v-list-tile-title class="title">
                                PhotoPrism
                            </v-list-tile-title>
                        </v-list-tile-content>
                        <v-list-tile-action class="hidden-md-and-down">
                            <v-btn icon @click.stop="mini = !mini" class="p-navigation-minimize">
                                <v-icon>chevron_left</v-icon>
                            </v-btn>
                        </v-list-tile-action>
                    </v-list-tile>
                </v-list>
            </v-toolbar>

            <v-list class="pt-3">
                <v-list-tile v-if="mini" @click.stop="mini = !mini" class="p-navigation-expand">
                    <v-list-tile-action>
                        <v-icon>chevron_right</v-icon>
                    </v-list-tile-action>
                </v-list-tile>

                <v-list-tile v-if="mini" to="/photos" @click="" class="p-navigation-photos">
                    <v-list-tile-action>
                        <v-icon>photo</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title v-translate>Photos</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-group v-if="!mini" prepend-icon="photo" no-action>
                    <v-list-tile slot="activator" to="/photos" @click.stop="" class="p-navigation-photos">
                        <v-list-tile-content>
                            <v-list-tile-title v-translate>Photos</v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>

                    <v-list-tile :to="{name: 'photos', query: { q: 'mono:true' }}" :exact="true" @click="">
                        <v-list-tile-content>
                            <v-list-tile-title v-translate>Monochrome</v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>

                    <v-list-tile :to="{name: 'photos', query: { q: 'chroma:50' }}" :exact="true" @click="">
                        <v-list-tile-content>
                            <v-list-tile-title v-translate>Vibrant</v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>
                </v-list-group>

                <v-list-tile v-if="mini" to="/albums" @click="">
                    <v-list-tile-action>
                        <v-icon>folder</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title v-translate>Albums</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-group v-if="!mini" prepend-icon="folder" no-action>
                    <v-list-tile slot="activator" to="/albums" @click="">
                        <v-list-tile-content>
                            <v-list-tile-title v-translate>Albums</v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>

                    <v-list-tile @click.stop="$notify.warning('Work in progress')">
                        <v-list-tile-content>
                            <v-list-tile-title v-translate>Work in progress...</v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>
                </v-list-group>

                <v-list-tile to="/favorites" @click="" class="p-navigation-favorites">
                    <v-list-tile-action>
                        <v-icon>favorite</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title v-translate>Favorites</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile to="/places" @click="" class="p-navigation-places">
                    <v-list-tile-action>
                        <v-icon>place</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title v-translate>Places</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile to="/labels" @click="" class="p-navigation-labels">
                    <v-list-tile-action>
                        <v-icon>label</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title v-translate>Labels</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <!-- v-list-tile to="/events" @click="" class="p-navigation-events">
                    <v-list-tile-action>
                        <v-icon>date_range</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>Events</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile -->

                <!-- v-list-tile to="/people" @click="" class="p-navigation-people">
                    <v-list-tile-action>
                        <v-icon>people</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>People</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile -->

                <v-list-tile to="/library" @click="" class="p-navigation-library" v-if="session.auth || isPublic">
                    <v-list-tile-action>
                        <v-icon>camera_roll</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title v-translate>Library</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile to="/settings" @click="" class="p-navigation-settings" v-if="session.auth || isPublic">
                    <v-list-tile-action>
                        <v-icon>settings</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title v-translate>Settings</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile @click="logout" class="p-navigation-logout" v-if="!isPublic && session.auth">
                    <v-list-tile-action>
                        <v-icon>power_settings_new</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title v-translate>Logout</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile to="/login" @click=""  class="p-navigation-login" v-if="!isPublic && !session.auth">
                    <v-list-tile-action>
                        <v-icon>lock</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title v-translate>Login</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>
            </v-list>
        </v-navigation-drawer>
    </div>
</template>

<script>
    export default {
        name: "p-navigation",
        data() {
            let mini = (window.innerWidth < 1600);

            return {
                drawer: null,
                mini: mini,
                session: this.$session,
                isPublic: this.$config.getValue("public"),
            };
        },
        methods: {
            showNavigation: function () {
                this.drawer = true;
                this.mini = false;
            },
            logout() {
                this.$session.logout();
            },
        }
    };
</script>
