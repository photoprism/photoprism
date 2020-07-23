<template>
  <div class="p-tab p-tab-photo-labels">
    <v-data-table
            :headers="listColumns"
            :items="model.Labels"
            hide-actions
            class="elevation-0 p-files p-files-list p-results"
            disable-initial-sort
            item-key="ID"
            v-model="selected"
            :no-data-text="$gettext('No labels found')"
    >
      <template v-slot:items="props" class="p-file">
        <td>
          <v-edit-dialog
                  :return-value.sync="props.item.Label.Name"
                  lazy
                  @save="renameLabel(props.item.Label)"
                  class="p-inline-edit"
          >
            {{ props.item.Label.Name | capitalize }}
            <template v-slot:input>
              <v-text-field
                      v-model="props.item.Label.Name"
                      :rules="[nameRule]"
                      :label="labels.name"
                      color="secondary-dark"
                      single-line
                      autofocus
                      class="input-rename"
              ></v-text-field>
            </template>
          </v-edit-dialog>
        </td>
        <td class="text-xs-left">{{ sourceName(props.item.LabelSrc) }}</td>
        <td class="text-xs-center">{{ 100 - props.item.Uncertainty }}%</td>
        <td class="text-xs-center">
          <v-btn v-if="disabled" icon small flat :ripple="false"
                 class="action-view" title="Search"
                 @click.stop.prevent="searchLabel(props.item.Label)">
            <v-icon color="secondary-dark">search</v-icon>
          </v-btn>
          <v-btn v-else-if="props.item.Uncertainty < 100 && props.item.LabelSrc === 'manual'" icon
                 small flat :ripple="false"
                 class="action-delete" title="Delete"
                 @click.stop.prevent="removeLabel(props.item.Label)">
            <v-icon color="secondary-dark">delete</v-icon>
          </v-btn>
          <v-btn v-else-if="props.item.Uncertainty < 100" icon small flat :ripple="false"
                 class="action-remove" title="Remove"
                 @click.stop.prevent="removeLabel(props.item.Label)">
            <v-icon color="secondary-dark">remove</v-icon>
          </v-btn>
          <v-btn v-else icon small flat :ripple="false"
                 class="action-on" title="Activate"
                 @click.stop.prevent="activateLabel(props.item.Label)">
            <v-icon color="secondary-dark">add</v-icon>
          </v-btn>
        </td>
      </template>
      <template v-slot:footer v-if="!disabled">
        <td>
          <v-text-field
                  v-model="newLabel"
                  :rules="[nameRule]"
                  color="secondary-dark"
                  browser-autocomplete="off"
                  :label="labels.addLabel"
                  single-line
                  flat solo hide-details
                  autofocus
                  @keyup.enter.native="addLabel"
                  class="input-label"
          ></v-text-field>
        </td>
        <td class="text-xs-left">{{ sourceName('manual') }}</td>
        <td class="text-xs-center">100%</td>
        <td class="text-xs-center">
          <v-btn icon small flat :ripple="false" title="Add"
                 class="p-photo-label-add"
                 @click.stop.prevent="addLabel">
            <v-icon color="secondary-dark">add</v-icon>
          </v-btn>
        </td>
      </template>
    </v-data-table>
  </div>
</template>

<script>
    import Label from "model/label";

    export default {
        name: 'p-tab-photo-labels',
        props: {
            model: Object,
            uid: String,
        },
        data() {
            return {
                disabled: !this.$config.feature("edit"),
                config: this.$config.values,
                readonly: this.$config.get("readonly"),
                selected: [],
                newLabel: "",
                listColumns: [
                    {text: this.$gettext('Label'), value: '', sortable: false, align: 'left'},
                    {text: this.$gettext('Source'), value: 'LabelSrc', sortable: false, align: 'left'},
                    {text: this.$gettext('Confidence'), value: 'Uncertainty', sortable: false, align: 'center'},
                    {text: this.$gettext('Action'), value: '', sortable: false, align: 'center'},
                ],
                labels: {
                    addLabel: "",
                    search: this.$gettext("Search"),
                    name: this.$gettext("Label Name"),
                },
                nameRule: v => v.length <= this.$config.get('clip') || this.$gettext("Name too long"),
            };
        },
        computed: {},
        methods: {
            refresh() {
            },
            sourceName(s) {
                switch (s) {
                    case "manual": return this.$gettext("manual");
                    case "image": return this.$gettext("image");
                    case "location": return this.$gettext("location");
                    default: return s;
                }
            },
            removeLabel(label) {
                if (!label) {
                    return
                }

                const name = label.Name;

                this.model.removeLabel(label.ID).then((m) => {
                    this.$notify.success("removed " + name);
                });
            },
            addLabel() {
                if (!this.newLabel) {
                    return
                }

                this.model.addLabel(this.newLabel).then((m) => {
                    this.$notify.success("added " + this.newLabel);

                    this.newLabel = "";
                });
            },
            activateLabel(label) {
                if (!label) {
                    return
                }

                this.model.activateLabel(label.ID);
            },
            renameLabel(label) {
                if (!label) {
                    return
                }

                this.model.renameLabel(label.ID, label.Name);
            },
            searchLabel(label) {
                this.$router.push({name: 'photos', query: {q: 'label:' + label.Slug}}).catch(err => {
                });
                this.$emit('close');
            },
        },
    };
</script>
