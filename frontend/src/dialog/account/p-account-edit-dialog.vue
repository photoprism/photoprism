<template>
    <v-dialog lazy v-model="show" persistent max-width="500" class="p-account-edit-dialog" @keydown.esc="cancel">
        <v-card raised elevation="24">
            <v-card-title primary-title>
                <v-layout row wrap v-if="scope === 'sharing'">
                    <v-flex xs9>
                        <h3 class="headline mb-0">{{ $gettext('Upload') }}</h3>
                    </v-flex>
                    <v-flex xs3 text-xs-right>
                        <v-switch
                                v-model="model.AccShare"
                                color="secondary-dark"
                                :true-value="true"
                                :false-value="false"
                                :label="model.AccShare ? label.enable : label.disable"
                                :disabled="model.AccType !== 'webdav'"
                                class="mt-0 hidden-xs-only"
                                hide-details
                        ></v-switch>
                        <v-switch
                                v-model="model.AccShare"
                                color="secondary-dark"
                                :true-value="true"
                                :false-value="false"
                                :disabled="model.AccType !== 'webdav'"
                                class="mt-0 hidden-sm-and-up"
                                hide-details
                        ></v-switch>
                    </v-flex>
                </v-layout>
                <v-layout row wrap v-else-if="scope === 'sync'">
                    <v-flex xs9>
                        <h3 class="headline mb-0">{{ $gettext('Remote Sync') }}</h3>
                    </v-flex>
                    <v-flex xs3 text-xs-right>
                        <v-switch
                                v-model="model.AccSync"
                                color="secondary-dark"
                                :true-value="true"
                                :false-value="false"
                                :label="model.AccSync ? label.enable : label.disable"
                                :disabled="model.AccType !== 'webdav'"
                                class="mt-0 hidden-xs-only"
                                hide-details
                        ></v-switch>
                        <v-switch
                                v-model="model.AccSync"
                                color="secondary-dark"
                                :true-value="true"
                                :false-value="false"
                                :disabled="model.AccType !== 'webdav'"
                                class="mt-0 hidden-sm-and-up"
                                hide-details
                        ></v-switch>
                    </v-flex>
                </v-layout>
                <v-layout row wrap v-else>
                    <v-flex xs10>
                        <h3 class="headline mb-0">{{ $gettext('Edit Account') }}</h3>
                    </v-flex>
                    <v-flex xs2 text-xs-right>
                        <v-btn icon flat :ripple="false"
                               class="action-remove mt-0"
                               @click.stop.prevent="remove()">
                            <v-icon color="secondary-dark">delete</v-icon>
                        </v-btn>
                    </v-flex>
                </v-layout>
                <h3 class="headline mb-0" v-else>{{ $gettext('Edit Account') }}</h3>
            </v-card-title>
            <v-container fluid class="pt-0 pb-2 pr-2 pl-2">
                <v-layout row wrap v-if="scope === 'sharing'">
                    <v-flex xs12 class="pa-2">
                        <v-autocomplete
                                color="secondary-dark"
                                hide-details hide-no-data flat
                                v-model="model.SharePath"
                                browser-autocomplete="off"
                                hint="Folder"
                                :search-input.sync="search"
                                :items="pathItems"
                                :loading="loading"
                                item-text="abs"
                                item-value="abs"
                                :label="label.SharePath"
                                :disabled="!model.AccShare || loading"
                        >
                        </v-autocomplete>
                    </v-flex>
                    <v-flex xs12 sm6 class="pa-2 input-share-size">
                        <v-select
                                :disabled="!model.AccShare"
                                :label="label.ShareSize"
                                browser-autocomplete="off"
                                hide-details
                                color="secondary-dark"
                                item-text="text"
                                item-value="value"
                                v-model="model.ShareSize"
                                :items="items.sizes">
                        </v-select>
                    </v-flex>
                    <v-flex xs12 sm6 class="pa-2">
                        <v-select
                                :disabled="!model.AccShare"
                                :label="label.ShareExpires"
                                browser-autocomplete="off"
                                hide-details
                                color="secondary-dark"
                                item-text="text"
                                item-value="value"
                                v-model="model.ShareExpires"
                                :items="items.expires">
                        </v-select>
                    </v-flex>
                </v-layout>
                <v-layout row wrap v-else-if="scope === 'sync'">
                    <v-flex xs12 sm6 class="pa-2">
                        <v-autocomplete
                                color="secondary-dark"
                                hide-details hide-no-data flat
                                v-model="model.SyncPath"
                                browser-autocomplete="off"
                                hint="Folder"
                                :search-input.sync="search"
                                :items="pathItems"
                                :loading="loading"
                                item-text="abs"
                                item-value="abs"
                                :label="label.SyncPath"
                                :disabled="!model.AccSync || loading"
                        >
                        </v-autocomplete>
                    </v-flex>
                    <v-flex xs12 sm6 class="pa-2">
                        <v-select
                                :disabled="!model.AccSync"
                                :label="label.SyncInterval"
                                browser-autocomplete="off"
                                hide-details
                                color="secondary-dark"
                                item-text="text"
                                item-value="value"
                                v-model="model.SyncInterval"
                                :items="items.intervals">
                        </v-select>
                    </v-flex>
                    <v-flex xs12 sm6 class="px-2">
                        <v-checkbox
                                :disabled="!model.AccSync || readonly"
                                hide-details
                                color="secondary-dark"
                                :label="label.SyncDownload"
                                v-model="model.SyncDownload"
                        ></v-checkbox>
                    </v-flex>
                    <v-flex xs12 sm6 class="px-2">
                        <v-checkbox
                                :disabled="!model.AccSync"
                                hide-details
                                color="secondary-dark"
                                :label="label.SyncFilenames"
                                v-model="model.SyncFilenames"
                        ></v-checkbox>
                    </v-flex>
                    <v-flex xs12 sm6 class="px-2">
                        <v-checkbox
                                :disabled="!model.AccSync"
                                hide-details
                                color="secondary-dark"
                                :label="label.SyncUpload"
                                v-model="model.SyncUpload"
                        ></v-checkbox>
                    </v-flex>
                    <v-flex xs12 sm6 class="px-2">
                        <v-checkbox
                                :disabled="!model.AccSync"
                                hide-details
                                color="secondary-dark"
                                :label="label.SyncRaw"
                                v-model="model.SyncRaw"
                        ></v-checkbox>
                    </v-flex>
                </v-layout>
                <v-layout row wrap v-else>
                    <v-flex xs12 class="pa-2">
                        <v-text-field
                                hide-details
                                browser-autocomplete="off"
                                :label="label.name"
                                placeholder=""
                                color="secondary-dark"
                                v-model="model.AccName"
                                required
                        ></v-text-field>
                    </v-flex>
                    <v-flex xs12 class="pa-2">
                        <v-text-field
                                hide-details
                                browser-autocomplete="off"
                                :label="label.url"
                                placeholder="https://www.example.com/"
                                color="secondary-dark"
                                v-model="model.AccURL"
                        ></v-text-field>
                    </v-flex>
                    <v-flex xs12 sm6 class="pa-2">
                        <v-text-field
                                hide-details
                                browser-autocomplete="off"
                                :label="label.user"
                                placeholder="optional"
                                color="secondary-dark"
                                v-model="model.AccUser"
                        ></v-text-field>
                    </v-flex>
                    <v-flex xs12 sm6 class="pa-2">
                        <v-text-field
                                hide-details
                                browser-autocomplete="off"
                                :label="label.pass"
                                placeholder="optional"
                                color="secondary-dark"
                                v-model="model.AccPass"
                                :append-icon="showPassword ? 'visibility' : 'visibility_off'"
                                :type="showPassword ? 'text' : 'password'"
                                @click:append="showPassword = !showPassword"
                        ></v-text-field>
                    </v-flex>
                    <v-flex xs12 sm6 class="pa-2">
                        <v-text-field
                                hide-details
                                browser-autocomplete="off"
                                :label="label.apiKey"
                                placeholder="optional"
                                color="secondary-dark"
                                v-model="model.AccKey"
                                required
                        ></v-text-field>
                    </v-flex>
                    <v-flex xs12 sm6 pa-2 class="input-account-type">
                        <v-select
                                :label="label.AccType"
                                browser-autocomplete="off"
                                hide-details
                                color="secondary-dark"
                                item-text="text"
                                item-value="value"
                                v-model="model.AccType"
                                :items="items.types">
                        </v-select>
                    </v-flex>
                </v-layout>
                <v-layout row wrap>
                    <v-flex xs12 text-xs-right class="pt-3">
                        <v-btn @click.stop="cancel" depressed color="secondary-light"
                               class="action-cancel">
                            <span>{{ label.cancel }}</span>
                        </v-btn>
                        <v-btn depressed dark color="secondary-dark" @click.stop="save"
                               class="action-save">
                            <span>{{ label.save }}</span>
                        </v-btn>
                    </v-flex>
                </v-layout>
            </v-container>
        </v-card>
    </v-dialog>
