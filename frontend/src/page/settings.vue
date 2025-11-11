<template>
  <div ref="page" tabindex="-1" class="p-page p-page-settings" :class="$config.aclClasses('settings')">
    <v-tabs v-model="active" :height="$vuetify.display.smAndDown ? 48 : 64" class="p-page__navigation">
      <v-tab v-for="t in tabs" :id="'tab-' + t.name" :key="t.name" :class="t.class" ripple @click="changePath(t.path)">
        <v-icon v-if="$vuetify.display.smAndDown" :title="t.label">{{ t.icon }}</v-icon>
        <template v-else>
          <v-icon :size="18" start>{{ t.icon }}</v-icon>
          {{ t.label }}
        </template>
      </v-tab>
    </v-tabs>

    <v-tabs-window v-model="active">
      <v-tabs-window-item v-for="(t, index) in tabs" :key="index">
        <component :is="t.component"></component>
      </v-tabs-window-item>
    </v-tabs-window>
  </div>
</template>

<script>
import General from "page/settings/general.vue";
import Content from "page/settings/content.vue";
import Advanced from "page/settings/advanced.vue";
import Services from "page/settings/services.vue";
import Account from "page/settings/account.vue";
import { $config } from "app/session";
import { markRaw } from "vue";

function initTabs(flag, tabs) {
  let i = 0;
  while (i < tabs.length) {
    if (!tabs[i][flag]) {
      tabs.splice(i, 1);
    } else {
      i++;
    }
  }
}

export default {
  name: "PPageSettings",
  props: {
    tab: {
      type: String,
      default: "",
    },
  },
  data() {
    const isDemo = this.$config.isDemo();
    const isPublic = this.$config.isPublic();
    const isPortal = this.$config.isPortal();
    const isSuperAdmin = this.$session.isSuperAdmin();
    const hasScope = this.$session.hasScope();

    const tabs = [
      {
        name: "settings_general",
        component: markRaw(General),
        label: this.$gettext("General"),
        class: "",
        path: "/settings",
        icon: "mdi-television",
        public: true,
        portal: true,
        admin: true,
        demo: true,
        show: $config.feature("settings"),
      },
      {
        name: "settings_content",
        component: markRaw(Content),
        label: this.$gettext("Content"),
        class: "",
        path: "/settings/content",
        icon: "mdi-camera-iris",
        public: true,
        portal: false,
        admin: true,
        demo: true,
        show: $config.feature("settings") && !hasScope,
      },
      {
        name: "settings_advanced",
        component: markRaw(Advanced),
        label: this.$gettext("Advanced"),
        class: "",
        path: "/settings/advanced",
        icon: "mdi-wrench",
        public: false,
        portal: true,
        admin: true,
        demo: true,
        show: $config.allow("config", "manage") && isSuperAdmin,
      },
      {
        name: "settings_services",
        component: markRaw(Services),
        label: this.$gettext("Services"),
        class: "",
        path: "/settings/services",
        icon: "mdi-swap-horizontal",
        public: false,
        portal: false,
        admin: true,
        demo: true,
        show: $config.feature("services") && $config.allow("services", "manage"),
      },
      {
        name: "settings_account",
        component: markRaw(Account),
        label: this.$gettext("Account"),
        class: "",
        path: "/settings/account",
        icon: "mdi-shield-account-variant",
        public: false,
        portal: true,
        admin: true,
        demo: true,
        show: $config.feature("account"),
      },
    ];

    if (isPortal) {
      initTabs("portal", tabs);
    } else if (isDemo) {
      initTabs("demo", tabs);
    } else if (isPublic) {
      initTabs("public", tabs);
    } else {
      initTabs("show", tabs);
    }

    let active = 0;

    if (typeof this.$route.name === "string" && this.$route.name !== "") {
      active = tabs.findIndex((t) => t.name === this.$route.name);
    } else if (typeof this.tab === "string" && this.tab !== "") {
      active = tabs.findIndex((t) => t.name === this.tab);
    }

    if (active < 0) {
      active = 0;
    }

    return {
      tabs: tabs,
      demo: isDemo,
      public: isPublic,
      readonly: this.$config.get("readonly"),
      active: active,
      rtl: this.$isRtl,
    };
  },
  watch: {
    $route() {
      if (!this.$view.isActive(this)) {
        return;
      }

      this.$view.focus(this.$refs?.page);

      let active = this.active;

      if (typeof this.$route.name === "string" && this.$route.name !== "") {
        active = this.tabs.findIndex((t) => t.name === this.$route.name);
      }

      if (active >= 0) {
        this.active = active;
      }
    },
  },
  mounted() {
    this.$view.enter(this);
  },
  unmounted() {
    this.$view.leave(this);
  },
  created() {
    if (!this.tabs || this.tabs.length === 0) {
      this.$router.push({ name: "albums" });
    }
  },
  methods: {
    changePath: function (path) {
      if (this.$route.path !== path) {
        this.$router.replace(path);
      }
    },
  },
};
</script>
