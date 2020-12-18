<template>
  <div class="p-page p-page-settings">
    <v-tabs
        v-model="active"
        flat
        grow
        touchless
        color="secondary"
        slider-color="secondary-dark"
        :height="$vuetify.breakpoint.smAndDown ? 48 : 64"
    >
      <v-tab v-for="(tab, index) in tabs" :key="index" :id="'tab-' + tab.name" :class="tab.class" ripple
             @click="changePath(tab.path)">
        <v-icon v-if="$vuetify.breakpoint.smAndDown" :title="tab.label">{{ tab.icon }}</v-icon>
        <template v-else>
          <v-icon :size="18" left>{{ tab.icon }}</v-icon> {{ tab.label }}
        </template>
      </v-tab>

      <v-tabs-items touchless>
        <v-tab-item lazy v-for="(tab, index) in tabs" :key="index">
          <component v-bind:is="tab.component"></component>
        </v-tab-item>
      </v-tabs-items>
    </v-tabs>
  </div>
</template>

<script>
import tabGeneral from "pages/settings/general.vue";
import tabLibrary from "pages/settings/library.vue";
import tabSync from "pages/settings/sync.vue";
import tabAccount from "pages/settings/account.vue";

function initTabs(flag, tabs) {
  let i = 0;
  while(i < tabs.length) {
    if(!tabs[i][flag]) {
      tabs.splice(i,1);
    } else {
      i++;
    }
  }
}

export default {
  name: 'p-page-settings',
  props: {
    tab: String,
  },
  components: {
    'p-settings-general': tabGeneral,
    'p-settings-library': tabLibrary,
    'p-settings-sync': tabSync,
    'p-settings-account': tabAccount,
  },
  data() {
    const isDemo = this.$config.get("demo");
    const isPublic = this.$config.get("public");
    const tabs = [
      {
        'name': 'settings-general',
        'component': tabGeneral,
        'label': this.$gettext('General'),
        'class': '',
        'path': '/settings',
        'icon': 'tv',
        'public': true,
        'demo': true,
      },
      {
        'name': 'settings-library',
        'component': tabLibrary,
        'label': this.$gettext('Library'),
        'class': '',
        'path': '/settings/library',
        'icon': 'camera_roll',
        'public': true,
        'demo': true,
      },
      {
        'name': 'settings-sync',
        'component': tabSync,
        'label': this.$gettext('Sync'),
        'class': '',
        'path': '/settings/sync',
        'icon': 'sync_alt',
        'public': true,
        'demo': true,
      },
      {
        'name': 'settings-account',
        'component': tabAccount,
        'label': this.$gettext('Account'),
        'class': '',
        'path': '/settings/account',
        'icon': 'person',
        'public': false,
        'demo': true,
      },
    ];

    if(isDemo) {
      initTabs("demo", tabs);
    } else if(isPublic) {
      initTabs("public", tabs);
    }

    let active = 0;

    if (typeof this.tab === 'string' && this.tab !== '') {
      active = tabs.findIndex((t) => t.name === this.tab);
    }

    return {
      tabs: tabs,
      demo: isDemo,
      public: isPublic,
      readonly: this.$config.get("readonly"),
      active: active,
    }
  },
  methods: {
    changePath: function (path) {
      if (this.$route.path !== path) {
        this.$router.replace(path)
      }
    }
  },
};
</script>
