<template>
  <div class="p-tab p-tab-photo-labels">
    <v-form
      ref="form"
      class="p-form p-form--table p-form-photo-labels"
      validate-on="invalid-input"
      accept-charset="UTF-8"
      tabindex="1"
      @submit.prevent
    >
      <div class="form-body">
        <div class="form-controls">
          <v-row dense align="start">
            <v-col cols="0" sm="2" class="form-thumb">
              <div>
                <img
                  :alt="view?.model.Title"
                  :src="view?.model.thumbnailUrl('tile_500')"
                  class="clickable"
                  @click.stop.prevent.exact="openPhoto()"
                />
              </div>
            </v-col>
            <v-col cols="12" sm="10" class="d-flex flex-column ga-4">
              <div
                :class="$vuetify.display.smAndDown ? 'v-table--density-compact' : 'v-table--density-comfortable'"
                class="v-table v-table--has-top v-table--hover v-data-table elevation-0 edit-table list-view"
              >
                <div class="v-table__wrapper">
                  <table>
                    <thead>
                      <tr>
                        <th
                          class="v-data-table__td v-data-table-column--align-left v-data-table__th"
                          colspan="1"
                          rowspan="1"
                        >
                          <div class="v-data-table-header__content">
                            <span>{{ $gettext(`Label`) }}</span>
                          </div>
                        </th>
                        <th
                          class="v-data-table__td v-data-table-column--align-left v-data-table__th"
                          colspan="1"
                          rowspan="1"
                        >
                          <div class="v-data-table-header__content">
                            <span>{{ $gettext(`Source`) }}</span>
                          </div>
                        </th>
                        <th
                          class="v-data-table__td v-data-table-column--align-center v-data-table__th"
                          colspan="1"
                          rowspan="1"
                        >
                          <div class="v-data-table-header__content">
                            <span>{{ $gettext(`Confidence`) }}</span>
                          </div>
                        </th>
                        <th
                          class="v-data-table__td v-data-table-column--align-center v-data-table__th"
                          colspan="1"
                          rowspan="1"
                        >
                          <div class="v-data-table-header__content">
                            <span>{{ $gettext(`Action`) }}</span>
                          </div>
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="label in view.model.Labels" :key="label.LabelID" class="label result">
                        <td class="text-start">
                          {{ label.Label.Name }}
                          <!--                  TODO: add this dialog later-->
                          <!--                  <v-dialog class="p-inline-edit" @save="renameLabel(props.item.Label)">-->
                          <!--                    {{ props.item.Label.Name }}-->
                          <!--                    <template #input>-->
                          <!--                      <v-text-field v-model="props.item.Label.Name" :rules="[nameRule]" :label="$gettext('Name')" color="surface-variant" class="input-rename background-inherit elevation-0" single-line autofocus variant="solo" hide-details></v-text-field>-->
                          <!--                    </template>-->
                          <!--                  </v-dialog>-->
                        </td>
                        <td class="text-start">
                          {{ sourceName(label.LabelSrc) }}
                        </td>
                        <td class="text-center">{{ 100 - label.Uncertainty }}%</td>
                        <td class="text-center">
                          <v-btn
                            v-if="disabled"
                            icon
                            density="comfortable"
                            variant="text"
                            :ripple="false"
                            class="action-view"
                            title="Search"
                            @click.stop.prevent="searchLabel(label.Label)"
                          >
                            <v-icon color="surface-variant">mdi-magnify</v-icon>
                          </v-btn>
                          <v-btn
                            v-else-if="label.Uncertainty < 100 && label.LabelSrc === 'manual'"
                            icon
                            density="comfortable"
                            variant="text"
                            :ripple="false"
                            class="action-delete"
                            title="Delete"
                            @click.stop.prevent="removeLabel(label.Label)"
                          >
                            <v-icon color="surface-variant">mdi-delete</v-icon>
                          </v-btn>
                          <v-btn
                            v-else-if="label.Uncertainty < 100"
                            icon
                            density="comfortable"
                            variant="text"
                            :ripple="false"
                            class="action-remove"
                            title="Remove"
                            @click.stop.prevent="removeLabel(label.Label)"
                          >
                            <v-icon color="surface-variant">mdi-minus</v-icon>
                          </v-btn>
                          <v-btn
                            v-else
                            icon
                            density="comfortable"
                            variant="text"
                            :ripple="false"
                            class="action-on"
                            title="Activate"
                            @click.stop.prevent="activateLabel(label.Label)"
                          >
                            <v-icon color="surface-variant">mdi-plus</v-icon>
                          </v-btn>
                        </td>
                      </tr>
                      <tr v-if="!disabled" class="label result">
                        <td class="text-start">
                          <v-text-field
                            v-model="newLabel"
                            :rules="[nameRule]"
                            color="surface-variant"
                            autocomplete="off"
                            autofocus
                            single-line
                            flat
                            variant="plain"
                            hide-details
                            class="input-label ma-0 pa-0"
                            @keyup.enter="addLabel"
                          ></v-text-field>
                        </td>
                        <td class="text-start">
                          {{ sourceName("manual") }}
                        </td>
                        <td class="text-center">100%</td>
                        <td class="text-center">
                          <v-btn
                            icon
                            density="comfortable"
                            variant="text"
                            :ripple="false"
                            title="Add"
                            class="p-photo-label-add"
                            @click.stop.prevent="addLabel"
                          >
                            <v-icon color="surface-variant">mdi-plus</v-icon>
                          </v-btn>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </v-col>
          </v-row>
        </div>
      </div>
    </v-form>
  </div>
