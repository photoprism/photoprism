<template>
  <div :class="$config.aclClasses('library')" class="p-page p-page-library">
    <v-tabs v-model="active" elevation="0" class="bg-transparent" grow bg-color="secondary" slider-color="secondary-dark" :height="$vuetify.display.smAndDown ? 48 : 64">
      <v-tab v-for="(item, index) in tabs" :id="'tab-' + item.name" :key="index" :class="item.class" ripple @click="changePath(item.path)">
        <v-icon v-if="$vuetify.display.smAndDown" :title="item.label">{{ item.icon }}</v-icon>
        <template v-else>
          <v-icon :size="18" :start="!rtl" :end="rtl">{{ item.icon }}</v-icon> {{ item.label }}
        </template>
      </v-tab>
    </v-tabs>

    <v-tabs-window v-model="active">
      <v-tabs-window-item v-for="(item, index) in tabs" :key="index">
        <component :is="item.component"></component>
      </v-tabs-window-item>
    </v-tabs-window>
  </div>
</template>

<script>
import Import from "page/library/import.vue";
import Index from "page/library/index.vue";
import Logs from "page/library/logs.vue";
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
  name: "PPageLibrary",
  props: {
    tab: {
      type: String,
      default: "",
    },
  },
  data() {
    const config = this.$config.values;
    const isDemo = this.$config.get("demo");
    const isPublic = this.$config.get("public");
    const isReadOnly = this.$config.get("readonly");
    const canImport = this.$config.feature("import") && !isReadOnly;

    const tabs = [
      {
        name: "library_index",
        component: markRaw(Index),
        label: this.$gettext("Index"),
        class: "",
        path: "/index",
        icon: "mdi-film",
        readonly: true,
        demo: true,
      },
      {
        name: "library_import",
        component: markRaw(Import),
        label: this.$gettext("Import"),
        class: "",
        path: "/import",
        icon: "mdi-folder-plus",
        readonly: false,
        demo: true,
      },
    ];

    if (this.$config.feature("logs")) {
      // <!-- TODO: change this icon probably to file-document -->
      tabs.push({
        name: "library_logs",
        component: markRaw(Logs),
        label: this.$gettext("Logs"),
        class: "",
        path: "/logs",
        icon: "feed",
        readonly: true,
        demo: true,
      });
    }

    if (isDemo) {
      initTabs("demo", tabs);
    }

    if (!canImport) {
      initTabs("readonly", tabs);
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
      config: config,
      readonly: isReadOnly,
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
