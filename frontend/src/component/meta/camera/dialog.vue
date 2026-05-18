<template>
  <v-dialog
    ref="dialog"
    :model-value="visible"
    :max-width="480"
    :fullscreen="$vuetify.display.xs"
    persistent
    scrim
    scrollable
    class="p-meta-camera-dialog"
    @keydown.esc.exact.stop="close"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-card ref="content" tabindex="-1" :tile="$vuetify.display.xs">
      <v-toolbar flat color="navigation" density="comfortable">
        <v-toolbar-title>
          {{ $gettext("Adjust Camera Info") }}
        </v-toolbar-title>
        <v-btn icon :aria-label="$gettext('Close')" @click.stop="close">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-toolbar>
      <v-card-text class="dense">
        <v-form ref="form" v-model="valid" validate-on="invalid-input" @submit.prevent="confirm">
        <v-row dense class="py-2">
          <v-col cols="12">
            <v-select
              v-model="cameraID"
              :placeholder="$gettext('Camera')"
              :menu-props="{ maxHeight: 346 }"
              autocomplete="off"
              hide-details
              item-value="ID"
              item-title="Name"
              :items="cameraOptions"
              prepend-inner-icon="mdi-camera"
              density="comfortable"
              class="input-camera"
            >
            </v-select>
          </v-col>
          <v-col cols="12">
            <v-select
              v-model="lensID"
              :placeholder="$gettext('Lens')"
              :menu-props="{ maxHeight: 346 }"
              autocomplete="off"
              hide-details
              item-value="ID"
              item-title="Name"
              :items="lensOptions"
              prepend-inner-icon="mdi-camera-iris"
              density="comfortable"
              class="input-lens"
            >
            </v-select>
          </v-col>
          <v-col cols="6">
            <v-text-field
              v-model="iso"
              autocomplete="off"
              autocorrect="off"
              autocapitalize="none"
              label="ISO"
              placeholder=""
              density="comfortable"
              validate-on="input"
              :rules="rules.number(false, 0, 128000)"
              class="input-iso"
            ></v-text-field>
          </v-col>
          <v-col cols="6">
            <v-text-field
              v-model="exposure"
              autocomplete="off"
              autocorrect="off"
              autocapitalize="none"
              :label="$gettext('Exposure')"
              placeholder=""
              density="comfortable"
              validate-on="input"
              :rules="rules.text(false, 0, PhotoMaxLength.Exposure, $gettext('Exposure'))"
              class="input-exposure"
            ></v-text-field>
          </v-col>
          <v-col cols="6">
            <v-text-field
              v-model="fNumber"
              autocomplete="off"
              autocorrect="off"
              autocapitalize="none"
              :label="$gettext('F Number')"
              placeholder=""
              density="comfortable"
              validate-on="input"
              :rules="rules.number(false, 0, 256)"
              class="input-fnumber"
            ></v-text-field>
          </v-col>
          <v-col cols="6">
            <v-text-field
              v-model="focalLength"
              autocomplete="off"
              :label="$gettext('Focal Length')"
              placeholder=""
              density="comfortable"
              validate-on="input"
              :rules="rules.number(false, 0, 128000)"
              class="input-focal-length"
            ></v-text-field>
          </v-col>
        </v-row>
        </v-form>
      </v-card-text>
      <v-card-actions class="action-buttons">
        <v-btn variant="flat" color="button" class="action-cancel" min-width="100" @click.stop="close">
          {{ $gettext("Cancel") }}
        </v-btn>
        <v-btn variant="flat" color="highlight" class="action-save action-confirm" min-width="100" :disabled="!valid" :aria-label="$gettext('Save changes')" @click="confirm">
          {{ $gettext("Save") }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
import { MaxLength as PhotoMaxLength } from "model/photo";
import { rules } from "common/form";

export default {
  name: "PMetaCameraDialog",
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    photo: {
      type: Object,
      default: null,
    },
  },
  emits: ["close", "confirm"],
  data() {
    return {
      rules,
      PhotoMaxLength,
      // Reflects the v-form's aggregate validity via v-model="valid".
      // Drives the Save button's `:disabled` state so the user gets a
      // visual cue before clicking — mirrors the datetime dialog's
      // `:disabled="invalidDate"` pattern. Starts true so a freshly
      // opened dialog on valid data doesn't flash disabled; afterEnter()
      // seeds validate() to set the real state on mount.
      valid: true,
      cameraID: 0,
      lensID: 0,
      iso: "",
      exposure: "",
      fNumber: "",
      focalLength: "",
    };
  },
  computed: {
    cameraOptions() {
      return this.$config.values.cameras || [];
    },
    lensOptions() {
      return this.$config.values.lenses || [];
    },
  },
  watch: {
    visible(show) {
      if (show) {
        this.loadFromPhoto();
      }
    },
  },
  methods: {
    afterEnter() {
      this.$view.enter(this);
      // Seed validation so rules are active from the first render.
      // Mirrors the canonical pattern in page/settings/account.vue.
      this.$nextTick(() => this.$refs.form?.validate?.());
    },
    afterLeave() {
      this.$view.leave(this);
    },
    loadFromPhoto() {
      if (!this.photo) {
        return;
      }

      this.cameraID = this.photo.CameraID || 0;
      this.lensID = this.photo.LensID || 0;
      this.iso = this.photo.Iso || "";
      this.exposure = this.photo.Exposure || "";
      this.fNumber = this.photo.FNumber || "";
      this.focalLength = this.photo.FocalLength || "";
    },
    close() {
      this.$emit("close");
    },
    confirm() {
      // Gate the emit on form validation so an out-of-range ISO /
      // FNumber / FocalLength or an overlength Exposure cannot bypass
      // the inline rules and reach the parent's save handler. Falls
      // back to permissive when no v-form ref is mounted (test stub).
      const form = this.$refs.form;
      const validate = typeof form?.validate === "function" ? form.validate() : Promise.resolve({ valid: true });

      return Promise.resolve(validate).then((result) => {
        if (result && result.valid === false) {
          this.$notify.error(this.$gettext("Changes could not be saved"));
          return;
        }
        this.$emit("confirm", {
          CameraID: this.cameraID,
          LensID: this.lensID,
          Iso: this.iso,
          Exposure: this.exposure,
          FNumber: this.fNumber,
          FocalLength: this.focalLength,
        });
      });
    },
  },
};
</script>
