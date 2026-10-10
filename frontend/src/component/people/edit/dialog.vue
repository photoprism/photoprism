<template>
  <v-dialog
    ref="dialog"
    :model-value="visible"
    persistent
    max-width="500"
    class="dialog-person-edit"
    color="background"
    @keydown.esc.exact="close"
    @keydown.enter.capture="onEnterDown"
    @keyup.enter="onEnterUp"
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
          <v-row class="align-center" density="compact">
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
            <v-col cols="12">
              <v-date-input
                v-model:menu="birthdayMenu"
                :model-value="model.getBirthday()"
                :label="$gettext('Birth Date')"
                :min="minBirthday"
                :max="today"
                :disabled="disabled"
                view-mode="year"
                clearable
                class="input-birthday"
                @update:model-value="model.setBirthday($event)"
              ></v-date-input>
            </v-col>
            <v-col cols="12" sm="6">
              <v-checkbox v-model="model.Favorite" :disabled="disabled" :label="$gettext('Favorite')" density="comfortable" hide-details> </v-checkbox>
            </v-col>
            <v-col cols="12" sm="6">
              <v-checkbox
                v-model="model.Verified"
                :disabled="disabled"
                :label="$gettext('Verified')"
                density="comfortable"
                hide-details
              >
              </v-checkbox>
            </v-col>
            <v-col cols="12" sm="6">
              <v-checkbox v-model="model.Hidden" :disabled="disabled" :label="$gettext('Hidden')" density="comfortable" hide-details> </v-checkbox>
            </v-col>
            <v-col cols="12" sm="6">
              <v-checkbox v-model="model.Private" :disabled="disabled" :label="$gettext('Private')" density="comfortable" hide-details> </v-checkbox>
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
import Subject, { BirthYearMin, MaxLength as SubjectMaxLength } from "model/subject";
import { rules } from "common/form";

// isModifiedKey reports whether a modifier was held, or an input method was composing text.
const isModifiedKey = (ev) => !!ev && (ev.shiftKey || ev.ctrlKey || ev.altKey || ev.metaKey || ev.isComposing);

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
      // Keep the picker inside the range the API accepts, so an implausible year is never offered.
      today: new Date(),
      minBirthday: new Date(BirthYearMin, 0, 1),
      model: new Subject(),
      birthdayMenu: false,
      enterConfirms: false,
      rules,
      SubjectMaxLength,
    };
  },
  watch: {
    visible: function (show) {
      if (show) {
        this.model = this.person.clone();
        // Re-read on open rather than once at mount, since the dialog stays mounted between edits.
        this.today = new Date();
        this.enterConfirms = false;
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
    // onEnterDown records whether Enter went down in the dialog while the date picker was closed.
    // It runs in the capture phase, before the date input opens its picker on the same key.
    onEnterDown(ev) {
      this.enterConfirms = !this.birthdayMenu && !isModifiedKey(ev);
    },
    // onEnterUp confirms the dialog unless Enter belonged to the date picker, which it closes instead.
    // A key pressed in the picker's menu goes down outside the dialog, so only its keyup can arrive.
    onEnterUp(ev) {
      const confirms = this.enterConfirms && !isModifiedKey(ev);
      this.enterConfirms = false;

      if (confirms) {
        this.confirm();
      } else if (this.birthdayMenu) {
        this.birthdayMenu = false;
      }
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
