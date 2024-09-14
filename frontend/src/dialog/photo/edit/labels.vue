<template>
  <div class="p-tab p-tab-photo-labels">
    <v-form ref="form" lazy-validation dense accept-charset="UTF-8" @submit.prevent>
      <v-layout class="pa-2-md-and-up" row wrap align-top fill-height>
        <v-flex class="pa-2 hidden-sm-and-down" xs12 md2 xxl1 fill-height>
          <p-photo-preview :model="model"></p-photo-preview>
        </v-flex>
        <v-flex class="pa-2-md-and-up ra-4-table-md-and-up" xs12 md10 xxl11 fill-width fill-height>
          <v-data-table v-model="selected" :headers="listColumns" :items="model.Labels" hide-actions class="elevation-0 p-results" disable-initial-sort item-key="ID" :no-data-text="$gettext('No labels found')">
            <template #items="props" class="p-file">
              <td>
                <v-edit-dialog :return-value.sync="props.item.Label.Name" lazy class="p-inline-edit" @save="renameLabel(props.item.Label)">
                  {{ props.item.Label.Name }}
                  <template #input>
                    <v-text-field v-model="props.item.Label.Name" :rules="[nameRule]" :label="$gettext('Name')" color="secondary-dark" class="input-rename background-inherit elevation-0" single-line autofocus solo hide-details></v-text-field>
                  </template>
                </v-edit-dialog>
              </td>
              <td class="text-xs-left">
                {{ sourceName(props.item.LabelSrc) }}
              </td>
              <td class="text-xs-center"> {{ 100 - props.item.Uncertainty }}% </td>
              <td class="text-xs-center">
                <v-btn v-if="disabled" icon small flat :ripple="false" class="action-view" title="Search" @click.stop.prevent="searchLabel(props.item.Label)">
                  <v-icon color="secondary-dark">search</v-icon>
                </v-btn>
                <v-btn v-else-if="props.item.Uncertainty < 100 && props.item.LabelSrc === 'manual'" icon small flat :ripple="false" class="action-delete" title="Delete" @click.stop.prevent="removeLabel(props.item.Label)">
                  <v-icon color="secondary-dark">delete</v-icon>
                </v-btn>
                <v-btn v-else-if="props.item.Uncertainty < 100" icon small flat :ripple="false" class="action-remove" title="Remove" @click.stop.prevent="removeLabel(props.item.Label)">
                  <v-icon color="secondary-dark">remove</v-icon>
                </v-btn>
                <v-btn v-else icon small flat :ripple="false" class="action-on" title="Activate" @click.stop.prevent="activateLabel(props.item.Label)">
                  <v-icon color="secondary-dark">add</v-icon>
                </v-btn>
              </td>
            </template>
            <template v-if="!disabled" #footer>
              <td>
                <v-text-field v-model="newLabel" :rules="[nameRule]" color="secondary-dark" browser-autocomplete="off" :label="$gettext('Name')" single-line flat solo hide-details autofocus class="input-label" @keyup.enter.native="addLabel"></v-text-field>
              </td>
              <td class="text-xs-left">
                {{ sourceName("manual") }}
              </td>
              <td class="text-xs-center"> 100% </td>
              <td class="text-xs-center">
                <v-btn icon small flat :ripple="false" title="Add" class="p-photo-label-add" @click.stop.prevent="addLabel">
                  <v-icon color="secondary-dark">add</v-icon>
                </v-btn>
              </td>
            </template>
          </v-data-table>
        </v-flex>
      </v-layout>
      <!-- div class="mt-1 clear"></div -->
    </v-form>
  </div>
</template>

<script>
import Label from "model/label";
import Thumb from "model/thumb";

export default {
  name: "PTabPhotoLabels",
  props: {
    model: {
      type: Object,
      default: () => {},
    },
    uid: {
      type: String,
      default: "",
    },
  },
  data() {
    return {
      disabled: !this.$config.feature("edit"),
      config: this.$config.values,
      readonly: this.$config.get("readonly"),
      selected: [],
      newLabel: "",
      listColumns: [
        { text: this.$gettext("Label"), value: "", sortable: false, align: "left" },
        { text: this.$gettext("Source"), value: "LabelSrc", sortable: false, align: "left" },
        {
          text: this.$gettext("Confidence"),
          value: "Uncertainty",
          sortable: false,
          align: "center",
        },
        { text: this.$gettext("Action"), value: "", sortable: false, align: "center" },
      ],
      nameRule: (v) => v.length <= this.$config.get("clip") || this.$gettext("Name too long"),
    };
  },
  computed: {},
  methods: {
    refresh() {},
    sourceName(s) {
      switch (s) {
        case "manual":
          return this.$gettext("manual");
        case "image":
          return this.$gettext("image");
        case "location":
          return this.$gettext("location");
        default:
          return s;
      }
    },
    removeLabel(label) {
      if (!label) {
        return;
      }

      const name = label.Name;

      this.model.removeLabel(label.ID).then((m) => {
        this.$notify.success("removed " + name);
      });
    },
    addLabel() {
      if (!this.newLabel) {
        return;
      }

      this.model.addLabel(this.newLabel).then((m) => {
        this.$notify.success("added " + this.newLabel);

        this.newLabel = "";
      });
    },
    activateLabel(label) {
      if (!label) {
        return;
      }

      this.model.activateLabel(label.ID);
    },
    renameLabel(label) {
      if (!label) {
        return;
      }

      this.model.renameLabel(label.ID, label.Name);
    },
    searchLabel(label) {
      this.$router.push({ name: "all", query: { q: "label:" + label.Slug } }).catch(() => {});
      this.$emit("close");
    },
    openPhoto() {
      this.$viewer.show(Thumb.fromFiles([this.model]), 0);
    },
  },
};
</script>
