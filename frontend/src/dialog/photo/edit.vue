<template>
  <v-dialog :value="show" fullscreen hide-overlay scrollable persistent class="p-photo-edit-dialog" @keydown.esc="close">
    <v-card color="application">
      <v-toolbar theme="dark" flat color="navigation" :dense="$vuetify.display.smAndDown">
        <v-btn icon theme="dark" class="action-close" @click.stop="close">
          <v-icon>mdi-close</v-icon>
        </v-btn>
        <v-toolbar-title
          >{{ title }}
          <v-icon v-if="isPrivate" title="Private">mdi-lock</v-icon>
        </v-toolbar-title>
        <v-spacer></v-spacer>
        <v-toolbar-items v-if="selection.length > 1">
          <v-btn icon :disabled="selected < 1" class="action-previous" @click.stop="prev">
            <v-icon v-if="!rtl">mdi-chevron-left</v-icon>
            <v-icon v-else>mdi-chevron-right</v-icon>
          </v-btn>

          <v-btn icon :disabled="selected >= selection.length - 1" class="action-next" @click.stop="next">
            <v-icon v-if="!rtl">mdi-chevron-right</v-icon>
            <v-icon v-else>mdi-chevron-left</v-icon>
          </v-btn>
        </v-toolbar-items>
    </v-toolbar>
      <v-tabs v-model="active" elevation="0" grow class="form transparent" background-color="secondary" slider-color="primary-dark" :height="$vuetify.display.smAndDown ? 48 : 64">
        <v-tab id="tab-details" ripple>
          <v-icon v-if="$vuetify.display.smAndDown" :title="$gettext('Details')">mdi-pencil</v-icon>
          <template v-else>
            <v-icon :size="18" :left="!rtl" :right="rtl">mdi-pencil</v-icon>
            <translate key="Details">Details</translate>
          </template>
        </v-tab>

        <v-tab id="tab-labels" ripple :disabled="!$config.feature('labels')">
          <v-icon v-if="$vuetify.display.smAndDown" :title="$gettext('Labels')">mdi-label</v-icon>
          <template v-else>
            <v-icon :size="18" :left="!rtl" :right="rtl">mdi-label</v-icon>
            <v-badge color="secondary-dark" :left="rtl" :right="!rtl">
              <template #badge>
                <span v-if="model.Labels.length">{{ model.Labels.length }}</span>
              </template>
              <translate key="Labels">Labels</translate>
            </v-badge>
          </template>
        </v-tab>

        <v-tab id="tab-people" :disabled="!$config.feature('people')" ripple>
          <v-icon v-if="$vuetify.display.smAndDown" :title="$gettext('People')">mdi-account-multiple</v-icon>
          <template v-else>
            <v-icon :size="18" :left="!rtl" :right="rtl">mdi-account-multiple</v-icon>
            <v-badge color="secondary-dark" :left="rtl" :right="!rtl">
              <template #badge>
                <span v-if="model.Faces">{{ model.Faces }}</span>
              </template>
              <translate key="People">People</translate>
            </v-badge>
          </template>
        </v-tab>

        <v-tab id="tab-files" ripple>
          <v-icon v-if="$vuetify.display.smAndDown" :title="$gettext('Files')">mdi-film</v-icon>
          <template v-else>
            <v-icon :size="18" :left="!rtl" :right="rtl">mdi-film</v-icon>
            <v-badge color="secondary-dark" :left="rtl" :right="!rtl">
              <template #badge>
                <span v-if="model.Files.length">{{ model.Files.length }}</span>
              </template>
              <translate key="Files">Files</translate>
            </v-badge>
          </template>
        </v-tab>

        <v-tab v-if="$config.feature('edit')" id="tab-info" ripple>
          <v-icon>mdi-cog</v-icon>
        </v-tab>

        <!-- TODO: check property touchless TEST -->
        <v-tabs-items touchless v-model="active">
          <v-tab-item>
            <p-tab-photo-details :key="uid" ref="details" :model="model" :uid="uid" @close="close" @prev="prev" @next="next"></p-tab-photo-details>
          </v-tab-item>

          <v-tab-item>
            <p-tab-photo-labels :key="uid" :model="model" :uid="uid" @close="close"></p-tab-photo-labels>
          </v-tab-item>

          <v-tab-item>
            <p-tab-photo-people :key="uid" :model="model" :uid="uid" @close="close"></p-tab-photo-people>
          </v-tab-item>

          <v-tab-item>
            <p-tab-photo-files :key="uid" :model="model" :uid="uid" @close="close"></p-tab-photo-files>
          </v-tab-item>

          <v-tab-item v-if="$config.feature('edit')">
            <p-tab-photo-info :key="uid" :model="model" :uid="uid" @close="close"></p-tab-photo-info>
          </v-tab-item>
        </v-tabs-items>
      </v-tabs>
    </v-card>
  </v-dialog>
