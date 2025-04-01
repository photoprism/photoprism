<template>
  <v-menu
    v-if="activator"
    transition="slide-y-transition"
    :activator="activator"
    open-on-click
    open-on-hover
    class="p-lightbox-menu p-action-menu action-menu action-menu--lightbox"
    @update:model-value="onShow"
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
  },
  emits: ["show", "hide"],
  data() {
    return {
      actions: [],
    };
  },
  methods: {
    onShow(visible) {
      if (visible) {
        this.$emit("show");
        this.actions = this.items().filter((action) => action.visible);
      } else {
        this.$emit("hide");
      }
    },
  },
};
</script>
