<template>
    <div class="p-tab p-settings-general">
        <v-form lazy-validation dense
                ref="form" class="p-form-settings" accept-charset="UTF-8"
                @submit.prevent="onChange">
            <v-card flat tile class="mt-0 px-1 application">
                <v-card-title primary-title class="pb-0">
                    <h3 class="body-2 mb-0">Library</h3>
                </v-card-title>

                <v-card-actions>
                    <v-layout wrap align-top>
                        <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
                            <v-checkbox
                                    @change="onChange"
                                    :disabled="busy"
                                    class="ma-0 pa-0"
                                    v-model="settings.library.private"
                                    color="secondary-dark"
                                    :label="labels.private"
                                    hint="Exclude photos marked as private from search results by default."
                                    prepend-icon="lock"
                                    persistent-hint
                            >
                            </v-checkbox>
                        </v-flex>

                        <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
                            <v-checkbox
                                    @change="onChange"
                                    :disabled="busy"
                                    class="ma-0 pa-0"
                                    v-model="settings.library.review"
                                    color="secondary-dark"
                                    :label="labels.review"
                                    hint="Low-quality photos require a review before they appear in search results."
                                    prepend-icon="remove_red_eye"
                                    persistent-hint
                            >
                            </v-checkbox>
                        </v-flex>

                        <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
                            <v-checkbox
                                    @change="onChange"
                                    :disabled="busy"
                                    class="ma-0 pa-0"
                                    v-model="settings.library.group"
                                    color="secondary-dark"
                                    :label="labels.group"
                                    hint="Files with sequential names like 'IMG_1234 (2)' or 'IMG_1234 copy 2' belong to the same photo."
                                    prepend-icon="photo_library"
                                    persistent-hint
                            >
                            </v-checkbox>
                        </v-flex>

                        <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
                            <v-checkbox
                                    @change="onChange"
                                    :disabled="busy"
                                    class="ma-0 pa-0"
                                    v-model="settings.library.move"
                                    color="secondary-dark"
                                    :label="labels.move"
                                    hint="Move files from import to originals to save storage.
                                    Unsupported file types will never be deleted, they remain in their current location."
                                    prepend-icon="delete"
                                    persistent-hint
                            >
                            </v-checkbox>
                        </v-flex>
                    </v-layout>
                </v-card-actions>
            </v-card>

            <v-card flat tile class="mt-0 px-1 application">
                <v-card-title primary-title class="pb-0">
                    <h3 class="body-2 mb-0">Features</h3>
                </v-card-title>

                <v-card-actions>
                    <v-layout wrap align-top>
                        <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
                            <v-checkbox
                                    @change="onChange"
                                    :disabled="busy"
                                    class="ma-0 pa-0"
                                    v-model="settings.features.places"
                                    color="secondary-dark"
                                    label="Places"
                                    hint="Search and display photos on a map."
                                    prepend-icon="place"
                                    persistent-hint
                            >
                            </v-checkbox>
                        </v-flex>

                        <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
                            <v-checkbox
                                    @change="onChange"
                                    :disabled="busy"
                                    class="ma-0 pa-0"
                                    v-model="settings.features.labels"
                                    color="secondary-dark"
                                    label="Labels"
                                    hint="Browse and edit image classification labels."
                                    prepend-icon="label"
                                    persistent-hint
                            >
                            </v-checkbox>
                        </v-flex>

                        <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
                            <v-checkbox
                                    @change="onChange"
                                    :disabled="busy"
                                    class="ma-0 pa-0"
                                    v-model="settings.features.import"
                                    color="secondary-dark"
                                    label="Import"
                                    hint="Imported files will be sorted by date and given a unique name."
                                    prepend-icon="create_new_folder"
                                    persistent-hint
                            >
                            </v-checkbox>
                        </v-flex>

                        <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
                            <v-checkbox
                                    @change="onChange"
                                    :disabled="busy"
                                    class="ma-0 pa-0"
                                    v-model="settings.features.archive"
                                    color="secondary-dark"
                                    label="Archive"
                                    hint="Hide photos that have been moved to archive."
                                    prepend-icon="archive"
                                    persistent-hint
                            >
                            </v-checkbox>
                        </v-flex>

                        <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
                            <v-checkbox
                                    @change="onChange"
                                    :disabled="busy"
                                    class="ma-0 pa-0"
                                    v-model="settings.features.upload"
                                    color="secondary-dark"
                                    label="Upload"
                                    hint="Add files to your library via Web Upload."
                                    prepend-icon="cloud_upload"
                                    persistent-hint
                            >
                            </v-checkbox>
                        </v-flex>

                        <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
                            <v-checkbox
                                    @change="onChange"
                                    :disabled="busy"
                                    class="ma-0 pa-0"
                                    v-model="settings.features.download"
                                    color="secondary-dark"
                                    label="Download"
                                    hint="Download single files and zip archives."
                                    prepend-icon="cloud_download"
                                    persistent-hint
                            >
                            </v-checkbox>
                        </v-flex>

                        <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
                            <v-checkbox
                                    @change="onChange"
                                    :disabled="busy"
                                    class="ma-0 pa-0"
                                    v-model="settings.features.edit"
                                    color="secondary-dark"
                                    label="Edit"
                                    hint="Change photo titles, locations and other metadata."
                                    prepend-icon="edit"
                                    persistent-hint
                            >
                            </v-checkbox>
                        </v-flex>

                        <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
                            <v-checkbox
                                    @change="onChange"
                                    :disabled="busy"
                                    class="ma-0 pa-0"
                                    v-model="settings.features.share"
                                    color="secondary-dark"
                                    label="Share"
                                    hint="Upload to WebDAV and other remote services."
                                    prepend-icon="share"
                                    persistent-hint
                            >
                            </v-checkbox>
                        </v-flex>
                    </v-layout>
                </v-card-actions>
            </v-card>

            <v-card flat tile class="px-1 application">
                <v-card-title primary-title class="pb-2">
                    <h3 class="body-2 mb-0">User Interface</h3>
                </v-card-title>

                <v-card-actions>
                    <v-layout wrap align-top>
                        <v-flex xs12 sm6 class="px-2 pb-2">
                            <v-select
                                    @change="onChange"
                                    :disabled="busy"
                                    :items="options.themes"
                                    :label="labels.theme"
                                    color="secondary-dark"
                                    background-color="secondary-light"
                                    v-model="settings.theme"
                                    hide-details box
                            ></v-select>
                        </v-flex>

                        <v-flex xs12 sm6 class="px-2 pb-2">
                            <v-select
                                    @change="onChange"
                                    :disabled="busy"
                                    :items="options.languages"
                                    :label="labels.language"
                                    color="secondary-dark"
                                    background-color="secondary-light"
                                    v-model="settings.language"
                                    hide-details box
                            ></v-select>
                        </v-flex>
                    </v-layout>
                </v-card-actions>
            </v-card>

            <v-card flat tile class="mt-0 px-1 application" v-if="settings.features.places">
                <v-card-title primary-title class="pb-2">
                    <h3 class="body-2 mb-0">Places</h3>
                </v-card-title>

                <v-card-actions>
                    <v-layout wrap align-top>
                        <v-flex xs12 sm6 class="px-2 pb-2">
                            <v-select
                                    @change="onChange"
                                    :disabled="busy"
                                    :items="options.mapsStyle"
                                    :label="labels.mapsStyle"
                                    color="secondary-dark"
                                    background-color="secondary-light"
                                    v-model="settings.maps.style"
                                    hide-details box
                            ></v-select>
                        </v-flex>

                        <v-flex xs12 sm6 class="px-2 pb-2">
                            <v-select
                                    @change="onChange"
                                    :disabled="busy"
                                    :items="options.mapsAnimate"
                                    :label="labels.mapsAnimate"
                                    color="secondary-dark"
                                    background-color="secondary-light"
                                    v-model="settings.maps.animate"
                                    hide-details box
                            ></v-select>
                        </v-flex>
                    </v-layout>
                </v-card-actions>
            </v-card>

            <v-card flat tile class="mt-0 px-1 application">
                <v-card-actions>
                    <v-layout wrap align-top>
                        <v-flex xs12 sm6 class="px-2 pb-2 body-1">
                            PhotoPrism {{$config.get("version")}}
                            <br>{{$config.get("copyright")}}
                        </v-flex>

                        <v-flex xs12 sm6 class="px-2 pb-2 body-1 text-xs-left text-sm-right">
                            A big <a href="https://docs.photoprism.org/en/latest/credits/" class="secondary-dark--text"
                                     target="_blank">thank you</a> to everyone who made this possible!
                            <br>
                            <a href="https://app.fossa.com/attribution/d068c7e2-586d-421a-86a4-f23104a408f9"
                               class="secondary-dark--text" target="_blank">
                                3rd-party software packages</a>
                        </v-flex>
                    </v-layout>
                </v-card-actions>
            </v-card>
        </v-form>
    </div>