</template>
<script>
    import options from "resources/options";

    export default {
        name: 'p-account-edit-dialog',
        props: {
            show: Boolean,
            scope: String,
            model: Object,
        },
        data() {
            const thumbs = this.$config.values.thumbnails;

            thumbs.sort((a, b) => a.Width - b.Width);

            return {
                showPassword: false,
                loading: false,
                search: null,
                path: "/",
                paths: [
                    {"abs": "/"}
                ],
                pathItems: [],
                newPath: "",
                items: {
                    thumbs: thumbs,
                    sizes: this.sizes(thumbs),
                    types: [
                        {"value": "web", "text": "Web"},
                        {"value": "webdav", "text": "WebDAV / Nextcloud"},
                        {"value": "facebook", "text": "Facebook"},
                        {"value": "twitter", "text": "Twitter"},
                        {"value": "flickr", "text": "Flickr"},
                        {"value": "instagram", "text": "Instagram"},
                        {"value": "eyeem", "text": "EyeEm"},
                        {"value": "telegram", "text": "Telegram"},
                        {"value": "whatsapp", "text": "WhatsApp"},
                        {"value": "gphotos", "text": "Google Photos"},
                        {"value": "gdrive", "text": "Google Drive"},
                        {"value": "onedrive", "text": "Microsoft OneDrive"},
                    ],
                    intervals: [
                        {"value": 0, "text": "Never"},
                        {"value": 3600, "text": "1 hour"},
                        {"value": 3600 * 4, "text": "4 hours"},
                        {"value": 3600 * 12, "text": "12 hours"},
                        {"value": 86400, "text": "Daily"},
                        {"value": 86400 * 2, "text": "Every two days"},
                        {"value": 86400 * 7, "text": "Once a week"},
                    ],
                    expires: [
                        {"value": 0, "text": "Never"},
                        {"value": 86400, "text": "After 1 day"},
                        {"value": 86400 * 3, "text": "After 3 days"},
                        {"value": 86400 * 7, "text": "After 7 days"},
                        {"value": 86400 * 14, "text": "After two weeks"},
                        {"value": 86400 * 31, "text": "After one month"},
                        {"value": 86400 * 60, "text": "After two months"},
                        {"value": 86400 * 365, "text": "After one year"},
                    ],
                },
                readonly: this.$config.get("readonly"),
                options: options,
                label: {
                    cancel: this.$gettext("Cancel"),
                    confirm: this.$gettext("Save"),
                    save: this.$gettext("Save"),
                    enable: this.$gettext("Enable"),
                    disable: this.$gettext("Disable"),
                    name: this.$gettext("Name"),
                    url: this.$gettext("Service URL"),
                    user: this.$gettext("Username"),
                    pass: this.$gettext("Password"),
                    owner: this.$gettext("Owner"),
                    apiKey: this.$gettext("API Key"),
                    AccType: this.$gettext("Type"),
                    SharePath: this.$gettext("Default Folder"),
                    ShareSize: this.$gettext("Size"),
                    ShareExpires: this.$gettext("Expires"),
                    SyncPath: this.$gettext("Folder"),
                    SyncInterval: this.$gettext("Interval"),
                    SyncFilenames: this.$gettext("Preserve filenames"),
                    SyncStart: this.$gettext("Start"),
                    SyncDownload: this.$gettext("Download remote files"),
                    SyncUpload: this.$gettext("Upload local files"),
                    SyncDelete: this.$gettext("Remote delete"),
                    SyncRaw: this.$gettext("Sync raw images"),
                }
            }
        },
        computed: {},
        methods: {
            cancel() {
                this.$emit('cancel');
            },
            remove() {
                this.$emit('remove');
            },
            confirm() {
                this.model.AccShare = true;
                this.save();
            },
            disable(prop) {
                this.model[prop] = false;

                this.save();
            },
            enable(prop) {
                this.model[prop] = true;
            },
            save() {
                if (this.loading) {
                    this.$notify.wait();
                    return;
                }

                this.loading = true;

                this.model.update().then(() => {
                    this.loading = false;
                    this.$emit('confirm');
                });
            },
            sizes(thumbs) {
                const result = [
                    {"text": this.$gettext("Original"), "value": ""}
                ];

                for (let i in thumbs) {
                    const s = thumbs[i];
                    result.push({"text": s["Width"] + 'x' + s["Height"], "value": s["Name"]});
                }

                return result;
            },
            onChange() {
                this.paths = [{"abs": "/"}];

                this.loading = true;
                this.model.Dirs().then(p => {
                    for (let i = 0; i < p.length; i++) {
                        this.paths.push(p[i]);
                    }

                    this.pathItems = [...this.paths];
                    this.path = this.model.SharePath;
                }).finally(() => this.loading = false);
            },
        },
        watch: {
            search(q) {
                if (this.loading) return;

                const exists = this.paths.findIndex((p) => p.value === q);

                if (exists !== -1 || !q) {
                    this.pathItems = this.paths;
                    this.newPath = "";
                } else {
                    this.newPath = q;
                    this.pathItems = this.paths.concat([{"abs": q}]);
                }
            },
            show: function (show) {
                if (show) {
                    this.onChange();
                }
            }
        },
    }
</script>
