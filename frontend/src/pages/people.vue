<template>
  <div class="p-page p-page-people">
    <v-tabs
        v-model="active"
        flat
        grow
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
          <component :is="item.component" :static-filter="item.filter"></component>
        </v-tab-item>
      </v-tabs-items>
    </v-tabs>
  </div>
</template>

<script>
import Subjects from "pages/people/subjects.vue";
import Faces from "pages/people/faces.vue";

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
  name: 'PPagePeople',
  props: {
    tab: String,
  },
  data() {
    const config = this.$config.values;
    const isDemo = this.$config.get("demo");
    const isPublic = this.$config.get("public");
    const isReadOnly = this.$config.get("readonly");

    const tabs = [
      {
        'name': 'people-subjects',
        'component': Subjects,
        'filter': { files: 1, type: "person" },
        'label': this.$gettext('Recognized'),
        'class': '',
        'path': '/people',
        'icon': 'people_alt',
        'readonly': true,
        'demo': true,
      },
      {
        'name': 'people-faces',
        'component': Faces,
        'filter': { markers: true, unknown: true },
        'label': this.$gettext('New'),
        'class': '',
        'path': '/people/new',
        'icon': 'person_add',
        'readonly': true,
        'demo': true,
      },
    ];

    let active = 0;

    if (typeof this.tab === 'string' && this.tab !== '') {
      active = tabs.findIndex((t) => t.name === this.tab);
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
  methods: {
    changePath: function (path) {
      if (this.$route.path !== path) {
        this.$router.replace(path);
      }
    }
  }
};
</script>
