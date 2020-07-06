<template>
  <div class="p-tab p-tab-index">
    <v-form ref="form" class="p-photo-index" lazy-validation @submit.prevent="submit" dense>
      <v-container fluid>
        <p class="subheading">
          <span v-if="fileName">{{ action }} {{ fileName }}…</span>
          <span v-else-if="busy"><translate>Indexing media and sidecar files…</translate></span>
          <span v-else-if="completed"><translate>Done.</translate></span>
          <span v-else><translate>Press button to start indexing…</translate></span>
        </p>

        <v-autocomplete
                @change="onChange"
                color="secondary-dark"
                class="my-3 input-index-folder"
                hide-details hide-no-data flat solo
                v-model="settings.index.path"
                browser-autocomplete="off"
                :items="dirs"
                :loading="loading"
                :disabled="busy || loading"
                item-text="name"
                item-value="path"
        >
        </v-autocomplete>

        <p class="options">
          <v-progress-linear color="secondary-dark" :value="completed"
                             :indeterminate="busy"></v-progress-linear>
        </p>

        <v-layout wrap align-top class="pb-3">
          <v-flex xs12 sm6 lg4 class="px-2 pb-2 pt-2">
            <v-checkbox
                    @change="onChange"
                    :disabled="busy"
                    class="ma-0 pa-0"
                    v-model="settings.index.rescan"
                    color="secondary-dark"
                    :label="labels.rescan"
                    :hint="hints.rescan"
                    prepend-icon="cached"
                    persistent-hint
            >
            </v-checkbox>
          </v-flex>
        </v-layout>

        <v-btn
                :disabled="!busy"
                color="secondary-dark"
                class="white--text ml-0 mt-2 action-cancel"
                depressed
                @click.stop="cancelIndexing()"
        >
          <translate>Cancel</translate>
        </v-btn>

        <v-btn
                :disabled="busy"
                color="secondary-dark"
                class="white--text ml-0 mt-2 action-index"
                depressed
                @click.stop="startIndexing()"
        >
          <translate>Start</translate>
          <v-icon right dark>update</v-icon>
        </v-btn>

        <v-alert
                :value="true"
                color="error"
                icon="priority_high"
                class="mt-3"
                outline
                v-if="config.count.hidden > 1"
        >
          <translate :translate-params="{n: config.count.hidden}">The index currently contains %{n} hidden files.</translate>
          <translate>Their format may not be supported, they haven't been converted to JPEG yet or there are duplicates.</translate>
        </v-alert>
      </v-container>
    </v-form>
  </div>
</template>

<script>
    import Api from "common/api";
    import Axios from "axios";
    import Notify from "common/notify";
    import Event from "pubsub-js";
    import Settings from "model/settings";
    import Util from "common/util";
    import {Folder, RootOriginals} from "model/folder";

    export default {
        name: 'p-tab-index',
        data() {
            const root = {"path": "/", "name": this.$gettext("All originals")}

            return {
                settings: new Settings(this.$config.settings()),
                readonly: this.$config.get("readonly"),
                config: this.$config.values,
                started: false,
                busy: false,
                loading: false,
                completed: 0,
                subscriptionId: "",
                action: "",
                fileName: "",
                source: null,
                root: root,
                dirs: [root],
                labels: {
                    rescan: this.$gettext("Complete Rescan"),
                    convert: this.$gettext("Convert to JPEG"),
                    path: this.$gettext("Folder"),
                },
                hints: {
                    rescan: this.$gettext("Re-index all originals, including already indexed and unchanged files."),
                    convert: this.$gettext("File types like RAW might need to be converted so that they can be displayed in a browser. JPEGs will be stored in the same folder next to the original using the best possible quality."),
                }
            }
        },
        methods: {
            onChange() {
                this.settings.save();
            },
            submit() {
                // DO NOTHING
            },
            cancelIndexing() {
                Api.delete('index');
            },
            startIndexing() {
                this.source = Axios.CancelToken.source();
                this.started = Date.now();
                this.busy = true;
                this.completed = 0;
                this.fileName = '';

                const ctx = this;
                Notify.blockUI();

                Api.post('index', this.settings.index, {cancelToken: this.source.token}).then(function () {
                    Notify.unblockUI();
                    ctx.busy = false;
                    ctx.completed = 100;
                    ctx.fileName = '';
                }).catch(function (e) {
                    Notify.unblockUI();

                    if (Axios.isCancel(e)) {
                        // run in background
                        return
                    }

                    Notify.error(ctx.$gettext("Indexing failed"));

                    ctx.busy = false;
                    ctx.completed = 0;
                    ctx.fileName = '';
                });
            },
            handleEvent(ev, data) {
                if (this.source) {
                    this.source.cancel('run in background');
                    this.source = null;
                    Notify.unblockUI();
                }

                const type = ev.split('.')[1];

                switch (type) {
                    case "indexing":
                        this.action = this.$gettext("Indexing");
                        this.busy = true;
                        this.completed = 0;
                        this.fileName = data.fileName;
                        break;
                    case "converting":
                        this.action = this.$gettext("Converting");
                        this.busy = true;
                        this.completed = 0;
                        this.fileName = data.fileName;
                        break;
                    case "thumbnails":
                        this.action = this.$gettext("Creating thumbnails for");
                        this.busy = true;
                        this.completed = 0;
                        this.fileName = data.fileName;
                        break;
                    case 'completed':
                        this.action = "";
                        this.busy = false;
                        this.completed = 100;
                        this.fileName = '';
                        break;
                    default:
                        console.log(data)
                }
            },
        },
        created() {
            this.subscriptionId = Event.subscribe('index', this.handleEvent);
            this.loading = true;

            Folder.findAll(RootOriginals).then((r) => {
                const folders = r.models ? r.models : [];
                const currentPath = this.settings.index.path;
                let found = currentPath === this.root.path;

                this.dirs = [this.root];

                for (let i = 0; i < folders.length; i++) {
                    if (currentPath === folders[i].Path) {
                        found = true;
                    }

                    this.dirs.push({path: folders[i].Path, name: "/" + Util.truncate(folders[i].Path, 100, "…")});
                }

                if (!found) {
                    this.settings.index.path = this.root.path;
                }
            }).finally(() => this.loading = false);
        },
        destroyed() {
            Event.unsubscribe(this.subscriptionId);
        },
    };
</script>
