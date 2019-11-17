<template>
    <div id="p-navigation">
        <v-app-bar dark scroll-off-screen color="grey darken-3" class="hidden-lg-and-up p-navigation-small"
                   @click.stop="showNavigation()">
            <v-app-bar-nav-icon class="p-navigation-show"></v-app-bar-nav-icon>

            <v-toolbar-title class="p-navigation-title">{{ $router.currentRoute.meta.area }}</v-toolbar-title>

            <v-spacer></v-spacer>
        </v-app-bar>
        <v-navigation-drawer
                v-model="drawer"
                :mini-variant="mini"
                class="p-navigation-sidebar"
                width="270"
                fixed dark app
        >
            <v-list-item class="p-navigation-logo" color="black">
                <v-list-item-avatar>
                    <v-img src="/static/img/logo.png"></v-img>
                </v-list-item-avatar>
                <v-list-item-title class="title">
                    PhotoPrism
                </v-list-item-title>

                <v-btn
                        icon
                        @click.stop="mini = !mini" class="p-navigation-minimize"
                >
                    <v-icon>chevron_left</v-icon>
                </v-btn>
            </v-list-item>

            <v-divider></v-divider>

            <v-list flat class="pt-3">
                <v-list-item v-if="mini" @click.stop="mini = !mini" class="p-navigation-expand">
                    <v-list-item-icon>
                        <v-icon>chevron_right</v-icon>
                    </v-list-item-icon>
                </v-list-item>

                <v-list-item v-if="mini" to="/photos" @click="" class="p-navigation-photos">
                    <v-list-item-icon>
                        <v-icon>photo</v-icon>
                    </v-list-item-icon>

                    <v-list-item-content>
                        <v-list-item-title>Photos</v-list-item-title>
                    </v-list-item-content>
                </v-list-item>

                <v-list-group v-if="!mini" prepend-icon="photo" no-action>
                    <v-list-item slot="activator" to="/photos" @click.stop="" class="p-navigation-photos">
                        <v-list-item-content>
                            <v-list-item-title>Photos</v-list-item-title>
                        </v-list-item-content>
                    </v-list-item>

                    <v-list-item :to="{name: 'photos', query: { q: 'mono:true' }}" :exact="true" @click="">
                        <v-list-item-content>
                            <v-list-item-title>Monochrome</v-list-item-title>
                        </v-list-item-content>
                    </v-list-item>

                    <v-list-item :to="{name: 'photos', query: { q: 'chroma:50' }}" :exact="true" @click="">
                        <v-list-item-content>
                            <v-list-item-title>Vibrant</v-list-item-title>
                        </v-list-item-content>
                    </v-list-item>
                </v-list-group>

                <v-list-item v-if="mini" to="/albums" @click="">
                    <v-list-item-icon>
                        <v-icon>folder</v-icon>
                    </v-list-item-icon>

                    <v-list-item-content>
                        <v-list-item-title>Albums</v-list-item-title>
                    </v-list-item-content>
                </v-list-item>

                <v-list-group v-if="!mini" prepend-icon="folder" no-action>
                    <v-list-item slot="activator" to="/albums" @click="">
                        <v-list-item-content>
                            <v-list-item-title>Albums</v-list-item-title>
                        </v-list-item-content>
                    </v-list-item>

                    <v-list-item @click.stop="$notify.warning('Work in progress')">
                        <v-list-item-content>
                            <v-list-item-title>Work in progress...</v-list-item-title>
                        </v-list-item-content>
                    </v-list-item>
                </v-list-group>

                <v-list-item to="/favorites" @click="" class="p-navigation-favorites">
                    <v-list-item-icon>
                        <v-icon>favorite</v-icon>
                    </v-list-item-icon>

                    <v-list-item-content>
                        <v-list-item-title>Favorites</v-list-item-title>
                    </v-list-item-content>
                </v-list-item>

                <v-list-item to="/places" @click="" class="p-navigation-places">
                    <v-list-item-icon>
                        <v-icon>place</v-icon>
                    </v-list-item-icon>

                    <v-list-item-content>
                        <v-list-item-title>Places</v-list-item-title>
                    </v-list-item-content>
                </v-list-item>

                <v-list-item to="/labels" @click="" class="p-navigation-labels">
                    <v-list-item-icon>
                        <v-icon>label</v-icon>
                    </v-list-item-icon>

                    <v-list-item-content>
                        <v-list-item-title>Labels</v-list-item-title>
                    </v-list-item-content>
                </v-list-item>

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

                <v-list-item to="/library" @click="" class="p-navigation-library" v-if="session.auth || isPublic">
                    <v-list-item-icon>
                        <v-icon>camera_roll</v-icon>
                    </v-list-item-icon>

                    <v-list-item-content>
                        <v-list-item-title>Library</v-list-item-title>
                    </v-list-item-content>
                </v-list-item>

                <v-list-item to="/settings" @click="" class="p-navigation-settings" v-if="session.auth || isPublic">
                    <v-list-item-icon>
                        <v-icon>settings</v-icon>
                    </v-list-item-icon>

                    <v-list-item-content>
                        <v-list-item-title>Settings</v-list-item-title>
                    </v-list-item-content>
                </v-list-item>

                <v-list-item @click="logout" class="p-navigation-logout" v-if="!isPublic && session.auth">
                    <v-list-item-icon>
                        <v-icon>power_settings_new</v-icon>
                    </v-list-item-icon>

                    <v-list-item-content>
                        <v-list-item-title>Logout</v-list-item-title>
                    </v-list-item-content>
                </v-list-item>

                <v-list-item to="/login" @click="" class="p-navigation-login" v-if="!isPublic && !session.auth">
                    <v-list-item-icon>
                        <v-icon>lock</v-icon>
                    </v-list-item-icon>

                    <v-list-item-content>
                        <v-list-item-title>Login</v-list-item-title>
                    </v-list-item-content>
                </v-list-item>
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
