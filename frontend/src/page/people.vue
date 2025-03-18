<template>
  <div
    ref="page"
    tabindex="1"
    class="p-page p-page-people"
    :class="$config.aclClasses('people')"
    @keydown.ctrl="onCtrl"
  >
    <v-tabs
      v-model="active"
      elevation="0"
      grow
      bg-color="secondary"
      slider-color="surface-variant"
      :height="$vuetify.display.smAndDown ? 48 : 64"
      class="bg-transparent p-page__navigation"
    >
      <v-tab
        v-for="t in tabs"
        :id="'tab-' + t.name"
        :key="t.name"
        :class="t.class"
        ripple
        @click.stop.prevent="changePath(t.path)"
      >
        <v-icon v-if="$vuetify.display.smAndDown" :title="t.label">{{ t.icon }}</v-icon>
        <template v-else>
          <v-icon :size="18" start>{{ t.icon }}</v-icon>
          {{ t.label }}
        </template>
        <v-badge v-if="t.count" color="surface-variant" inline :content="t.count"></v-badge>
      </v-tab>
    </v-tabs>

    <v-tabs-window v-model="active">
      <v-tabs-window-item v-for="(t, index) in tabs" :key="t.name" eager>
        <component
          :is="t.component"
          :static-filter="t.filter"
          :active="active === index"
          @updateFaceCount="onUpdateFaceCount"
        ></component>
      </v-tabs-window-item>
    </v-tabs-window>
  </div>
</template>

<script>
import Recognized from "page/people/recognized.vue";
import NewFaces from "page/people/new.vue";
import { markRaw } from "vue";

export default {
  name: "PPagePeople",
  data() {
    const config = this.$config.values;
    const isDemo = this.$config.get("demo");
    const isPublic = this.$config.get("public");
    const isReadOnly = this.$config.get("readonly");

    const tabs = [
      {
        name: "people",
        component: markRaw(Recognized),
        filter: { files: 1, type: "person" },
        label: this.$gettext("Recognized"),
        class: "",
        path: "/people",
        icon: "mdi-account-multiple",
        count: 0,
      },
    ];

    if (this.$config.allow("people", "manage")) {
      tabs.push({
        name: "people_faces",
        component: markRaw(NewFaces),
        filter: { markers: true, unknown: true },
        label: this.$gettext("New"),
        class: "",
        path: "/people/new",
        icon: "mdi-account-plus",
        count: 0,
      });
    }

    let active = 0;

    if (typeof this.$route.name === "string" && this.$route.name !== "") {
      active = tabs.findIndex((t) => t.name === this.$route.name);
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
      rtl: this.$isRtl,
    };
  },
  watch: {
    $route() {
      if (!this.$view.isActive(this)) {
        return;
      }

      this.$view.focus(this.$refs?.page);

      this.openTab();
    },
  },
  mounted() {
    this.$view.enter(this, this.$refs?.page);
    this.openTab();
  },
  unmounted() {
    this.$view.leave(this);
  },
  methods: {
    onCtrl(ev) {
      if (!ev || !(ev instanceof KeyboardEvent) || !ev.ctrlKey || !this.$view.isActive(this)) {
        return;
      }

      switch (ev.code) {
        case "KeyF":
          ev.preventDefault();
          this.$view.focus(this.$refs?.page, ".v-window-item--active .input-search input", true);
          break;
      }
    },
    openTab() {
      const activeTab = this.tabs.findIndex((t) => t.name === this.$route.name);

      if (activeTab > -1 && this.active !== activeTab) {
        this.active = activeTab;
      }
    },
    onUpdateFaceCount(count) {
      this.tabs[1].count = count;
    },
    changePath(path) {
      if (this.$route.path !== path) {
        this.$router.replace(path);
      }
    },
  },
};
</script>
