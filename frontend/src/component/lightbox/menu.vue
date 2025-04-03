<template>
  <div v-if="activator" class="p-lightbox-menu">
    <v-menu
      :model-value="visible"
      :open-on-hover="openOnHover"
      :activator="activator"
      :attach="attach"
      class="p-action-menu action-menu action-menu--lightbox"
      @update:model-value="onMenu"
    >
      <v-list slim nav density="compact" class="action-menu__list">
        <v-list-item
          v-for="action in actions"
          :key="action.name"
          :value="action.name"
          :prepend-icon="action.icon"
          :title="action.text"
          :class="action.class ? action.class : 'action-' + action.name"
          :to="action.to ? action.to : undefined"
          :href="action.href ? action.href : undefined"
          :link="true"
          :target="action.target ? '_blank' : '_self'"
          :disabled="action.disabled"
          :nav="true"
          class="action-menu__item"
          @click="action.click"
        >
          <template v-if="action.shortcut && !$isMobile" #append>
            <div class="action-menu__shortcut">{{ action.shortcut }}</div>
          </template>
        </v-list-item>
      </v-list>
    </v-menu>
  </div>
</template>
<script>
export default {
  name: "PLightboxMenu",
  props: {
    items: {
      type: Function,
      default: () => [],
    },
    activator: {
      type: HTMLElement,
      default: null,
    },
    attach: {
      type: String,
      default: ".v-dialog--lightbox.v-overlay--active",
    },
  },
  emits: ["show", "hide"],
  data() {
    return {
      visible: false,
      actions: [],
      openOnHover: !this.$util.hasTouch(),
    };
  },
  methods: {
    show() {
      this.actions = this.items().filter((action) => action.visible);
      this.visible = true;
      this.$emit("show");
    },
    hide() {
      this.actions = [];
      this.visible = false;
      this.$emit("hide");
    },
    toggle() {
      if (this.visible) {
        this.hide();
      } else {
        this.show();
      }
    },
    onMenu(show) {
      if (show) {
        this.show();
      } else {
        this.hide();
      }
    },
  },
};
</script>
