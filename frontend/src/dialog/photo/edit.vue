<template>
  <v-dialog :model-value="show" fullscreen :scrim="false" scrollable persistent class="p-photo-edit-dialog" @click.stop @keydown.esc="close">
    <v-card tile color="background">
      <v-toolbar flat color="navigation" :density="$vuetify.display.smAndDown ? 'compact' : 'comfortable'">
        <v-btn icon class="action-close" @click.stop="close">
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
      <v-tabs v-model="active" elevation="0" class="form" :density="$vuetify.display.smAndDown ? 'compact' : 'default'">
        <v-tab id="tab-details" ripple>
          <v-icon v-if="$vuetify.display.smAndDown" :title="$gettext('Details')">mdi-pencil</v-icon>
          <template v-else>
            <v-icon :size="18" start>mdi-pencil</v-icon>
            <translate key="Details">Details</translate>
          </template>
        </v-tab>

        <v-tab id="tab-labels" ripple :disabled="!$config.feature('labels')">
          <v-icon v-if="$vuetify.display.smAndDown" :title="$gettext('Labels')">mdi-label</v-icon>
          <template v-else>
            <v-icon :size="18" start>mdi-label</v-icon>
            <translate key="Labels">Labels</translate>
            <v-badge v-if="model.Labels.length" color="surface-variant" inline :content="model.Labels.length"></v-badge>
          </template>
        </v-tab>

        <v-tab id="tab-people" :disabled="!$config.feature('people')" ripple>
          <v-icon v-if="$vuetify.display.smAndDown" :title="$gettext('People')">mdi-account-multiple</v-icon>
          <template v-else>
            <v-icon :size="18" start>mdi-account-multiple</v-icon>
            <translate key="People">People</translate>
            <v-badge v-if="model.Faces" color="surface-variant" inline :content="model.Faces"></v-badge>
          </template>
        </v-tab>

        <v-tab id="tab-files" ripple>
          <v-icon v-if="$vuetify.display.smAndDown" :title="$gettext('Files')">mdi-film</v-icon>
          <template v-else>
            <v-icon :size="18" start>mdi-film</v-icon>
            <translate key="Files">Files</translate>
            <v-badge v-if="model.Files.length" color="surface-variant" inline :content="model.Files.length"></v-badge>
          </template>
        </v-tab>

        <v-tab v-if="$config.feature('edit')" id="tab-info" ripple>
          <v-icon>mdi-cog</v-icon>
        </v-tab>
      </v-tabs>

      <v-tabs-window v-model="active" class="overflow-y-auto" style="height: 100%">
        <v-tabs-window-item>
          <p-tab-photo-details :key="uid" ref="details" :model="model" :uid="uid" @close="close" @prev="prev" @next="next"></p-tab-photo-details>
        </v-tabs-window-item>

        <v-tabs-window-item>
          <p-tab-photo-labels :key="uid" :model="model" :uid="uid" @close="close"></p-tab-photo-labels>
        </v-tabs-window-item>

        <v-tabs-window-item>
          <p-tab-photo-people :key="uid" :model="model" :uid="uid" @close="close"></p-tab-photo-people>
        </v-tabs-window-item>

        <v-tabs-window-item>
          <p-tab-photo-files :key="uid" :model="model" :uid="uid" @close="close"></p-tab-photo-files>
        </v-tabs-window-item>

        <v-tabs-window-item v-if="$config.feature('edit')">
          <p-tab-photo-info :key="uid" :model="model" :uid="uid" @close="close"></p-tab-photo-info>
        </v-tabs-window-item>
      </v-tabs-window>
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
