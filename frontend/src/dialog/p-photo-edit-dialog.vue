<template>
    <v-dialog fullscreen hide-overlay scrollable lazy
              v-model="show" persistent class="p-photo-edit-dialog" @keydown.esc="close">
        <v-card color="application">
            <v-toolbar dark color="navigation">
                <v-btn icon dark @click.stop="close">
                    <v-icon>close</v-icon>
                </v-btn>
                <v-toolbar-title>{{ title }}
                    <v-icon v-if="isPrivate" title="Private">lock</v-icon>
                </v-toolbar-title>
                <v-spacer></v-spacer>
                <v-toolbar-items v-if="selection.length > 1">
                    <v-btn icon disabled @click.stop="prev" :disabled="selected < 1">
                        <v-icon>navigate_before</v-icon>
                    </v-btn>

                    <v-btn icon @click.stop="next" :disabled="selected >= selection.length - 1">
                        <v-icon>navigate_next</v-icon>
                    </v-btn>
                </v-toolbar-items>
            </v-toolbar>
            <v-tabs
                    v-model="active"
                    flat
                    grow
                    color="secondary"
                    slider-color="secondary-dark"
                    height="64"
                    class="form"
            >
                <v-tab id="tab-edit-details" ripple>
                    <translate>Details</translate>
                </v-tab>

                <v-tab id="tab-edit-labels" ripple :disabled="!$config.feature('labels')">
                    <translate>Labels</translate>
                </v-tab>

                <v-tab id="tab-edit-files" ripple>
                    <translate>Files</translate>
                </v-tab>

                <v-tabs-items touchless>
                    <v-tab-item>
                        <p-tab-photo-edit-details :model="model" ref="details"
                                                  @close="close" @prev="prev" @next="next"></p-tab-photo-edit-details>
                    </v-tab-item>

                    <v-tab-item lazy>
                        <p-tab-photo-edit-labels :model="model" @close="close"></p-tab-photo-edit-labels>
                    </v-tab-item>

                    <v-tab-item lazy>
                        <p-tab-photo-edit-files :model="model" @close="close"></p-tab-photo-edit-files>
                    </v-tab-item>
                </v-tabs-items>
            </v-tabs>
        </v-card>
    </v-dialog>
</template>
<script>
    import Photo from "../model/photo";
    import PhotoEditDetails from "./photo/details.vue";
    import PhotoEditLabels from "./photo/labels.vue";
    import PhotoEditFiles from "./photo/files.vue";

    export default {
        name: 'p-photo-edit-dialog',
        props: {
            index: Number,
            show: Boolean,
            selection: Array,
            album: Object,
        },
        components: {
            'p-tab-photo-edit-details': PhotoEditDetails,
            'p-tab-photo-edit-labels': PhotoEditLabels,
            'p-tab-photo-edit-files': PhotoEditFiles,
        },
        computed: {
            title: function () {
                if (this.model && this.model.PhotoTitle) {
                    return this.model.PhotoTitle
                }

                this.$gettext("Edit Photo");
            },
            isPrivate: function () {
                if (this.model && this.model.PhotoPrivate && this.$config.settings().features.private) {
                    return this.model.PhotoPrivate
                }

                return false;
            },
        },
        data() {
            return {
                selected: 0,
                selectedId: "",
                model: new Photo,
                loading: false,
                search: null,
                items: [],
                readonly: this.$config.get("readonly"),
                active: this.tab,
            }
        },
        methods: {
            changePath: function (path) {
                /* if (this.$route.path !== path) {
                    this.$router.replace(path)
                } */
            },
            close() {
                this.$emit('close');
            },
            prev() {
                if (this.selected > 0) {
                    this.find(this.selected - 1);
                }
            },
            next() {
                if (this.selected < this.selection.length) {
                    this.find(this.selected + 1);
                }
            },
            find(index) {
                if (this.loading) {
                    return;
                }

                if (!this.selection || !this.selection[index]) {
                    this.$notify.error("Invalid photo selected");
                    return
                }

                this.loading = true;
                this.selected = index;
                this.selectedId = this.selection[index];

                this.model.find(this.selectedId).then(model => {
                    model.refreshFileAttr();
                    this.model = model;
                    this.$refs.details.refresh(model);
                    this.loading = false;
                }).catch(() => this.loading = false);
            },
        },
        watch: {
            show: function (show) {
                if (show) {
                    this.find(this.index);
                }
            }
        },
    }
</script>
