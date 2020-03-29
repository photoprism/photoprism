<template>
    <v-dialog lazy v-model="show" persistent max-width="500" class="p-account-edit-dialog" @keydown.esc="cancel">
        <v-card raised elevation="24">
            <v-card-title primary-title>
                <div>
                    <h3 class="headline mb-0" v-if="scope === 'sharing'">Upload Settings</h3>
                    <h3 class="headline mb-0" v-else-if="scope === 'sync'">Continuous Sync</h3>
                    <h3 class="headline mb-0" v-else>Edit Account</h3>
                </div>
            </v-card-title>
            <v-container fluid class="pt-0 pb-2 pr-2 pl-2">
                <v-layout row wrap v-if="scope === 'sharing'">
                    <v-flex xs12 class="pa-2">
                        <v-text-field
                                hide-details
                                :label="label.SharePath"
                                placeholder=""
                                color="secondary-dark"
                                v-model="model.SharePath"
                                required
                        ></v-text-field>
                    </v-flex>
                    <v-flex xs12 sm6 pa-2 class="input-share-size">
                        <v-select
                                :label="label.ShareSize"
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
                                :label="label.ShareExpires"
                                hide-details
                                color="secondary-dark"
                                item-text="text"
                                item-value="value"
                                v-model="model.ShareExpires"
                                :items="items.expires">
                        </v-select>
                    </v-flex>
                    <v-flex xs12 sm6 class="pa-2">
                        <v-checkbox
                                hide-details
                                color="secondary-dark"
                                :label="label.ShareExif"
                                v-model="model.ShareExif"
                        ></v-checkbox>
                    </v-flex>
                    <v-flex xs12 sm6 class="pa-2">
                        <v-checkbox
                                hide-details
                                color="secondary-dark"
                                :label="label.ShareSidecar"
                                v-model="model.ShareSidecar"
                        ></v-checkbox>
                    </v-flex>
                </v-layout>
                <v-layout row wrap v-else-if="scope === 'sync'">
                    <v-flex xs12 class="pa-2">
                        <v-text-field
                                hide-details
                                :label="label.SyncPath"
                                placeholder=""
                                color="secondary-dark"
                                v-model="model.SyncPath"
                                required
                        ></v-text-field>
                    </v-flex>
                    <v-flex xs12 sm6 class="px-2">
                        <v-checkbox
                                hide-details
                                color="secondary-dark"
                                :label="label.SyncDownload"
                                v-model="model.SyncDownload"
                        ></v-checkbox>
                    </v-flex>
                    <v-flex xs12 sm6 class="px-2">
                        <v-checkbox
                                hide-details
                                color="secondary-dark"
                                :label="label.SyncUpload"
                                v-model="model.SyncUpload"
                        ></v-checkbox>
                    </v-flex>
                    <!-- v-flex xs12 sm6 class="px-2">
                        <v-checkbox
                                hide-details
                                color="secondary-dark"
                                :label="label.SyncDelete"
                                v-model="model.SyncDelete"
                        ></v-checkbox>
                    </v-flex -->
                    <v-flex xs12 sm6 class="px-2">
                        <v-checkbox
                                hide-details
                                color="secondary-dark"
                                :label="label.SyncRaw"
                                v-model="model.SyncRaw"
                        ></v-checkbox>
                    </v-flex>
                    <v-flex xs12 sm6 class="px-2">
                        <v-checkbox
                                hide-details
                                color="secondary-dark"
                                :label="label.SyncVideo"
                                v-model="model.SyncVideo"
                        ></v-checkbox>
                    </v-flex>
                    <v-flex xs12 sm6 class="px-2">
                        <v-checkbox
                                hide-details
                                color="secondary-dark"
                                :label="label.SyncSidecar"
                                v-model="model.SyncSidecar"
                        ></v-checkbox>
                    </v-flex>
                </v-layout>
                <v-layout row wrap v-else>
                    <v-flex xs12 class="pa-2">
                        <v-text-field
                                hide-details
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
                                :label="label.url"
                                placeholder="https://www.example.com/"
                                color="secondary-dark"
                                v-model="model.AccURL"
                        ></v-text-field>
                    </v-flex>
                    <v-flex xs12 sm6 class="pa-2">
                        <v-text-field
                                hide-details
                                :label="label.user"
                                placeholder="optional"
                                color="secondary-dark"
                                v-model="model.AccUser"
                        ></v-text-field>
                    </v-flex>
                    <v-flex xs12 sm6 class="pa-2">
                        <v-text-field
                                hide-details
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
                                :label="label.apiKey"
                                placeholder="optional"
                                color="secondary-dark"
                                v-model="model.AccKey"
                                required
                        ></v-text-field>
                    </v-flex>
                    <!-- v-flex xs12 sm6 class="pa-2">
                        <v-text-field
                                hide-details
                                :label="label.owner"
                                placeholder="optional"
                                color="secondary-dark"
                                v-model="model.AccOwner"
                                required
                        ></v-text-field>
                    </v-flex -->
                </v-layout>
                <v-layout row wrap>
                    <v-flex xs12 text-xs-right class="pt-3">
                        <v-btn @click.stop="cancel" depressed color="grey lighten-3"
                               class="action-cancel">
                            <span>{{ label.cancel }}</span>
                        </v-btn>
                        <v-btn @click.stop="disable('AccShare')" depressed color="grey lighten-3"
                               class="action-disable" v-if="model.AccShare && scope === 'sharing'">
                            <span>{{ label.disable }}</span>
                        </v-btn>
                        <v-btn @click.stop="disable('AccSync')" depressed color="grey lighten-3"
                               class="action-disable" v-if="model.AccSync && scope === 'sync'">
                            <span>{{ label.disable }}</span>
                        </v-btn>
                        <v-btn color="blue-grey lighten-2" depressed dark @click.stop="enable('AccShare')"
                               class="action-enable" v-if="!model.AccShare && scope === 'sharing'">
                            <span>{{ label.enable }}</span>
                        </v-btn>
                        <v-btn color="blue-grey lighten-2" depressed dark @click.stop="enable('AccSync')"
                               class="action-enable" v-if="!model.AccSync && scope === 'sync'">
                            <span>{{ label.enable }}</span>
                        </v-btn>
                        <v-btn color="blue-grey lighten-2" depressed dark @click.stop="confirm"
                               class="action-confirm" v-if="model.AccShare && scope === 'sharing'">
                            <span>{{ label.save }}</span>
                        </v-btn>
                        <v-btn color="blue-grey lighten-2" depressed dark @click.stop="confirm"
                               class="action-confirm" v-if="model.AccSync && scope === 'sync'">
                            <span>{{ label.save }}</span>
                        </v-btn>
                        <v-btn color="blue-grey lighten-2" depressed dark @click.stop="confirm"
                               class="action-confirm" v-if="scope === 'account'">
                            <span>{{ label.save }}</span>
                        </v-btn>
                    </v-flex>
                </v-layout>
            </v-container>
        </v-card>
    </v-dialog>
