<template>
  <div class="p-page p-page-settings" :class="$config.aclClasses('settings')">
    <v-tabs v-model="active" flat grow touchless color="secondary" slider-color="secondary-dark" :height="$vuetify.breakpoint.smAndDown ? 48 : 64">
      <v-tab v-for="(item, index) in tabs" :id="'tab-' + item.name" :key="index" :class="item.class" ripple @click="changePath(item.path)">
        <v-icon v-if="$vuetify.breakpoint.smAndDown" :title="item.label">{{ item.icon }}</v-icon>
        <template v-else>
          <v-icon :size="18" :left="!rtl" :right="rtl">{{ item.icon }}</v-icon>
          {{ item.label }}
        </template>
      </v-tab>

      <v-tabs-items touchless>
        <v-tab-item v-for="(item, index) in tabs" :key="index" lazy>
          <component :is="item.component"></component>
        </v-tab-item>
      </v-tabs-items>
    </v-tabs>
  </div>
</template>

<script>
import General from "page/settings/general.vue";
import Library from "page/settings/library.vue";
import Advanced from "page/settings/advanced.vue";
import Services from "page/settings/services.vue";
import Account from "page/settings/account.vue";
import { config } from "app/session";

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
    const isSuperAdmin = this.$session.isSuperAdmin();

    const tabs = [
      {
        name: "settings_general",
        component: General,
        label: this.$gettext("General"),
        class: "",
        path: "/settings",
        icon: "tv",
        public: true,
        admin: true,
        demo: true,
        show: config.feature("settings"),
      },
      {
        name: "settings_media",
        component: Library,
        label: this.$gettext("Library"),
        class: "",
        path: "/settings/media",
        icon: "camera_roll",
        public: true,
        admin: true,
        demo: true,
        show: config.allow("config", "manage") && isSuperAdmin,
      },
      {
        name: "settings_advanced",
        component: Advanced,
        label: this.$gettext("Advanced"),
        class: "",
        path: "/settings/advanced",
        icon: "build",
        public: false,
        admin: true,
        demo: true,
        show: config.allow("config", "manage"),
      },
      {
        name: "settings_services",
        component: Services,
        label: this.$gettext("Services"),
        class: "",
        path: "/settings/services",
        icon: "sync_alt",
        public: false,
        admin: true,
        demo: true,
        show: config.feature("services") && config.allow("services", "manage"),
      },
      {
        name: "settings_account",
        component: Account,
        label: this.$gettext("Account"),
        class: "",
        path: "/settings/account",
        icon: "admin_panel_settings",
        public: false,
        admin: true,
        demo: true,
        show: config.feature("account"),
      },
    ];

    if (isDemo) {
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
      rtl: this.$rtl,
    };
  },
  watch: {
    $route() {
      let active = this.active;

      if (typeof this.$route.name === "string" && this.$route.name !== "") {
        active = this.tabs.findIndex((t) => t.name === this.$route.name);
      }

      if (active >= 0) {
        this.active = active;
      }
    },
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
