<template>
    <v-dialog fullscreen hide-overlay scrollable lazy
              v-model="show" persistent class="p-photo-edit-dialog" @keydown.esc="cancel">
        <v-card color="application">
            <v-toolbar dark color="navigation">
                <v-btn icon dark @click.stop="cancel">
                    <v-icon>close</v-icon>
                </v-btn>
                <v-toolbar-title>{{ title }}</v-toolbar-title>
                <v-spacer></v-spacer>
                <v-toolbar-items>
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
                <v-tab-item>
                    <p-tab-photo-edit-details :model="model" ref="details" @cancel="cancel"></p-tab-photo-edit-details>
                </v-tab-item>

                <v-tab id="tab-edit-labels" ripple>
                    <translate>Labels</translate>
                </v-tab>
                <v-tab-item lazy>
                    <p-tab-photo-edit-labels :model="model"></p-tab-photo-edit-labels>
                </v-tab-item>

                <v-tab id="tab-edit-files" ripple>
                    <translate>Files</translate>
                </v-tab>
                <v-tab-item lazy>
                    <p-tab-photo-edit-files :model="model"></p-tab-photo-edit-files>
                </v-tab-item>
            </v-tabs>
        </v-card>
    </v-dialog>
</template>
<script>
    import Photo from "../model/photo";
    import PhotoEditDetails from "./photo-edit/details.vue";
    import PhotoEditLabels from "./photo-edit/labels.vue";
    import PhotoEditFiles from "./photo-edit/files.vue";

    export default {
        name: 'p-photo-edit-dialog',
        props: {
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
            title: function() {
                if(this.model && this.model.PhotoTitle) {
                    return this.model.PhotoTitle
                }

                this.$gettext("Edit Photo");
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
                readonly: this.$config.getValue("readonly"),
                active: this.tab,
            }
        },
        methods: {
            changePath: function (path) {
                /* if (this.$route.path !== path) {
                    this.$router.replace(path)
                } */
            },
            cancel() {
                this.$emit('cancel');
            },
            confirm() {
                this.$emit('confirm');
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
                    this.find(0);
                }
            }
        },
    }
</script>
