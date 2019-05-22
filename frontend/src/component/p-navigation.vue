<template>
    <div id="p-navigation">
        <v-toolbar dark fixed color="grey darken-3" class="hidden-lg-and-up p-navigation-small" @click.stop="showNavigation()">
            <v-toolbar-side-icon class="p-navigation-show"></v-toolbar-side-icon>

            <v-toolbar-title class="p-navigation-title">{{ $router.currentRoute.name }}</v-toolbar-title>

            <v-spacer></v-spacer>
        </v-toolbar>
        <v-toolbar dark flat color="grey darken-3" class="hidden-lg-and-up">
            <!-- Invisible -->
            <v-spacer></v-spacer>
        </v-toolbar>
        <v-navigation-drawer
                v-model="drawer"
                :mini-variant="mini"
                class="p-navigation-sidebar"
                width="270"
                fixed dark app
        >
            <v-toolbar flat>
                <v-list>
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

                <v-list-tile to="/photos" @click="" class="p-navigation-photos">
                    <v-list-tile-action>
                        <v-icon>photo</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>Photos</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile to="/places" @click="" class="p-navigation-places">
                    <v-list-tile-action>
                        <v-icon>place</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>Places</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile to="/calendar" @click="" class="p-navigation-timeline">
                    <v-list-tile-action>
                        <v-icon>date_range</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>Timeline</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile to="/tags" @click="" class="p-navigation-scenes">
                    <v-list-tile-action>
                        <v-icon>label</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>Scenes</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile to="/favorites" @click="">
                    <v-list-tile-action>
                        <v-icon>favorite</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>Favorites</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile to="/bookmarks" @click="">
                    <v-list-tile-action>
                        <v-icon>bookmark</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>Bookmarks</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile v-if="mini" to="/filters" @click="">
                    <v-list-tile-action>
                        <v-icon>search</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>Filters</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-group v-if="!mini" prepend-icon="search" no-action>
                    <v-list-tile slot="activator" to="/filters" @click="">
                        <v-list-tile-content>
                            <v-list-tile-title>Filters</v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>

                    <v-list-tile @click="">
                        <v-list-tile-content>
                            <v-list-tile-title>Canon EOS 6D</v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>

                    <v-list-tile @click="">
                        <v-list-tile-content>
                            <v-list-tile-title>iPhone</v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>

                </v-list-group>

                <v-list-tile v-if="mini" to="/albums" @click="">
                    <v-list-tile-action>
                        <v-icon>folder</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>Albums</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-group v-if="!mini" prepend-icon="folder" no-action>
                    <v-list-tile slot="activator" to="/albums" @click="">
                        <v-list-tile-content>
                            <v-list-tile-title>Albums</v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>

                    <v-list-tile @click="">
                        <v-list-tile-content>
                            <v-list-tile-title>South Africa</v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>

                    <v-list-tile @click="">
                        <v-list-tile-content>
                            <v-list-tile-title>Cats &amp; Dogs</v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>

                    <v-list-tile @click="">
                        <v-list-tile-content>
                            <v-list-tile-title>Create album</v-list-tile-title>
                        </v-list-tile-content>

                        <v-list-tile-action>
                            <v-icon>add</v-icon>
                        </v-list-tile-action>
                    </v-list-tile>
                </v-list-group>

                <v-list-tile to="/import" @click="">
                    <v-list-tile-action>
                        <v-icon>add_circle</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>Import</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile to="/export" @click="">
                    <v-list-tile-action>
                        <v-icon>save</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>Export</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile to="/settings" @click="">
                    <v-list-tile-action>
                        <v-icon>settings</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>Settings</v-list-tile-title>
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
            return {
                drawer: null,
                mini: false,
            };
        },
        methods: {
            showNavigation: function () {
                this.drawer = true;
                this.mini = false;
            }
        }
    };
</script>