</template>
<script>
    import options from "../resources/options";

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
                items: {
                    thumbs: thumbs,
                    sizes: this.sizes(thumbs),
                    expires: [
                        { "value": 0, "text": "Never" },
                        { "value": 86400, "text": "After 1 day" },
                        { "value": 86400 * 3, "text": "After 3 days" },
                        { "value": 86400 * 7, "text": "After 7 days" },
                        { "value": 86400 * 14, "text": "After two weeks" },
                        { "value": 86400 * 31, "text": "After one month" },
                        { "value": 86400 * 60, "text": "After two months" },
                        { "value": 86400 * 365, "text": "After one year" },
                    ],
                },
                readonly: this.$config.getValue("readonly"),
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
                    SharePath: this.$gettext("Folder"),
                    ShareSize: this.$gettext("Size"),
                    ShareExpires: this.$gettext("Expires"),
                    ShareExif: this.$gettext("Include metadata"),
                    ShareSidecar: this.$gettext("Include sidecar files"),
                    SyncPath: this.$gettext("Folder"),
                    SyncInterval: this.$gettext("Interval"),
                    SyncStart: this.$gettext("Start"),
                    SyncDownload: this.$gettext("Download new files"),
                    SyncUpload: this.$gettext("Upload local files"),
                    SyncDelete: this.$gettext("Remote delete"),
                    SyncRaw: this.$gettext("Sync RAW images"),
                    SyncVideo: this.$gettext("Sync videos"),
                    SyncSidecar: this.$gettext("Sync sidecar files"),
                }
            }
        },
        methods: {
            cancel() {
                this.$emit('cancel');
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
                this.loading = true;

                this.model.update().then(() => {
                    this.loading = false;
                    this.$emit('confirm');
                });
            },
            sizes(thumbs) {
                const result = [
                    {"text" : this.$gettext("Original"), "value": ""}
                ];

                for(let i in thumbs) {
                    const s = thumbs[i];
                    result.push({"text" : s["Width"] + 'x' + s["Height"], "value": s["Name"]});
                }

                return result;
            },
        },
        watch: {
            show: function (show) {
            }
        },
    }
</script>