</template>
<script>
import Photo from "model/photo";
import PhotoDetails from "./edit/details.vue";
import PhotoLabels from "./edit/labels.vue";
import PhotoPeople from "./edit/people.vue";
import PhotoFiles from "./edit/files.vue";
import PhotoInfo from "./edit/info.vue";
import Event from "pubsub-js";

export default {
  name: "PPhotoEditDialog",
  components: {
    "p-tab-photo-details": PhotoDetails,
    "p-tab-photo-labels": PhotoLabels,
    "p-tab-photo-people": PhotoPeople,
    "p-tab-photo-files": PhotoFiles,
    "p-tab-photo-info": PhotoInfo,
  },
  props: {
    index: {
      type: Number,
      default: 0,
    },
    show: Boolean,
    selection: {
      type: Array,
      default: () => [],
    },
    album: {
      type: Object,
      default: () => {},
    },
  },
  data() {
    return {
      selected: 0,
      selectedId: "",
      model: new Photo(),
      uid: "",
      loading: false,
      search: null,
      items: [],
      readonly: this.$config.get("readonly"),
      active: this.tab,
      rtl: this.$rtl,
      subscriptions: [],
    };
  },
  computed: {
    title() {
      if (this.model && this.model.Title) {
        return this.model.Title;
      }

      return this.$gettext("Edit Photo");
    },
    isPrivate() {
      if (this.model && this.model.Private && this.$config.settings().features.private) {
        return this.model.Private;
      }

      return false;
    },
  },
  watch: {
    show: function (show) {
      if (show) {
        this.find(this.index);
      }
    },
  },
  created() {
    this.subscriptions.push(Event.subscribe("photos.updated", (ev, data) => this.onUpdate(ev, data)));
  },
  unmounted() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      Event.unsubscribe(this.subscriptions[i]);
    }
  },
  methods: {
    onUpdate(ev, data) {
      if (!data || !data.entities || !Array.isArray(data.entities) || this.loading || !this.model || !this.model.UID) {
        return;
      }

      const type = ev.split(".")[1];

      switch (type) {
        case "updated":
          for (let i = 0; i < data.entities.length; i++) {
            const values = data.entities[i];
            if (values.UID && values.Title && this.model.UID === values.UID) {
              this.model.setValues({ Title: values.Title, Description: values.Description }, true);
            }
          }
          break;
      }
    },
    close() {
      this.$emit("close");
    },
    prev() {
      if (this.selected > 0) {
        this.find(this.selected - 1);
      }
    },
    next() {
      if (!this.selection) return;

      if (this.selected < this.selection.length) {
        this.find(this.selected + 1);
      }
    },
    find(index) {
      if (this.loading) {
        return;
      }

      if (!this.selection || !this.selection[index]) {
        this.$notify.error(this.$gettext("Invalid photo selected"));
        return;
      }

      this.loading = true;
      this.selected = index;
      this.selectedId = this.selection[index];

      this.model
        .find(this.selectedId)
        .then((model) => {
          model.refreshFileAttr();
          this.model = model;
          this.loading = false;
          this.uid = this.selectedId;
        })
        .catch(() => (this.loading = false));
    },
  },
};
</script>
