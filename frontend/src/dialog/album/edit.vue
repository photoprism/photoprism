<template>
  <v-dialog :model-value="show" persistent max-width="500" class="dialog-album-edit" color="background" @keydown.esc="close">
    <v-form ref="form" validate-on="blur" class="form-album-edit" accept-charset="UTF-8" @submit.prevent="confirm">
      <v-card>
        <v-card-title class="d-flex justify-start align-center ga-3">
          <h3 class="text-h5">
            <translate :translate-params="{ name: model.modelName() }">Edit %{name}</translate>
          </h3>
        </v-card-title>

        <v-card-text class="dense">
          <v-row dense>
            <v-col v-if="album.Type !== 'month'" cols="12">
              <v-text-field v-model="model.Title" hide-details autofocus :rules="[titleRule]" :label="$gettext('Name')" :disabled="disabled" class="input-title" @keyup.enter="confirm"></v-text-field>
            </v-col>
            <v-col cols="12">
              <v-text-field v-model="model.Location" hide-details :label="$gettext('Location')" :disabled="disabled" class="input-location"></v-text-field>
            </v-col>
            <v-col cols="12">
              <v-textarea :key="growDesc" v-model="model.Description" auto-grow hide-details autocomplete="off" :label="$gettext('Description')" :rows="1" :disabled="disabled" class="input-description"></v-textarea>
            </v-col>
            <v-col cols="12">
              <!-- TODO: check property return-masked-value TEST -->
              <!-- TODO: check property allow-overflow TEST -->
              <v-combobox v-model="model.Category" hide-details :search.sync="model.Category" :items="categories" :disabled="disabled" :label="$gettext('Category')" return-masked-value class="input-category"></v-combobox>
            </v-col>
            <v-col cols="12" sm="6">
              <v-select v-model="model.Order" :label="$gettext('Sort Order')" :menu-props="{ maxHeight: 400 }" hide-details :items="sorting" :disabled="disabled" item-value="value" item-title="text"></v-select>
            </v-col>
            <v-col sm="3">
              <!-- TODO: check property flat TEST -->
              <v-checkbox v-model="model.Favorite" :disabled="disabled" :label="$gettext('Favorite')" hide-details> </v-checkbox>
            </v-col>
            <v-col v-if="experimental && featPrivate" sm="3">
              <v-checkbox v-model="model.Private" :disabled="disabled" :label="$gettext('Private')" hide-details> </v-checkbox>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-btn variant="flat" color="button" class="action-cancel" @click.stop="close">
            <translate>Cancel</translate>
          </v-btn>
          <v-btn variant="flat" color="primary-button" class="action-confirm" :disabled="disabled" @click.stop="confirm">
            <translate>Save</translate>
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
    show: Boolean,
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
      categories: this.$config.albumCategories(),
      titleRule: (v) => v.length <= this.$config.get("clip") || this.$gettext("Name too long"),
    };
  },
  watch: {
    show: function (show) {
      if (show) {
        this.model = this.album.clone();
      }
    },
  },
  methods: {
    expand() {
      this.growDesc = !this.growDesc;
    },
    close() {
      this.$emit("close");
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
