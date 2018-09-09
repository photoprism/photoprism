<template>
    <v-navigation-drawer
            v-model="drawer"
            :mini-variant="mini"
            fixed
            dark
            app
            width="270"
            permanent
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

        <v-list class="pt-3">
            <v-list-tile v-if="mini" @click.stop="mini = !mini">
                <v-list-tile-action>
                    <v-icon>chevron_right</v-icon>
                </v-list-tile-action>
            </v-list-tile>

            <v-list-tile to="/photos" @click="">
                <v-list-tile-action>
                    <v-icon>photo</v-icon>
                </v-list-tile-action>

                <v-list-tile-content>
                    <v-list-tile-title>Photos</v-list-tile-title>
                </v-list-tile-content>
            </v-list-tile>

            <v-list-tile to="/places" @click="">
                <v-list-tile-action>
                    <v-icon>place</v-icon>
                </v-list-tile-action>

                <v-list-tile-content>
                    <v-list-tile-title>Places</v-list-tile-title>
                </v-list-tile-content>
            </v-list-tile>

            <v-list-tile to="/calendar" @click="">
                <v-list-tile-action>
                    <v-icon>date_range</v-icon>
                </v-list-tile-action>

                <v-list-tile-content>
                    <v-list-tile-title>Calendar</v-list-tile-title>
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
</template>

<script>
    export default {
        data() {
            const query = this.$route.query;
            const q = query.hasOwnProperty('q') ? query['q'] : '';

            return {
                items: [
                    { title: 'Photos', route: 'photos', icon: 'photo_library' },
                    { title: 'Filters', route: 'filters', icon: 'search' },
                    { title: 'Albums', route: 'albums', icon: 'folder' },
                    { title: 'Favorites', route: 'favorites', icon: 'favorite' },
                    { title: 'Bookmarks', route: 'bookmarks', icon: 'bookmark' },
                    { title: 'Places', route: 'places', icon: 'place' },
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
