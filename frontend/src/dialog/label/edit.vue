<template>
  <v-dialog :model-value="show" persistent max-width="500" class="dialog-label-edit" color="background" @keydown.esc="close">
    <v-form ref="form" lazy-validation class="form-label-edit" accept-charset="UTF-8" @submit.prevent="confirm">
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
                <!-- TODO: check property flat TEST -->
<!--                TODO: fix Favorite saving-->
                <v-checkbox v-model="model.Favorite" :disabled="disabled" color="surface-variant" :label="$gettext('Favorite')" hide-details flat> </v-checkbox>
              </v-col>
            </v-row>
          </v-container>
        </v-card-text>
        <v-card-actions class="pt-0 px-6">
          <v-row class="pa-2">
            <v-col cols="12" class="text-right">
              <v-btn variant="flat" color="button" class="action-cancel" @click.stop="close">
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
import Label from "model/label";

export default {
  name: "PLabelEditDialog",
  props: {
    show: Boolean,
    label: {
      type: Object,
      default: () => {},
    },
  },
  data() {
    return {
      disabled: !this.$config.allow("labels", "manage"),
      model: new Label(),
      titleRule: (v) => v.length <= this.$config.get("clip") || this.$gettext("Name too long"),
    };
  },
  watch: {
    show: function (show) {
      if (show) {
        this.model = this.label.clone();
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

      this.model.update().then((m) => {
        this.$notify.success(this.$gettext("Changes successfully saved"));
        this.$emit("close");
      });
    },
  },
};
</script>
