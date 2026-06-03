<template>
  <v-menu
    v-model="open"
    location="top"
    offset="6"
    :min-width="240"
    class="nav-user-menu"
    @update:model-value="onToggle"
  >
    <template #activator="{ props }">
      <div v-bind="props" class="nav-user-activator clickable" :title="$gettext('Account')">
        <slot></slot>
      </div>
    </template>
    <v-list slim nav density="compact" bg-color="navigation" class="nav-user-menu__list">
      <v-list-item prepend-icon="mdi-cog-outline" :title="$gettext('Settings')" class="action-settings" @click="$emit('settings')"></v-list-item>
      <template v-if="instances.length > 0">
        <v-divider class="my-1"></v-divider>
        <v-list-subheader class="nav-user-menu__subheader">{{ $gettext("Switch Instance") }}</v-list-subheader>
        <v-list-item
          v-for="instance in instances"
          :key="instance.namespace"
          prepend-icon="mdi-server"
          :title="instance.title"
          :subtitle="instance.url"
          class="action-switch-instance"
          @click="onSwitch(instance)"
        ></v-list-item>
      </template>
      <v-divider class="my-1"></v-divider>
      <v-list-item prepend-icon="mdi-power" :title="$gettext('Log Out')" base-color="danger" class="action-logout" @click="$emit('logout')"></v-list-item>
    </v-list>
  </v-menu>
</template>

<script>
import { listReachableInstances } from "common/instances";

// PUserMenu is the navigation avatar overlay menu: Settings, Log Out, and — when
// more than one instance is reachable on a shared domain — a Switch Instance list.
export default {
  name: "PUserMenu",
  emits: ["settings", "logout"],
  data() {
    return {
      open: false,
      instances: [],
    };
  },
  mounted() {
    this.refresh();
  },
  methods: {
    // refresh rescans shared localStorage for reachable peer instances.
    refresh() {
      const namespace = this.$config?.values?.storageNamespace;
      this.instances = listReachableInstances({ currentNamespace: namespace });
    },
    // onToggle rescans peers each time the menu opens so the switcher stays current.
    onToggle(open) {
      if (open) {
        this.refresh();
      }
    },
    // onSwitch navigates the browser to the selected instance's site URL, whose
    // own login completes the standard OIDC flow when a session is required.
    onSwitch(instance) {
      if (instance && instance.url) {
        this.navigate(instance.url);
      }
    },
    // navigate sends the browser to the given URL (a seam so onSwitch is testable).
    navigate(url) {
      window.location.assign(url);
    },
  },
};
</script>
