<template>
  <div class="p-input-chip-selector chip-selector">
    <div v-if="shouldRenderChips" class="chip-selector__chips">
      <v-tooltip
        v-for="item in processedItems"
        :key="item.value || item.title"
        :text="getChipTooltip(item)"
        location="top"
        :disabled="$vuetify?.display?.mobile === true"
      >
        <template #activator="{ props }">
          <div
            v-bind="props"
            :class="getChipClasses(item)"
            :aria-pressed="item.selected"
            :tabindex="0"
            role="button"
            @click="handleChipClick(item)"
            @keydown.enter="handleChipClick(item)"
            @keydown.space.prevent="handleChipClick(item)"
          >
            <div class="chip__content">
              <v-icon v-if="getChipIcon(item)" class="chip__icon">
                {{ getChipIcon(item) }}
              </v-icon>
              <span class="chip__text">{{ item.title }}</span>
            </div>
          </div>
        </template>
      </v-tooltip>

      <div v-if="processedItems.length === 0 && !showInput" class="chip-selector__empty">
        {{ emptyText }}
      </div>
    </div>

    <div v-if="allowCreate" class="chip-selector__input-container">
      <v-combobox
        ref="inputField"
        v-model="newItemTitle"
        v-model:menu="menuOpen"
        :placeholder="computedInputPlaceholder"
        :persistent-placeholder="true"
        :items="availableItems"
        item-title="title"
        item-value="value"
        density="comfortable"
        hide-details
        hide-no-data
        return-object
        class="chip-selector__input"
        @keydown.enter.prevent="onEnter"
        @blur="addNewItem"
        @update:model-value="onComboboxChange"
        @update:menu="onMenuUpdate"
      >
        <template #no-data>
          <v-list-item>
            <v-list-item-title>
              {{ $gettext("Press enter to create new item") }}
            </v-list-item-title>
          </v-list-item>
        </template>
      </v-combobox>
    </div>
  </div>
</template>

