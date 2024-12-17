<template>
  <v-dialog :model-value="show" persistent max-width="500" class="dialog-label-edit" color="background" @keydown.esc="close">
    <v-form ref="form" validate-on="blur" class="form-label-edit" accept-charset="UTF-8" @submit.prevent="confirm">
      <v-card>
        <v-card-title class="d-flex justify-start align-center ga-3">
          <v-icon size="28" color="primary">mdi-label</v-icon>
          <h6 class="text-h6"><translate :translate-params="{ name: model.modelName() }">Edit %{name}</translate></h6>
        </v-card-title>

        <v-card-text class="dense">
            <v-row dense>
              <v-col cols="12">
                <v-text-field v-model="model.Name" hide-details autofocus :rules="[titleRule]" :label="$gettext('Name')" :disabled="disabled" class="input-title" @keyup.enter="confirm"></v-text-field>
              </v-col>
              <v-col sm="4">
                <!-- TODO: check property flat TEST -->
<!--                TODO: fix Favorite saving-->
                <v-checkbox v-model="model.Favorite" :disabled="disabled" :label="$gettext('Favorite')" hide-details> </v-checkbox>
              </v-col>
            </v-row>
        </v-card-text>
        <v-card-actions>
          <v-btn variant="flat" color="button" class="action-cancel" @click.stop="close">
            <translate>Cancel</translate>
          </v-btn>
          <v-btn variant="flat" color="highlight" class="action-confirm" :disabled="disabled" @click.stop="confirm">
            <translate>Save</translate>
          </v-btn>
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
