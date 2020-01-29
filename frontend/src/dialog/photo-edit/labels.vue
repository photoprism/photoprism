<template>
    <div class="p-tab p-tab-photo-edit-labels">
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
                <td>{{ props.item.Label.LabelName }}</td>
                <td class="text-xs-left">{{ props.item.LabelSource }}</td>
                <td class="text-xs-center">{{ 100 - props.item.LabelUncertainty }}%</td>
                <td class="text-xs-center"><v-btn icon small flat :ripple="false"
                                                  class="p-photo-label-remove"
                                                  @click.stop.prevent="removeLabel(props.item.Label)">
                    <v-icon color="secondary-dark">delete</v-icon>
                </v-btn></td>
            </template>
            <template v-slot:footer>
                <td>
                    <v-text-field
                            v-model="newLabel"
                            :rules="[nameRule]"
                            color="secondary-dark"
                            :label="labels.addLabel"
                            single-line
                            flat solo hide-details
                            autofocus
                            @keyup.enter.native="addLabel"
                    ></v-text-field>
                </td>
                <td class="text-xs-left">manual</td>
                <td class="text-xs-center">100%</td>
                <td class="text-xs-center"><v-btn icon small flat :ripple="false"
                                                  class="p-photo-label-remove"
                                                  @click.stop.prevent="addLabel">
                    <v-icon color="secondary-dark">add</v-icon>
                </v-btn></td>
            </template>
        </v-data-table>
    </div>
</template>

<script>
    import Label from "model/label";

    export default {
        name: 'p-tab-photo-edit-labels',
        props: {
            model: Object,
        },
        data() {
            return {
                config: this.$config.values,
                readonly: this.$config.getValue("readonly"),
                selected: [],
                newLabel: "",
                listColumns: [
                    {text: this.$gettext('Label'), value: '', sortable: false, align: 'left'},
                    {text: this.$gettext('Source'), value: 'LabelSource', sortable: false, align: 'left'},
                    {text: this.$gettext('Confidence'), value: 'LabelUncertainty', sortable: false, align: 'center'},
                    {text: this.$gettext('Action'), value: '', sortable: false, align: 'center'},
                ],
                labels: {
                    addLabel: "",
                },
                nameRule: v => v.length <= 25 || this.$gettext("Name too long"),
            };
        },
        computed: {
        },
        methods: {
            refresh() {
            },
            removeLabel(label) {
                if(!label) {
                    return
                }

                const name = label.LabelName;

                this.model.removeLabel(label.ID).then((m) => {
                    this.$notify.success("removed " + name);
                });
            },
            addLabel() {
                if(!this.newLabel) {
                    return
                }

                this.model.addLabel(this.newLabel).then((m) => {
                    this.$notify.success("added " + this.newLabel);

                    this.newLabel = "";
                });
            },
        },
    };
</script>
