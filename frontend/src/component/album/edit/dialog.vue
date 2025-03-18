<template>
  <v-dialog
    :model-value="visible"
    persistent
    max-width="500"
    class="dialog-album-edit"
    color="background"
    @keydown.esc="close"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-form
      ref="form"
      validate-on="invalid-input"
      class="form-album-edit"
      accept-charset="UTF-8"
      @submit.prevent="confirm"
    >
      <v-card>
        <v-card-title class="d-flex justify-start align-center ga-3">
          <v-icon size="28" color="primary">mdi-bookmark</v-icon>
          <h6 class="text-h6">
            {{ $gettext(`Edit %{s}`, { s: model.modelName() }) }}
          </h6>
        </v-card-title>

        <v-card-text class="dense">
          <v-row align="center" dense>
            <v-col v-if="album.Type !== 'month'" cols="12">
              <v-text-field
                v-model="model.Title"
                hide-details
                autofocus
                :rules="[titleRule]"
                :label="$gettext('Name')"
                :disabled="disabled"
                class="input-title"
                @keyup.enter="confirm"
              ></v-text-field>
            </v-col>
            <v-col cols="12">
              <v-text-field
                v-model="model.Location"
                hide-details
                :label="$gettext('Location')"
                :disabled="disabled"
                class="input-location"
              ></v-text-field>
            </v-col>
            <v-col cols="12">
              <v-textarea
                v-model="model.Description"
                auto-grow
                hide-details
                autocomplete="off"
                :label="$gettext('Description')"
                :rows="1"
                :disabled="disabled"
                class="input-description"
              ></v-textarea>
            </v-col>
            <v-col cols="12">
              <v-combobox
                v-model="category"
                class="input-category"
                :search.sync="category"
                :items="categories"
                :disabled="disabled"
                :label="$gettext('Category')"
                hide-details
                @update:model-value="onChange"
              ></v-combobox>
            </v-col>
            <v-col cols="12" sm="6">
              <v-select
                v-model="model.Order"
                :label="$gettext('Sort Order')"
                :menu-props="{ maxHeight: 400 }"
                hide-details
                :items="sorting"
                :disabled="disabled"
                item-value="value"
                item-title="text"
              ></v-select>
            </v-col>
            <v-col sm="3">
              <v-checkbox
                v-model="model.Favorite"
                :disabled="disabled"
                :label="$gettext('Favorite')"
                density="comfortable"
                hide-details
              ></v-checkbox>
            </v-col>
            <v-col v-if="experimental && featPrivate" sm="3">
              <v-checkbox
                v-model="model.Private"
                :disabled="disabled"
                :label="$gettext('Private')"
                density="comfortable"
                hide-details
              ></v-checkbox>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions class="action-buttons">
          <v-btn variant="flat" color="button" class="action-cancel" @click.stop="close">
            {{ $gettext(`Cancel`) }}
          </v-btn>
          <v-btn variant="flat" color="highlight" class="action-confirm" :disabled="disabled" @click.stop="confirm">
            {{ $gettext(`Save`) }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-form>
  </v-dialog>
</template>
<script>
import Album from "model/album";

export default {
  name: "PAlbumEditDialog",
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    album: {
      type: Object,
      default: () => {},
    },
  },
  data() {
    return {
      featPrivate: this.$config.feature("private"),
      experimental: this.$config.get("experimental") && !this.$config.ce(),
      disabled: !this.$config.allow("albums", "manage"),
      model: new Album(),
      growDesc: false,
      loading: false,
      sorting: [
        { value: "newest", text: this.$gettext("Newest First") },
        { value: "oldest", text: this.$gettext("Oldest First") },
        { value: "added", text: this.$gettext("Recently Added") },
        { value: "title", text: this.$gettext("Picture Title") },
        { value: "name", text: this.$gettext("File Name") },
        { value: "size", text: this.$gettext("File Size") },
        { value: "duration", text: this.$gettext("Video Duration") },
        { value: "relevance", text: this.$gettext("Most Relevant") },
      ],
      category: null,
      categories: this.$config.albumCategories(),
      titleRule: (v) => v.length <= this.$config.get("clip") || this.$gettext("Name too long"),
    };
  },
  watch: {
    visible: function (show) {
      if (show) {
        this.model = this.album.clone();
        this.category = this.model.Category ? this.model.Category : null;
      }
    },
  },
  methods: {
    afterEnter() {
      this.$view.enter(this);
    },
    afterLeave() {
      this.$view.leave(this);
    },
    expand() {
      this.growDesc = !this.growDesc;
    },
    close() {
      this.$emit("close");
    },
    onChange() {
      if (this.category) {
        this.model.Category = this.category;
      } else {
        this.model.Category = "";
      }
    },
    confirm() {
      if (this.disabled) {
        this.close();
        return;
      }

      this.model.update().then((m) => {
        this.$notify.success(this.$gettext("Changes successfully saved"));
        this.categories = this.$config.albumCategories();
        this.$emit("close");
      });
    },
  },
};
</script>