</template>

<script>
    import Settings from "model/settings";
    import options from "resources/options.json";

    export default {
        name: 'p-settings-general',
        data() {
            return {
                readonly: this.$config.get("readonly"),
                settings: new Settings(this.$config.settings()),
                options: options,
                labels: {
                    language: this.$gettext("Language"),
                    theme: this.$gettext("Theme"),
                    mapsAnimate: this.$gettext("Animation"),
                    mapsStyle: this.$gettext("Style"),
                    rescan: this.$gettext("Complete rescan"),
                    thumbs: this.$gettext("Create thumbnails"),
                    raw: this.$gettext("Convert RAW files"),
                    move: this.$gettext("Remove imported files"),
                    group: this.$gettext("Group related files"),
                    private: this.$gettext("Hide private content"),
                    review: this.$gettext("Apply quality filter"),
                },
                busy: false,
            };
        },
        methods: {
            load() {
                this.settings.load();
            },
            onChange() {
                const reload = this.settings.changed("language");

                if (reload) {
                    this.busy = true;
                }

                this.settings.save().then((s) => {
                    if (reload) {
                        this.$notify.info(this.$gettext("Reloading..."));
                        this.$notify.blockUI();
                        setTimeout(() => window.location.reload(), 100);
                    } else {
                        this.$notify.success(this.$gettext("Settings saved"));
                    }
                }).finally(() => this.busy = false)
            },
        },
        created() {
            this.load();
        },
    };
</script>
