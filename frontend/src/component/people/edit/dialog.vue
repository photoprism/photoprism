<template>
  <v-dialog
    :model-value="visible"
    persistent
    max-width="500"
    class="dialog-person-edit"
    color="background"
    @keydown.esc="close"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-form
      ref="form"
      validate-on="invalid-input"
      class="form-person-edit"
      accept-charset="UTF-8"
      tabindex="1"
      @submit.prevent="confirm"
    >
      <v-card>
        <v-card-title class="d-flex justify-start align-center ga-3">
          <v-icon size="28" color="primary">mdi-account</v-icon>
          <h6 class="text-h6">{{ $gettext(`Edit %{s}`, { s: model.modelName() }) }}</h6>
        </v-card-title>
        <v-card-text class="dense">
          <v-row align="center" dense>
            <v-col cols="12">
              <v-text-field
                v-model="model.Name"
                hide-details
                autofocus
                :rules="[titleRule]"
                :label="$gettext('Name')"
                :disabled="disabled"
                class="input-title"
                @keyup.enter="confirm"
              ></v-text-field>
            </v-col>
            <v-col sm="4">
              <v-checkbox
                v-model="model.Favorite"
                :disabled="disabled"
                :label="$gettext('Favorite')"
                density="comfortable"
                hide-details
              >
              </v-checkbox>
            </v-col>
            <v-col sm="4">
              <v-checkbox
                v-model="model.Hidden"
                :disabled="disabled"
                :label="$gettext('Hidden')"
                density="comfortable"
                hide-details
              >
              </v-checkbox>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions class="action-buttons">
          <v-btn variant="flat" color="button" class="action-cancel" @click.stop="close">
            {{ $gettext(`Cancel`) }}
          </v-btn>
          <v-btn variant="flat" color="highlight" class="action-confirm" :disabled="disabled" @click.stop="confirm">
            {{ $gettext(`Save`) }}
          </v-btn>
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
    visible: {
      type: Boolean,
      default: false,
    },
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
    visible: function (show) {
      if (show) {
        this.model = this.person.clone();
      }
    },
  },
  methods: {
    afterEnter() {
      this.$view.enter(this);
    },
    afterLeave() {
      this.$view.leave(this);
    },
    close() {
      this.$emit("close");
    },
    confirm() {
      if (this.disabled) {
        this.close();
        return;
      }

      this.$emit("confirm", this.model);
    },
  },
};
</script>
