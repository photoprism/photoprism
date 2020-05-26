<template>
    <div class="p-tab p-settings-general">
        <v-form lazy-validation dense
                ref="form" class="p-form-settings" accept-charset="UTF-8"
                @submit.prevent="onChange">
            <v-card flat tile class="mt-0 px-1 application">
                <v-card-title primary-title class="pb-0">
                    <h3 class="body-2 mb-0"><translate>Library</translate></h3>
                </v-card-title>

                <v-card-actions>
                    <v-layout wrap align-top>
                        <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
                            <v-checkbox
                                    @change="onChange"
                                    :disabled="busy"
                                    class="ma-0 pa-0"
                                    v-model="settings.features.private"
                                    color="secondary-dark"
                                    :label="labels.private"
                                    :hint="hints.private"
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
                                    v-model="settings.features.review"
                                    color="secondary-dark"
                                    :label="labels.review"
                                    :hint="hints.review"
                                    prepend-icon="remove_red_eye"
                                    persistent-hint
                            >
                            </v-checkbox>
                        </v-flex>

                        <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
                            <v-checkbox
                                    @change="onChange"
                                    :disabled="busy || readonly"
                                    class="ma-0 pa-0"
                                    v-model="settings.index.convert"
                                    color="secondary-dark"
                                    :label="labels.convert"
                                    :hint="hints.convert"
                                    prepend-icon="photo_camera"
                                    persistent-hint
                            >
                            </v-checkbox>
                        </v-flex>

                        <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
                            <v-checkbox
                                    @change="onChange"
                                    :disabled="busy"
                                    class="ma-0 pa-0"
                                    v-model="settings.index.group"
                                    color="secondary-dark"
                                    :label="labels.group"
                                    :hint="hints.group"
                                    prepend-icon="photo_library"
                                    persistent-hint
                            >
                            </v-checkbox>
                        </v-flex>
                    </v-layout>
                </v-card-actions>
            </v-card>

            <v-card flat tile class="mt-0 px-1 application">
                <v-card-title primary-title class="pb-2">
                    <h3 class="body-2 mb-0"><translate>User Interface</translate></h3>
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

            <v-card flat tile class="mt-0 px-1 application">
                <v-card-actions>
                    <v-layout wrap align-top>
                        <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
                            <v-checkbox
                                    @change="onChange"
                                    :disabled="busy || readonly"
                                    class="ma-0 pa-0"
                                    v-model="settings.features.upload"
                                    color="secondary-dark"
                                    :label="labels.upload"
                                    :hint="hints.upload"
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
                                    :label="labels.download"
                                    :hint="hints.download"
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
                                    v-model="settings.features.share"
                                    color="secondary-dark"
                                    :label="labels.share"
                                    :hint="hints.share"
                                    prepend-icon="share"
                                    persistent-hint
                            >
                            </v-checkbox>
                        </v-flex>

                        <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
                            <v-checkbox
                                    @change="onChange"
                                    :disabled="busy || readonly"
                                    class="ma-0 pa-0"
                                    v-model="settings.features.import"
                                    color="secondary-dark"
                                    :label="labels.import"
                                    :hint="hints.import"
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
                                    :label="labels.archive"
                                    :hint="hints.archive"
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
                                    v-model="settings.features.edit"
                                    color="secondary-dark"
                                    :label="labels.edit"
                                    :hint="hints.edit"
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
                                    v-model="settings.features.files"
                                    color="secondary-dark"
                                    :label="labels.files"
                                    :hint="hints.files"
                                    prepend-icon="insert_drive_file"
                                    persistent-hint
                            >
                            </v-checkbox>
                        </v-flex>

                        <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
                            <v-checkbox
                                    @change="onChange"
                                    :disabled="busy"
                                    class="ma-0 pa-0"
                                    v-model="settings.features.moments"
                                    color="secondary-dark"
                                    :label="labels.moments"
                                    :hint="hints.moments"
                                    prepend-icon="star"
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
                                    :label="labels.labels"
                                    :hint="hints.labels"
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
                                    v-model="settings.features.logs"
                                    color="secondary-dark"
                                    :label="labels.logs"
                                    :hint="hints.logs"
                                    prepend-icon="notes"
                                    persistent-hint
                            >
                            </v-checkbox>
                        </v-flex>

                        <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
                            <v-checkbox
                                    @change="onChange"
                                    :disabled="busy"
                                    class="ma-0 pa-0"
                                    v-model="settings.features.places"
                                    color="secondary-dark"
                                    :label="labels.places"
                                    :hint="hints.places"
                                    prepend-icon="place"
                                    persistent-hint
                            >
                            </v-checkbox>
                        </v-flex>
                    </v-layout>
                </v-card-actions>
            </v-card>

            <v-card flat tile class="mt-0 px-1 application" v-if="settings.features.places">
                <v-card-title primary-title class="pb-2">
                    <h3 class="body-2 mb-0"><translate>Places</translate></h3>
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
                            PhotoPrism™ {{$config.get("version")}}
                            <br>© 2018-2020 <a href="mailto:hello@photoprism.org" class="secondary-dark--text" target="_blank">PhotoPrism.org</a>
                        </v-flex>

                        <v-flex xs12 sm6 class="px-2 pb-2 body-1 text-xs-left text-sm-right">
                            A big <a href="https://docs.photoprism.org/en/latest/credits/" class="secondary-dark--text"
                                                            target="_blank">thank you</a> to everyone who made this possible!
                            <br>
                            <a href="https://raw.githubusercontent.com/photoprism/photoprism/develop/NOTICE"
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
                experimental: this.$config.get("experimental"),
                settings: new Settings(this.$config.settings()),
                options: options,
                labels: {
                    language: this.$gettext("Language"),
                    theme: this.$gettext("Theme"),
                    mapsAnimate: this.$gettext("Animation"),
                    mapsStyle: this.$gettext("Style"),
                    rescan: this.$gettext("Complete rescan"),
                    thumbs: this.$gettext("Create thumbnails"),
                    move: this.$gettext("Remove imported files"),
                    group: this.$gettext("Group Sequential"),
                    archive: this.$gettext("Archive"),
                    private: this.$gettext("Hide Private"),
                    review: this.$gettext("Quality Filter"),
                    places: this.$gettext("Places"),
                    files: this.$gettext("Files"),
                    moments: this.$gettext("Moments"),
                    labels: this.$gettext("Labels"),
                    import: this.$gettext("Import"),
                    upload: this.$gettext("Upload"),
                    download: this.$gettext("Download"),
                    edit: this.$gettext("Edit"),
                    share: this.$gettext("Share"),
                    logs: this.$gettext("Logs"),
                    convert: this.$gettext("Convert to JPEG"),
                },
                hints: {
                    private: this.$gettext("Exclude photos marked as private from search results, shared albums, labels and places."),
                    review: this.$gettext("Non-photographic and low-quality images require a review before they appear in search results."),
                    group: this.$gettext("Files with sequential names like 'IMG_1234 (2)' or 'IMG_1234 copy 2' belong to the same photo."),
                    move: this.$gettext("Move files from import to originals to save storage. Unsupported file types will never be deleted, they remain in their current location."),
                    places: this.$gettext("Search and display photos on a map."),
                    files: this.$gettext("Browse indexed files and folders."),
                    moments: this.$gettext("Let PhotoPrism create albums from past events."),
                    labels: this.$gettext("Browse and edit image classification labels."),
                    import: this.$gettext("Imported files will be sorted by date and given a unique name."),
                    archive: this.$gettext("Hide photos that have been moved to archive."),
                    upload: this.$gettext("Add files to your library via Web Upload."),
                    download: this.$gettext("Download single files and zip archives."),
                    edit: this.$gettext("Change photo titles, locations and other metadata."),
                    share: this.$gettext("Upload to WebDAV and other remote services."),
                    logs: this.$gettext("Show server logs in Library."),
                    convert: this.$gettext("File types like RAW might need to be converted so that they can be displayed in a browser. JPEGs will be stored in the same folder next to the original using the best possible quality."),
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
