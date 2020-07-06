<template>
  <v-dialog fullscreen hide-overlay scrollable lazy
            v-model="show" persistent class="p-photo-edit-dialog" @keydown.esc="close">
    <v-card color="application">
      <v-toolbar dark flat color="navigation">
        <v-btn icon dark @click.stop="close" class="action-close">
          <v-icon>close</v-icon>
        </v-btn>
        <v-toolbar-title>{{ title }}
          <v-icon v-if="isPrivate" title="Private">lock</v-icon>
        </v-toolbar-title>
        <v-spacer></v-spacer>
        <v-toolbar-items v-if="selection.length > 1">
          <v-btn icon disabled @click.stop="prev" :disabled="selected < 1" class="action-previous">
            <v-icon>navigate_before</v-icon>
          </v-btn>

          <v-btn icon @click.stop="next" :disabled="selected >= selection.length - 1" class="action-next">
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
        <v-tab id="tab-details" ripple>
          <translate key="Details">Details</translate>
        </v-tab>

        <v-tab id="tab-labels" ripple :disabled="!$config.feature('labels')">
          <translate key="Labels">Labels</translate>
        </v-tab>

        <v-tab id="tab-files" ripple>
          <translate key="Files">Files</translate>
        </v-tab>

        <v-tab id="tab-info" ripple v-if="$config.feature('edit')">
          <v-icon>settings</v-icon>
        </v-tab>

        <v-tabs-items touchless>
          <v-tab-item>
            <p-tab-photo-details :model="model" :uid="uid" :key="uid" ref="details"
                                 @close="close" @prev="prev" @next="next"></p-tab-photo-details>
          </v-tab-item>

          <v-tab-item lazy>
            <p-tab-photo-labels :model="model" :uid="uid" :key="uid" @close="close"></p-tab-photo-labels>
          </v-tab-item>

          <v-tab-item lazy>
            <p-tab-photo-files :model="model" :uid="uid" :key="uid" @close="close"></p-tab-photo-files>
          </v-tab-item>

          <v-tab-item lazy v-if="$config.feature('edit')">
            <p-tab-photo-info :model="model" :uid="uid"  :key="uid" @close="close"></p-tab-photo-info>
          </v-tab-item>
        </v-tabs-items>
      </v-tabs>
    </v-card>
  </v-dialog>
</template>
<script>
    import Photo from "model/photo";
    import PhotoDetails from "./details.vue";
    import PhotoLabels from "./labels.vue";
    import PhotoFiles from "./files.vue";
    import PhotoInfo from "./info.vue";

    export default {
        name: 'p-photo-edit-dialog',
        props: {
            index: Number,
            show: Boolean,
            selection: Array,
            album: Object,
        },
        components: {
            'p-tab-photo-details': PhotoDetails,
            'p-tab-photo-labels': PhotoLabels,
            'p-tab-photo-files': PhotoFiles,
            'p-tab-photo-info': PhotoInfo,
        },
        computed: {
            title: function () {
                if (this.model && this.model.Title) {
                    return this.model.Title
                }

                this.$gettext("Edit Photo");
            },
            isPrivate: function () {
                if (this.model && this.model.Private && this.$config.settings().features.private) {
                    return this.model.Private
                }

                return false;
            },
        },
        data() {
            return {
                selected: 0,
                selectedId: "",
                model: new Photo,
                uid: "",
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
                if(!this.selection) return;

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
                    this.loading = false;
                    this.uid = this.selectedId;
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
