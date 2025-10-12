<template>
  <div class="p-action-menu">
    <v-menu
      :model-value="visible"
      :open-on-hover="openOnHover"
      class="action-menu action-menu--default"
      @update:model-value="onMenu"
    >
      <template #activator="{ props }">
        <v-btn
          v-bind="props"
          density="comfortable"
          :icon="buttonIcon"
          :tabindex="tabindex"
          class="action-menu__btn"
          :class="buttonClass"
        ></v-btn>
      </template>

      <v-list slim nav density="compact" bg-color="navigation" class="action-menu__list" :class="listClass">
        <template v-for="action in actions" :key="action.name">
          <v-divider v-if="action?.color === 'danger'"></v-divider>
          <v-list-item
            :value="action.name"
            :prepend-icon="action.icon"
            :title="action.text"
            :base-color="action.color"
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
        </template>
      </v-list>
    </v-menu>
  </div>
</template>
<script>
export default {
  name: "PActionMenu",
  props: {
    items: {
      type: Function,
      default: () => [],
    },
    tabindex: {
      type: Number,
      default: 3,
    },
    listClass: {
      type: String,
      default: "",
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
  emits: ["show", "hide"],
  expose: ["show", "hide", "toggle"],
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
