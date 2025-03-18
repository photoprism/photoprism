<template>
  <v-menu
    class="p-action-menu action-menu"
    transition="slide-y-transition"
    open-on-click
    open-on-hover
    @update:model-value="onShow"
  >
    <template #activator="{ props }">
      <v-btn
        density="comfortable"
        icon="mdi-dots-vertical"
        v-bind="props"
        class="action-menu__btn"
        :class="buttonClass"
      ></v-btn>
    </template>

    <v-list slim nav density="compact" bg-color="navigation" class="action-menu__list">
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
      </v-list-item>
    </v-list>
  </v-menu>
</template>
<script>
export default {
  name: "PActionMenu",
  props: {
    items: {
      type: Function,
      default: () => [],
    },
    buttonClass: {
      type: String,
      default: "",
    },
    buttonIcon: {
      type: String,
      default: "mdi-dots-vertical",
    },
  },
  data() {
    return {
      actions: [],
    };
  },
  methods: {
    onShow(visible) {
      if (visible) {
        this.actions = this.items().filter((action) => action.visible);
      }
    },
  },
};
</script>
