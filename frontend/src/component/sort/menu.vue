<template>
  <v-menu ref="menu" v-model="open" :close-on-content-click="false" location="bottom end" max-height="560">
    <template #activator="{ props }">
      <v-btn
        v-tooltip="label"
        v-bind="props"
        :aria-label="label"
        :title="label"
        :disabled="disabled"
        rounded="pill"
        class="action-sort ms-1"
        variant="tonal"
      >
        <v-icon icon="mdi-sort-variant"></v-icon>
        <v-icon v-if="active?.direction" :icon="ascending ? 'mdi-arrow-up' : 'mdi-arrow-down'" size="small"></v-icon>
      </v-btn>
    </template>
    <v-list class="p-sort-menu" role="group" :aria-label="$gettext('Sort By')" bg-color="navigation" nav>
      <v-list-subheader>{{ $gettext('Sort By') }}</v-list-subheader>
      <div v-for="item in options" :key="item.value" class="d-flex align-center" :class="{ 'bg-primary rounded': isActive(item) }">
        <v-list-item
          :active="isActive(item)"
          role="button"
          :aria-pressed="isActive(item)"
          :prepend-icon="item.icon"
          :title="item.text"
          :data-sort="item.value"
          class="flex-grow-1"
          @click="choose(item)"
        ></v-list-item>
        <template v-if="item.direction">
          <v-btn
            v-for="up in [true, false]"
            :key="String(up)"
            :icon="up ? 'mdi-arrow-up' : 'mdi-arrow-down'"
            :aria-label="directionLabel(item, up)"
            :title="directionLabel(item, up)"
            :aria-pressed="isActive(item) && ascending === up"
            :variant="isActive(item) && ascending === up ? 'tonal' : 'text'"
            :data-direction="up ? 'ascending' : 'descending'"
            :data-sort="item.value"
            size="small"
            @click.stop="choose(item, up)"
          ></v-btn>
        </template>
      </div>
    </v-list>
  </v-menu>
</template>
<script>
import { selectSort, sortAscending, sortLabel } from "common/sort";

export default {
  name: "PSortMenu",
  props: {
    options: { type: Array, required: true },
    modelValue: { type: Object, required: true },
    disabled: { type: Boolean, default: false },
  },
  emits: ["update:modelValue"],
  // data initializes the menu visibility.
  data() {
    return { open: false, restoreFocus: false };
  },
  computed: {
    // active returns the currently selected menu entry.
    active() {
      return this.options.find((item) => this.isActive(item));
    },
    // ascending reports the selected arrow direction.
    ascending() {
      return this.active ? sortAscending(this.modelValue, this.active) : false;
    },
    // label describes the current sort for tooltips and assistive technology.
    label() {
      return sortLabel(this.modelValue, this.active);
    },
  },
  watch: {
    // disabled restores keyboard focus after the selected sort finishes loading.
    disabled(value) {
      if (!value && this.restoreFocus) {
        this.$nextTick(this.focusButton);
      }
    },
  },
  methods: {
    // focusButton returns focus unless the user has moved to another control.
    focusButton(complete = true) {
      const active = document.activeElement;
      if (active === document.body || active?.classList.contains("p-page")) {
        this.$refs.menu?.activatorEl?.focus();
      }
      if (complete) {
        this.restoreFocus = false;
      }
    },
    // isActive matches a menu entry including the oldest date alias.
    isActive(item) {
      return this.modelValue.order === item.value || (item.value === "newest" && this.modelValue.order === "oldest");
    },
    // directionLabel describes the result of selecting a direction arrow.
    directionLabel(item, up) {
      return sortLabel(selectSort(this.modelValue, item, up), item);
    },
    // choose closes the menu and emits only a changed sort selection.
    choose(item, up) {
      if (this.disabled) {
        return;
      }
      const next = selectSort(this.modelValue, item, up);
      this.open = false;
      if (next.order !== this.modelValue.order || next.reverse !== this.modelValue.reverse) {
        this.restoreFocus = true;
        this.$emit("update:modelValue", next);
        this.$nextTick(() => {
          if (!this.disabled) {
            this.focusButton(false);
          }
        });
      }
    },
  },
};
</script>
<style scoped>
@layer vuetify-components {
  .p-sort-menu {
    width: min(340px, calc(100vw - 24px));
  }
}
</style>
