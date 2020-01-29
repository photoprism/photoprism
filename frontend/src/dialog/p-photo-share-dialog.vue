<template>
    <v-dialog fullscreen hide-overlay scrollable lazy
              v-model="show" persistent class="p-photo-share-dialog" @keydown.esc="cancel">
        <v-card color="application">
            <v-toolbar dark color="navigation">
                <v-btn icon dark @click.stop="cancel">
                    <v-icon>close</v-icon>
                </v-btn>
                <v-toolbar-title>Share Photo</v-toolbar-title>
                <v-spacer></v-spacer>
                <v-toolbar-items>
                    <v-btn icon disabled>
                        <v-icon>navigate_before</v-icon>
                    </v-btn>

                    <v-btn icon disabled>
                        <v-icon>navigate_next</v-icon>
                    </v-btn>
                </v-toolbar-items>
            </v-toolbar>
            <v-container grid-list-xs text-xs-center fluid>
                <v-form lazy-validation dense
                        ref="form" class="p-form-photo-edit-meta" accept-charset="UTF-8"
                        @submit.prevent="confirm">
                    <v-layout row wrap align-center fill-height>
                        <v-flex
                                class="p-photo pa-2"
                                xs12 sm4 md2
                        >
                            <v-card tile
                                    class="ma-1 elevation-0"
                                    :title="model.PhotoTitle">
                                <v-img :src="model.getThumbnailUrl('tile_500')"
                                       aspect-ratio="1"
                                       class="accent lighten-2 elevation-0"
                                       style="cursor: pointer"
                                       @click.exact="openPhoto()"
                                >
                                    <v-layout
                                            slot="placeholder"
                                            fill-height
                                            align-center
                                            justify-center
                                            ma-0
                                    >
                                        <v-progress-circular indeterminate
                                                             color="accent lighten-5"></v-progress-circular>
                                    </v-layout>
                                </v-img>

                            </v-card>
                        </v-flex>
                        <v-flex xs12 sm8 md10 fill-height>
                            <v-layout row wrap>
                                <v-flex xs12 class="pa-2">
                                    <p class="subheading pb-3">
                                        This is a very first draft for a share dialog. Feedback and contributions welcome.
                                    </p>
                                </v-flex>
                            </v-layout>
                        </v-flex>
                    </v-layout>
                </v-form>

            </v-container>
        </v-card>
    </v-dialog>
</template>
<script>
    import Photo from "../model/photo";
    import PhotoShareTodo from "./photo-share/todo.vue";

    export default {
        name: 'p-photo-edit-dialog',
        props: {
            show: Boolean,
            selection: Array,
            album: Object,
        },
        components: {
            'p-tab-photo-share-todo': PhotoShareTodo,
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
            find(index) {
                if (this.loading) {
                    return;
                }

                if(!this.selection || !this.selection[index]) {
                    this.$notify.error("Invalid photo selected");
                    return
                }

                this.loading = true;
                this.selected = index;
                this.selectedId = this.selection[index];

                this.model.find(this.selectedId).then(model => {
                    model.refreshFileAttr();
                    this.model = model;
                    this.$refs.meta.refresh(model);
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