</template>

<script>
import Thumb from "model/thumb";

export default {
  name: "PTabPhotoLabels",
  props: {
    uid: {
      type: String,
      default: "",
    },
  },
  data() {
    return {
      view: this.$view.data(),
      disabled: !this.$config.feature("edit"),
      config: this.$config.values,
      readonly: this.$config.get("readonly"),
      selected: [],
      newLabel: "",
      listColumns: [
        { title: this.$gettext("Label"), key: "", sortable: false, align: "left" },
        { title: this.$gettext("Source"), key: "LabelSrc", sortable: false, align: "left" },
        {
          title: this.$gettext("Confidence"),
          key: "Uncertainty",
          sortable: false,
          align: "center",
        },
        { title: this.$gettext("Action"), key: "", sortable: false, align: "center" },
      ],
      nameRule: (v) => v.length <= this.$config.get("clip") || this.$gettext("Name too long"),
    };
  },
  methods: {
    refresh() {},
    sourceName(s) {
      switch (s) {
        case "manual":
          return this.$gettext("Manual");
        case "title":
          return this.$gettext("Title");
        case "caption":
          return this.$gettext("Caption");
        case "subject":
          return this.$gettext("Subject");
        case "image":
          return this.$gettext("Image");
        case "location":
          return this.$gettext("Location");
        default:
          return this.$util.ucFirst(s);
      }
    },
    removeLabel(label) {
      if (!label || !this.view?.model) {
        return;
      }

      const name = label.Name;

      this.view.model.removeLabel(label.ID).then((m) => {
        this.$notify.success("removed " + name);
      });
    },
    addLabel() {
      if (!this.newLabel || !this.view?.model) {
        return;
      }

      this.view.model.addLabel(this.newLabel).then((m) => {
        this.$notify.success("added " + this.newLabel);

        this.newLabel = "";
      });
    },
    activateLabel(label) {
      if (!label || !this.view?.model) {
        return;
      }

      this.view.model.activateLabel(label.ID);
    },
    // TODO: add this dialog later
    // renameLabel(label) {
    //   if (!label) {
    //     return;
    //   }
    //
    //   this.view.model.renameLabel(label.ID, label.Name);
    // },
    searchLabel(label) {
      this.$router.push({ name: "all", query: { q: "label:" + label.Slug } }).catch(() => {});
      this.$emit("close");
    },
    openPhoto() {
      if (!this.view?.model) {
        return;
      }

      this.$lightbox.openModels(Thumb.fromFiles([this.view.model]), 0);
    },
  },
};
</script>