<script>
export default {
  name: "PInputChipSelector",
  props: {
    items: {
      type: Array,
      default: () => [],
    },
    normalizeTitleForCompare: {
      type: Function,
      default: null,
    },
    availableItems: {
      type: Array,
      default: () => [],
    },
    allowCreate: {
      type: Boolean,
      default: true,
    },
    emptyText: {
      type: String,
      default: "",
    },
    inputPlaceholder: {
      type: String,
      default: "",
    },
    loading: {
      type: Boolean,
      default: false,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    resolveItemFromText: {
      type: Function,
      default: null,
    },
  },
  emits: ["update:items"],
  data() {
    return {
      newItemTitle: null,
      menuOpen: false,
      suppressMenuOpen: false,
    };
  },
  computed: {
    processedItems() {
      return this.items.map((item) => ({
        ...item,
        // Ensure action is always a string, never null/undefined
        action: item.action || "none",
        selected: item.action === "add" || item.action === "remove",
      }));
    },
    computedInputPlaceholder() {
      return this.inputPlaceholder || this.$gettext("Enter item name...");
    },
    showInput() {
      return this.allowCreate;
    },
    shouldRenderChips() {
      // Render chips container only when there are chips
      return this.processedItems.length > 0 || !this.showInput;
    },
  },
  methods: {
    normalizeTitle(text) {
      const input = text == null ? "" : String(text);
      if (typeof this.normalizeTitleForCompare === "function") {
        try {
          return this.normalizeTitleForCompare(input);
        } catch (e) {
          return input.toLowerCase();
        }
      }
      return input.toLowerCase();
    },

    getChipClasses(item) {
      const baseClass = "chip";
      const classes = [baseClass];

      if (this.loading || this.disabled) {
        classes.push(`${baseClass}--loading`);
      }

      if (item.action === "add") {
        classes.push(item.mixed ? `${baseClass}--green-light` : `${baseClass}--green`);
      } else if (item.action === "remove") {
        classes.push(item.mixed ? `${baseClass}--red-light` : `${baseClass}--red`);
      } else if (item.mixed) {
        classes.push(`${baseClass}--gray-light`);
      } else {
        classes.push(`${baseClass}--gray`);
      }

      return classes;
    },

    getChipIcon(item) {
      if (item.action === "add") return "mdi-plus";
      if (item.action === "remove") return "mdi-minus";
      if (item.mixed) return "mdi-circle-half-full";
      return null;
    },

    getChipTooltip(item) {
      if (item.action === "add") {
        return item.mixed ? this.$gettext("Add to all selected photos") : this.$gettext("Add to all");
      } else if (item.action === "remove") {
        return item.mixed ? this.$gettext("Remove from all selected photos") : this.$gettext("Remove from all");
      } else if (item.mixed) {
        return this.$gettext("Part of some selected photos");
      }
      return this.$gettext("Part of all selected photos");
    },

    handleChipClick(item) {
      if (this.loading || this.disabled) return;

      let newAction;

      if (item.mixed) {
        // Handle mixed state cycling
        switch (item.action) {
          case "none":
            newAction = "add";
            break;
          case "add":
            newAction = "remove";
            break;
          case "remove":
            newAction = "none";
            break;
        }
      } else {
        // Handle normal state cycling
        if (item.isNew) {
          newAction = item.action === "add" ? "remove" : "add";
        } else {
          newAction = item.action === "remove" ? "none" : "remove";
        }
      }

      this.updateItemAction(item, newAction);
    },

    updateItemAction(itemToUpdate, action) {
      // Special case: remove new items completely
      if (itemToUpdate.isNew && action === "remove") {
        const updatedItems = this.items.filter(
          (item) => (item.value || item.title) !== (itemToUpdate.value || itemToUpdate.title)
        );
        this.$emit("update:items", updatedItems);
        return;
      }

      // Update action for existing item
      const updatedItems = this.items.map((item) =>
        (item.value || item.title) === (itemToUpdate.value || itemToUpdate.title) ? { ...item, action } : item
      );

      this.$emit("update:items", updatedItems);
    },

    onComboboxChange(value) {
      if (value && typeof value === "object" && value.title) {
        this.newItemTitle = value;
        this.addNewItem();
        this.menuOpen = false;
        // Immediately clear the input, remove focus and restore placeholder
        this.$nextTick(() => {
          this.newItemTitle = "";
          if (this.$refs.inputField) {
            this.$refs.inputField.blur();
            // Force the combobox to reset completely
            setTimeout(() => {
              this.newItemTitle = null;
            }, 10);
          }
        });
      } else {
        this.newItemTitle = value;
      }
    },

    addNewItem() {
      // Extract title and value from input
      let title, value;

      if (typeof this.newItemTitle === "string") {
        title = this.newItemTitle.trim();
        value = "";
      } else if (this.newItemTitle && typeof this.newItemTitle === "object") {
        title = this.newItemTitle.title;
        value = this.newItemTitle.value;
      } else {
        return;
      }

      if (!title) return;

      let resolvedApplied = false;
      if (typeof this.resolveItemFromText === "function") {
        const resolved = this.resolveItemFromText(title);
        if (resolved && typeof resolved === "object") {
          if (resolved.title) title = resolved.title;
          if (resolved.value) value = resolved.value;
          resolvedApplied = true;
        }
      }

      const normalizedTitle = this.normalizeTitle(title);
      const existingItem = this.items.find(
        (item) => (item.value && value && item.value === value) || this.normalizeTitle(item.title) === normalizedTitle
      );

      if (existingItem) {
        let changed = false;
        const updatedItems = this.items.map((item) => {
          const isSame =
            (item.value && value && item.value === value) || this.normalizeTitle(item.title) === normalizedTitle;
          if (!isSame) {
            return item;
          }

          const next = { ...item };
          let itemChanged = false;

          if (resolvedApplied) {
            if (value && item.value !== value) {
              next.value = value;
              itemChanged = true;
            }
            if (title && item.title !== title) {
              next.title = title;
              itemChanged = true;
            }
          }

          if (item.mixed && item.action !== "add") {
            next.action = "add";
            next.mixed = false;
            itemChanged = true;
          }

          if (itemChanged) {
            changed = true;
            return next;
          }

          return item;
        });

        if (changed) {
          this.$emit("update:items", updatedItems);
        }

        this.resetInputField();
        return;
      }

      const newItem = {
        value: value || "",
        title,
        mixed: false,
        action: "add",
        isNew: true,
      };

      this.$emit("update:items", [...this.items, newItem]);
      this.newItemTitle = null;
      this.menuOpen = false;

      // Refocus input field
      this.$nextTick(() => {
        if (this.$refs.inputField) {
          this.$refs.inputField.focus();
        }
      });
    },

    onEnter() {
      this.suppressMenuOpen = true;
      this.menuOpen = false;
      this.addNewItem();
      window.setTimeout(() => {
        this.suppressMenuOpen = false;
      }, 250);
    },

    onMenuUpdate(val) {
      if (val && this.suppressMenuOpen) {
        this.menuOpen = false;
        return;
      }
      this.menuOpen = val;
    },
    resetInputField() {
      this.menuOpen = false;
      this.$nextTick(() => {
        this.newItemTitle = "";
        if (this.$refs.inputField) {
          this.$refs.inputField.blur();
          setTimeout(() => {
            this.newItemTitle = null;
          }, 10);
        } else {
          this.newItemTitle = null;
        }
      });
    },
  },
};
</script>

<style src="../../css/chip-selector.css"></style>
