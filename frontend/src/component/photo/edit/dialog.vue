<template>
  <v-dialog
    ref="dialog"
    :model-value="visible"
    :fullscreen="$vuetify.display.smAndDown"
    scrim
    scrollable
    class="p-dialog p-photo-edit-dialog v-dialog--sidepanel"
    @click.stop="onClick"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-card ref="content" :tile="$vuetify.display.smAndDown" tabindex="1">
      <v-toolbar flat color="navigation" :density="$vuetify.display.smAndDown ? 'compact' : 'comfortable'">
        <v-btn icon class="action-close" @click.stop="onClose">
          <v-icon>mdi-close</v-icon>
        </v-btn>

        <v-toolbar-title
          >{{ title }}
          <v-icon v-if="isPrivate" title="Private">mdi-lock</v-icon>
        </v-toolbar-title>

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
      <v-tabs v-model="active" class="elevation-0" :density="$vuetify.display.smAndDown ? 'comfortable' : 'default'">
        <v-tab id="tab-details" value="details" ripple>
          <v-icon v-if="$vuetify.display.smAndDown" :title="$gettext('Details')">mdi-pencil</v-icon>
          <template v-else>
            <v-icon :size="18" start>mdi-pencil</v-icon>
            {{ $gettext(`Details`) }}
          </template>
        </v-tab>

        <v-tab id="tab-labels" value="labels" ripple :disabled="!$config.feature('labels')">
          <v-icon v-if="$vuetify.display.smAndDown" :title="$gettext('Labels')">mdi-label</v-icon>
          <template v-else>
            <v-icon :size="18" start>mdi-label</v-icon>
            {{ $gettext(`Labels`) }}
          </template>
          <v-badge v-if="model.Labels.length" color="surface-variant" inline :content="model.Labels.length"></v-badge>
        </v-tab>

        <v-tab id="tab-people" value="people" :disabled="!$config.feature('people')" ripple>
          <v-icon v-if="$vuetify.display.smAndDown" :title="$gettext('People')">mdi-account-multiple</v-icon>
          <template v-else>
            <v-icon :size="18" start>mdi-account-multiple</v-icon>
            {{ $gettext(`People`) }}
          </template>
          <v-badge v-if="model.Faces" color="surface-variant" inline :content="model.Faces"></v-badge>
        </v-tab>

        <v-tab id="tab-files" value="files" ripple>
          <v-icon v-if="$vuetify.display.smAndDown" :title="$gettext('Files')">mdi-film</v-icon>
          <template v-else>
            <v-icon :size="18" start>mdi-film</v-icon>
            {{ $gettext(`Files`) }}
          </template>
          <v-badge v-if="model.Files.length" color="surface-variant" inline :content="model.Files.length"></v-badge>
        </v-tab>

        <v-tab v-if="canEdit" id="tab-info" value="info" ripple>
          <v-icon>mdi-cog</v-icon>
        </v-tab>
      </v-tabs>

      <v-tabs-window v-if="ready" v-model="active">
        <v-tabs-window-item value="details">
          <p-tab-photo-details ref="details" :uid="uid" @close="close" @prev="prev" @next="next"></p-tab-photo-details>
        </v-tabs-window-item>

        <v-tabs-window-item value="labels">
          <p-tab-photo-labels :uid="uid" @close="close"></p-tab-photo-labels>
        </v-tabs-window-item>

        <v-tabs-window-item value="people">
          <p-tab-photo-people :uid="uid" @close="close"></p-tab-photo-people>
        </v-tabs-window-item>

        <v-tabs-window-item value="files">
          <p-tab-photo-files :uid="uid" @close="close"></p-tab-photo-files>
        </v-tabs-window-item>

        <v-tabs-window-item v-if="canEdit" value="info">
          <p-tab-photo-info :uid="uid" @close="close"></p-tab-photo-info>
        </v-tabs-window-item>
      </v-tabs-window>
    </v-card>
  </v-dialog>
</template>
<script>
import Photo from "model/photo";

import PhotoDetails from "component/photo/edit/details.vue";
import PhotoLabels from "component/photo/edit/labels.vue";
import PhotoPeople from "component/photo/edit/people.vue";
import PhotoFiles from "component/photo/edit/files.vue";
import PhotoInfo from "component/photo/edit/info.vue";

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
    visible: {
      type: Boolean,
      default: false,
    },
    index: {
      type: Number,
      default: 0,
    },
    selection: {
      type: Array,
      default: () => [],
    },
    album: {
      type: Object,
      default: () => {},
    },
    tab: {
      type: String,
      default: "",
    },
  },
  data() {
    return {
      selected: 0,
      selectedId: "",
      model: new Photo(),
      uid: "",
      loading: false,
      ready: false,
      search: null,
      items: [],
      canEdit: this.$config.feature("edit"),
      readonly: this.$config.get("readonly"),
      active: this.tab,
      rtl: this.$isRtl,
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
      if (this.model && this.model.Private && this.$config.getSettings().features.private) {
        return this.model.Private;
      }

      return false;
    },
  },
  watch: {
    visible: function (show) {
      if (show) {
        if (this.tab) {
          this.active = this.tab;
        }
        this.find(this.index);
      }
    },
  },
  created() {
    this.subscriptions.push(this.$event.subscribe("photos.updated", (ev, data) => this.onUpdate(ev, data)));
  },
  beforeUnmount() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      this.$event.unsubscribe(this.subscriptions[i]);
    }
  },
  methods: {
    afterEnter() {
      this.$view.enter(this);
      this.ready = true;
    },
    afterLeave() {
      this.ready = false;
      this.$view.leave(this);
    },
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
              this.model.setValues({ Title: values.Title, Caption: values.Caption }, true);
            }
          }
          break;
      }
    },
    onClick(ev) {
      // Closes dialog when user clicks on background and model data is unchanged.
      if (!ev || !ev?.target?.classList?.contains("v-overlay__scrim")) {
        return;
      }
      ev.preventDefault();
      this.onClose();
    },
    onClose() {
      // Closes the dialog only if model data is unchanged.
      if (this.model?.hasId() && this.model?.wasChanged()) {
        this.$refs?.dialog?.animateClick();
      } else {
        this.close();
      }
    },
    close() {
      // Closes the dialog.
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
        return Promise.reject();
      }

      if (!this.selection || !this.selection[index]) {
        this.$notify.error(this.$gettext("Invalid photo selected"));
        return Promise.reject();
      }

      this.loading = true;
      this.selected = index;
      this.selectedId = this.selection[index];

      return this.model
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
