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
      <v-tab v-for="(item, index) in tabs" :id="'tab-' + item.name" :key="index" :class="item.class" ripple
             @click="changePath(item.path)">
        <v-icon v-if="$vuetify.breakpoint.smAndDown" :title="item.label">{{ item.icon }}</v-icon>
        <template v-else>
          <v-icon :size="18" :left="!rtl" :right="rtl">{{ item.icon }}</v-icon> {{ item.label }}
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
import Account from "pages/settings/account.vue";

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
  name: 'PPageProfile',
  props: {
    tab: String,
  },
  data() {
    const isDemo = this.$config.get("demo");
    const isPublic = this.$config.get("public");
    const tabs = [
      {
        'name': 'settings-account',
        'component': Account,
        'label': this.$gettext('Account'),
        'class': '',
        'path': '/account',
        'icon': 'person',
        'public': false,
        'admin': true,
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
      rtl: this.$rtl,
    };
  },
  methods: {
    changePath: function (path) {
      if (this.$route.path !== path) {
        this.$router.replace(path);
      }
    }
  },
};
</script>
