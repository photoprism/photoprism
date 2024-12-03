<template>
  <v-dialog :model-value="show" persistent max-width="500" class="dialog-person-edit" color="background" @keydown.esc="close">
    <v-form ref="form" lazy-validation class="form-person-edit" accept-charset="UTF-8" @submit.prevent="confirm">
      <v-card elevation="24">
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
              <v-col cols="12" class="pa-2">
                <v-text-field v-model="model.Name" hide-details autofocus variant="solo" flat :rules="[titleRule]" :label="$gettext('Name')" :disabled="disabled" color="surface-variant" class="input-title" @keyup.enter="confirm"></v-text-field>
              </v-col>
              <v-col sm="4" class="pa-2">
                <v-checkbox v-model="model.Favorite" :disabled="disabled" color="surface-variant" :label="$gettext('Favorite')" hide-details flat> </v-checkbox>
              </v-col>
              <v-col sm="4" class="pa-2">
                <v-checkbox v-model="model.Hidden" :disabled="disabled" color="surface-variant" :label="$gettext('Hidden')" hide-details> </v-checkbox>
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
              <v-btn variant="flat" color="primary-button" class="action-confirm" :disabled="disabled" @click.stop="confirm">
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
import Subject from "model/subject";

export default {
  name: "PPeopleEditDialog",
  props: {
    show: Boolean,
    person: {
      type: Object,
      default: () => {},
    },
  },
  data() {
    return {
      disabled: !this.$config.allow("people", "manage"),
      model: new Subject(),
      titleRule: (v) => v.length <= this.$config.get("clip") || this.$gettext("Name too long"),
    };
  },
  watch: {
    show: function (show) {
      if (show) {
        this.model = this.person.clone();
      }
    },
  },
  methods: {
    close() {
      this.$emit("close");
    },
    confirm() {
      if (this.disabled) {
        this.close();
        return;
      }
      this.$emit("confirm");

      this.model.update().then((m) => {
        this.$notify.success(this.$gettext("Changes successfully saved"));
        this.$emit("close");
      });
    },
  },
};
</script>
