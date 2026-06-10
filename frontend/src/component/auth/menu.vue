<template>
  <v-menu
    v-model="open"
    location="top"
    offset="6"
    :min-width="250"
    :z-index="2500"
    :open-on-hover="openOnHover"
    class="p-action-menu action-menu nav-user-menu"
    content-class="nav-user-menu__content"
    @update:model-value="onToggle"
  >
    <template #activator="{ props }">
      <div v-bind="props" class="nav-user-activator clickable" :title="$gettext('Account')">
        <slot></slot>
      </div>
    </template>
    <v-list slim nav density="compact" bg-color="navigation" class="action-menu__list">
      <template v-if="instances.length > 0">
        <v-list-item
          v-for="instance in instances"
          :key="instance.namespace"
          :title="instance.title"
          :subtitle="instance.path || undefined"
          :nav="true"
          class="action-menu__item"
          @click="onSwitch(instance)"
        >
          <template #prepend>
            <img :src="instance.icon" :alt="instance.title" class="nav-user-menu__icon" @error="onIconError(instance)" />
          </template>
        </v-list-item>
        <v-divider class="my-1"></v-divider>
      </template>
      <v-list-item
        v-if="$config.feature('account')"
        prepend-icon="mdi-shield-account-variant"
        :title="$gettext('Manage Account')"
        class="action-menu__item action-account"
        @click="$emit('account')"
      ></v-list-item>
      <v-list-item prepend-icon="mdi-power" :title="$gettext('Sign Out')" class="action-menu__item action-logout" @click="$emit('logout')"></v-list-item>
    </v-list>
  </v-menu>
</template>

<script>
import { listReachableInstances, instancePath } from "common/instances";

// defaultIcon is shown when a peer has no recorded app icon, or its icon fails to
// load; it is the bundled PhotoPrism logo served at the origin root.
const defaultIcon = "/static/icons/logo.svg";

// PAuthMenu is the navigation avatar overlay menu: Manage Account, Sign Out, and —
// when more than one instance is reachable on a shared domain — a Switch Instance
// list. It emits "account" and "logout" for the host navigation to handle.
export default {
  name: "PAuthMenu",
  emits: ["logout", "account"],
  data() {
    return {
      open: false,
      openOnHover: this.$util.shouldOpenOnHover(),
      instances: [],
    };
  },
  mounted() {
    this.refresh();
  },
  methods: {
    // refresh rescans shared localStorage for reachable peer instances, adding a
    // display path (the SiteUrl's base path) for the menu subtitle.
    refresh() {
      const namespace = this.$config?.values?.storageNamespace;
      this.instances = listReachableInstances({ currentNamespace: namespace }).map((instance) => {
        // Only proxied instances (e.g. "/i/pro-1") get a path subtitle; a root
        // ("/") or empty path means the Portal itself, so leave it blank.
        const path = instancePath(instance.url);
        return {
          ...instance,
          path: path && path !== "/" ? path : "",
          icon: instance.icon || defaultIcon,
        };
      });
    },
    // onToggle rescans peers each time the menu opens so the switcher stays current.
    onToggle(open) {
      if (open) {
        this.refresh();
      }
    },
    // onSwitch navigates the browser to the selected instance's app entry point
    // (its frontend route, e.g. /i/pro-1/library or /portal) so the user lands
    // in the app instead of a web-overlay landing page; its own login then completes
    // the standard OIDC flow when a session is required. Falls back to the SiteUrl.
    onSwitch(instance) {
      const target = instance && (instance.route || instance.url);
      if (target) {
        this.navigate(target);
      }
    },
    // navigate sends the browser to the given URL (a seam so onSwitch is testable).
    navigate(url) {
      window.location.assign(url);
    },
    // onIconError falls back to the default logo when an instance's recorded app
    // icon fails to load (e.g. a removed asset or one served cross-origin). The
    // guard avoids an error loop if the default logo itself fails.
    onIconError(instance) {
      if (instance.icon !== defaultIcon) {
        instance.icon = defaultIcon;
      }
    },
  },
};
</script>
