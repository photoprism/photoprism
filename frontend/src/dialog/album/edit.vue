<template>
  <v-dialog :value="show" lazy persistent max-width="500" class="dialog-album-edit" color="application" @keydown.esc="close">
    <v-form ref="form" lazy-validation dense class="form-album-edit" accept-charset="UTF-8" @submit.prevent="confirm">
      <v-card raised elevation="24">
        <v-card-title primary-title class="pb-0">
          <v-layout row wrap>
            <v-flex xs12>
              <h3 class="headline mx-2 mb-0">
                <translate :translate-params="{ name: model.modelName() }">Edit %{name}</translate>
              </h3>
            </v-flex>
          </v-layout>
        </v-card-title>

        <v-card-text>
          <v-container fluid class="pa-0">
            <v-layout row wrap>
              <v-flex v-if="album.Type !== 'month'" xs12 pa-2>
                <v-text-field v-model="model.Title" hide-details autofocus box flat :rules="[titleRule]" :label="$gettext('Name')" :disabled="disabled" color="secondary-dark" class="input-title" @keyup.enter.native="confirm"></v-text-field>
              </v-flex>
              <v-flex xs12 pa-2>
                <v-text-field v-model="model.Location" hide-details box flat :label="$gettext('Location')" :disabled="disabled" color="secondary-dark" class="input-location"></v-text-field>
              </v-flex>
              <v-flex xs12 pa-2>
                <v-textarea :key="growDesc" v-model="model.Description" auto-grow hide-details box flat browser-autocomplete="off" :label="$gettext('Description')" :rows="1" :disabled="disabled" class="input-description" color="secondary-dark"></v-textarea>
              </v-flex>
              <v-flex xs12 pa-2>
                <v-combobox v-model="model.Category" hide-details box flat :search-input.sync="model.Category" :items="categories" :disabled="disabled" :label="$gettext('Category')" :allow-overflow="false" return-masked-value color="secondary-dark" class="input-category"></v-combobox>
              </v-flex>
              <v-flex xs12 sm6 pa-2>
                <v-select v-model="model.Order" :label="$gettext('Sort Order')" :menu-props="{ maxHeight: 400 }" hide-details box flat :items="sorting" :disabled="disabled" item-value="value" item-text="text" color="secondary-dark"></v-select>
              </v-flex>
              <v-flex sm3 pa-2>
                <v-checkbox v-model="model.Favorite" :disabled="disabled" color="secondary-dark" :label="$gettext('Favorite')" hide-details flat> </v-checkbox>
              </v-flex>
              <v-flex v-if="experimental && featPrivate" sm3 pa-2>
                <v-checkbox v-model="model.Private" :disabled="disabled" color="secondary-dark" :label="$gettext('Private')" hide-details flat> </v-checkbox>
              </v-flex>
            </v-layout>
          </v-container>
        </v-card-text>
        <v-card-actions class="pt-0 px-3">
          <v-layout row wrap class="pa-2">
            <v-flex xs12 text-xs-right>
              <v-btn depressed color="secondary-light" class="action-cancel" @click.stop="close">
                <translate>Cancel</translate>
              </v-btn>
              <v-btn depressed dark color="primary-button" class="action-confirm" :disabled="disabled" @click.stop="confirm">
                <translate>Save</translate>
              </v-btn>
            </v-flex>
          </v-layout>
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
        { value: "edited", text: this.$gettext("Recently Edited") },
        { value: "title", text: this.$gettext("Picture Title") },
        { value: "name", text: this.$gettext("File Name") },
        { value: "size", text: this.$gettext("File Size") },
        { value: "duration", text: this.$gettext("Video Duration") },
        { value: "similar", text: this.$gettext("Visual Similarity") },
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
