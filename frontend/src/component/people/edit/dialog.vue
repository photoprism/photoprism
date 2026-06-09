<template>
  <v-dialog
    ref="dialog"
    :model-value="visible"
    persistent
    max-width="500"
    class="dialog-person-edit"
    color="background"
    @keydown.esc.exact="close"
    @keyup.enter.exact="confirm"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-form ref="form" validate-on="invalid-input" class="form-person-edit" accept-charset="UTF-8" tabindex="-1" @submit.prevent="confirm">
      <v-card>
        <v-toolbar flat color="navigation" class="mb-4" density="comfortable">
          <v-toolbar-title>
            {{ $gettext(`Edit %{s}`, { s: model.modelName() }) }}
          </v-toolbar-title>
          <v-btn icon class="action-close" :aria-label="$gettext('Close')" @click.stop="close">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-toolbar>
        <v-card-text class="dense">
          <v-row align="center" dense>
            <v-col cols="12">
              <v-text-field
                v-model="model.Name"
                autofocus
                :rules="rules.text(false, 0, SubjectMaxLength.Name, $gettext('Name'))"
                :label="$gettext('Name')"
                :disabled="disabled"
                class="input-title"
              ></v-text-field>
            </v-col>
            <v-col sm="4">
              <v-checkbox v-model="model.Favorite" :disabled="disabled" :label="$gettext('Favorite')" density="comfortable" hide-details> </v-checkbox>
            </v-col>
            <v-col sm="4">
              <v-checkbox v-model="model.Hidden" :disabled="disabled" :label="$gettext('Hidden')" density="comfortable" hide-details> </v-checkbox>
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
import Subject, { MaxLength as SubjectMaxLength } from "model/subject";
import { rules } from "common/form";

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
  emits: ["close", "confirm"],
  data() {
    return {
      disabled: !this.$config.allow("people", "manage"),
      model: new Subject(),
      rules,
      SubjectMaxLength,
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
      // Seed validation so pre-filled overlong input surfaces the inline error on first render.
      this.$refs.form?.validate?.();
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

      // Form-level gate: :rules alone only renders the inline error.
      const form = this.$refs.form;
      const validate = typeof form?.validate === "function" ? form.validate() : Promise.resolve({ valid: true });

      return Promise.resolve(validate).then((result) => {
        if (result && result.valid === false) {
          this.$notify.error(this.$gettext("Changes could not be saved"));
          return;
        }

        // Trim runs in Subject.update() at the model boundary (parent emits-then-saves).
        this.$emit("confirm", this.model);
      });
    },
  },
};
</script>
