<template>
  <v-card flat tile class="mt-0 px-1 bg-background">
    <v-card-title class="pb-0 text-subtitle-2 d-flex align-center justify-space-between">
      <span>{{ title }}</span>
      <v-btn
        type="button"
        size="small"
        variant="text"
        color="surface-variant"
        prepend-icon="mdi-restore"
        class="action-metadata-reset"
        @click.stop="resetLayout"
      >
        {{ $gettext("Default") }}
      </v-btn>
    </v-card-title>

    <v-card-subtitle class="pt-1 pb-0 text-caption">
      {{ hint }}
    </v-card-subtitle>

    <v-card-actions class="pt-3">
      <div class="metadata-layout-editor w-100">
        <div v-if="rows.length === 0" class="metadata-layout-editor__empty text-body-2 text-medium-emphasis">
          {{ $gettext("No data available") }}
        </div>

        <div v-for="(fieldId, index) in rows" :key="`${view}-${index}`" class="metadata-layout-editor__row">
          <v-select
            :model-value="fieldId"
            :items="options"
            item-title="label"
            item-value="id"
            density="comfortable"
            variant="solo-filled"
            hide-details
            class="metadata-layout-editor__select"
            @update:model-value="updateField(index, $event)"
          >
            <template #item="{ props, item }">
              <v-list-item v-bind="props" :prepend-icon="item.raw.icon" :title="item.raw.label"></v-list-item>
            </template>
            <template #selection="{ item }">
              <div class="d-flex align-center">
                <v-icon :icon="item.raw.icon" size="18" class="me-2"></v-icon>
                <span>{{ item.raw.label }}</span>
              </div>
            </template>
          </v-select>

          <div class="metadata-layout-editor__actions">
            <v-btn
              type="button"
              icon="mdi-chevron-up"
              size="small"
              variant="text"
              :disabled="index === 0"
              :title="$gettext('Previous')"
              @click.stop="moveField(index, -1)"
            ></v-btn>
            <v-btn
              type="button"
              icon="mdi-chevron-down"
              size="small"
              variant="text"
              :disabled="index >= rows.length - 1"
              :title="$gettext('Next')"
              @click.stop="moveField(index, 1)"
            ></v-btn>
            <v-btn type="button" icon="mdi-delete-outline" size="small" variant="text" :title="$gettext('Remove')" @click.stop="removeField(index)"></v-btn>
          </div>
        </div>

        <div class="metadata-layout-editor__footer">
          <v-btn type="button" color="highlight" variant="flat" prepend-icon="mdi-plus" class="action-metadata-add" @click.stop="addField">
            {{ $gettext("Add") }}
          </v-btn>
        </div>
      </div>
    </v-card-actions>
  </v-card>
</template>

<script>
import { defaultMetadataLayout, metadataFieldOptions } from "common/metadata";

export default {
  name: "PSettingsMetadataLayout",
  props: {
    modelValue: {
      type: Array,
      default: () => [],
    },
    view: {
      type: String,
      default: "",
    },
    title: {
      type: String,
      default: "",
    },
    hint: {
      type: String,
      default: "",
    },
  },
  emits: ["update:modelValue"],
  computed: {
    options() {
      return metadataFieldOptions(this.view);
    },
    rows() {
      return Array.isArray(this.modelValue) ? this.modelValue : [];
    },
  },
  methods: {
    emitLayout(layout) {
      this.$emit("update:modelValue", layout);
    },
    addField() {
      const fallback = this.options[0]?.id;

      if (!fallback) {
        return;
      }

      this.emitLayout(this.rows.concat(fallback));
    },
    updateField(index, fieldId) {
      const layout = this.rows.slice();
      layout[index] = fieldId;
      this.emitLayout(layout);
    },
    moveField(index, delta) {
      const target = index + delta;

      if (target < 0 || target >= this.rows.length) {
        return;
      }

      const layout = this.rows.slice();
      const [field] = layout.splice(index, 1);
      layout.splice(target, 0, field);
      this.emitLayout(layout);
    },
    removeField(index) {
      const layout = this.rows.slice();
      layout.splice(index, 1);
      this.emitLayout(layout);
    },
    resetLayout() {
      this.emitLayout(defaultMetadataLayout(this.view));
    },
  },
};
</script>
