<template>
  <v-dialog lazy v-model="show" persistent max-width="500" class="dialog-album-edit" color="application"
            @keydown.esc="close">
    <v-form lazy-validation dense
            ref="form" class="form-album-edit" accept-charset="UTF-8"
            @submit.prevent="confirm">
      <v-card raised elevation="24">
        <v-card-title primary-title class="pb-0">
          <v-layout row wrap>
            <v-flex xs12>
              <h3 class="headline mx-2 mb-0">
                <translate :translate-params="{name: model.modelName()}">Edit %{name}</translate>
              </h3>
            </v-flex>
          </v-layout>
        </v-card-title>

        <v-card-text>
          <v-container fluid class="pa-0">
            <v-layout row wrap>
              <v-flex xs12 pa-2 v-if="album.Type !== 'month'">
                <v-text-field hide-details
                              v-model="model.Title"
                              :rules="[titleRule]"
                              :label="labels.title"
                              color="secondary-dark"
                              class="input-title"
                ></v-text-field>
              </v-flex>
              <v-flex xs12 pa-2>
                <v-text-field hide-details
                              v-model="model.Location"
                              :label="$gettext('Location')"
                              color="secondary-dark"
                              class="input-location"
                ></v-text-field>
              </v-flex>
              <v-flex xs12 pa-2>
                <v-textarea auto-grow hide-details
                            browser-autocomplete="off"
                            :label="labels.description"
                            :rows="1"
                            :key="growDesc"
                            v-model="model.Description"
                            class="input-description"
                            color="secondary-dark">
                </v-textarea>
              </v-flex>
              <v-flex xs12 md6 pa-2>
                <v-combobox hide-details :search-input.sync="model.Category"
                            v-model="model.Category"
                            :items="categories"
                            :label="labels.category"
                            :allow-overflow="false"
                            return-masked-value
                            color="secondary-dark"
                            class="input-category"
                >
                </v-combobox>
              </v-flex>
              <v-flex xs12 md6 pa-2>
                <v-select
                        :label="labels.sort"
                        hide-details
                        v-model="model.Order"
                        :items="sorting"
                        item-value="value"
                        item-text="text"
                        color="secondary-dark">
                </v-select>
              </v-flex>
            </v-layout>
          </v-container>
        </v-card-text>
        <v-card-actions class="pt-0">
          <v-layout row wrap class="pa-2">
            <v-flex xs12 text-xs-right>
              <v-btn @click.stop="close" depressed
                     color="secondary-light"
                     class="action-cancel">
                <translate>Cancel</translate>
              </v-btn>
              <v-btn @click.stop="confirm" depressed dark
                     color="secondary-dark"
                     class="action-confirm">
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
        name: 'p-album-edit-dialog',
        props: {
            show: Boolean,
            album: Object,
        },
        data() {
            return {
                model: new Album(),
                growDesc: false,
                loading: false,
                sorting: [
                    {value: 'added', text: this.$gettext('Recently added')},
                    {value: 'edited', text: this.$gettext('Recently edited')},
                    {value: 'newest', text: this.$gettext('Newest first')},
                    {value: 'oldest', text: this.$gettext('Oldest first')},
                    {value: 'name', text: this.$gettext('Sort by file name')},
                    {value: 'similar', text: this.$gettext('Group by similarity')},
                    {value: 'relevance', text: this.$gettext('Most relevant')},
                ],
                categories: this.$config.albumCategories(),
                titleRule: v => v.length <= this.$config.get('clip') || this.$gettext("Name too long"),
                labels: {
                    title: this.$gettext("Name"),
                    description: this.$gettext("Description"),
                    search: this.$gettext("Search"),
                    view: this.$gettext("View"),
                    country: this.$gettext("Country"),
                    camera: this.$gettext("Camera"),
                    sort: this.$gettext("Sort Order"),
                    category: this.$gettext("Category"),
                },
            }
        },
        methods: {
            expand() {
                this.growDesc = !this.growDesc;
            },
            close() {
                this.$emit('close');
            },
            confirm() {
                this.model.update().then((m) => {
                    this.categories = this.$config.albumCategories();
                    this.$emit('close');
                });
            },
        },
        watch: {
            show: function (show) {
                if (show) {
                    this.model = this.album.clone();
                }
            }
        },
    }
</script>
