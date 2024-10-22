<template>
  <v-dialog :value="show" persistent max-width="500" class="dialog-album-edit" color="application" @keydown.esc="close">
    <v-form ref="form" lazy-validation class="form-album-edit" accept-charset="UTF-8" @submit.prevent="confirm">
      <v-card raised elevation="24">
        <v-card-title class="pb-0">
          <v-row>
            <v-col cols="12">
              <h3 class="text-h5 mx-2 mb-0">
                <translate :translate-params="{ name: model.modelName() }">Edit %{name}</translate>
              </h3>
            </v-col>
          </v-row>
        </v-card-title>

        <v-card-text>
          <v-container fluid class="pa-0">
            <v-row>
              <v-col v-if="album.Type !== 'month'" cols="12" class="pa-2">
                <v-text-field v-model="model.Title" hide-details autofocus filled flat :rules="[titleRule]" :label="$gettext('Name')" :disabled="disabled" color="secondary-dark" class="input-title" @keyup.enter="confirm"></v-text-field>
              </v-col>
              <v-col cols="12" class="pa-2">
                <v-text-field v-model="model.Location" hide-details filled flat :label="$gettext('Location')" :disabled="disabled" color="secondary-dark" class="input-location"></v-text-field>
              </v-col>
              <v-col cols="12" class="pa-2">
                <v-textarea :key="growDesc" v-model="model.Description" auto-grow hide-details filled flat autocomplete="off" :label="$gettext('Description')" :rows="1" :disabled="disabled" class="input-description" color="secondary-dark"></v-textarea>
              </v-col>
              <v-col cols="12" class="pa-2">
                <!-- TODO: check property return-masked-value TEST -->
                <!-- TODO: check property allow-overflow TEST -->
                <v-combobox v-model="model.Category" hide-details filled flat :search-input.sync="model.Category" :items="categories" :disabled="disabled" :label="$gettext('Category')" :allow-overflow="false" return-masked-value color="secondary-dark" class="input-category"></v-combobox>
              </v-col>
              <v-col cols="12" sm="6" class="pa-2">
                <v-select v-model="model.Order" :label="$gettext('Sort Order')" :menu-props="{ maxHeight: 400 }" hide-details filled flat :items="sorting" :disabled="disabled" item-value="value" item-text="text" color="secondary-dark"></v-select>
              </v-col>
              <v-col sm="3" class="pa-2">
                <!-- TODO: check property flat TEST -->
                <v-checkbox v-model="model.Favorite" :disabled="disabled" color="secondary-dark" :label="$gettext('Favorite')" hide-details flat> </v-checkbox>
              </v-col>
              <v-col v-if="experimental && featPrivate" sm="3" class="pa-2">
                <v-checkbox v-model="model.Private" :disabled="disabled" color="secondary-dark" :label="$gettext('Private')" hide-details flat> </v-checkbox>
              </v-col>
            </v-row>
          </v-container>
        </v-card-text>
        <v-card-actions class="pt-0 px-6">
          <v-row class="pa-2">
            <v-col cols="12" class="text-right">
              <v-btn variant="flat" color="secondary-light" class="action-cancel" @click.stop="close">
                <translate>Cancel</translate>
              </v-btn>
              <v-btn variant="flat" theme="dark" color="primary-button" class="action-confirm" :disabled="disabled" @click.stop="confirm">
                <translate>Save</translate>
              </v-btn>
            </v-col>
          </v-row>
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
