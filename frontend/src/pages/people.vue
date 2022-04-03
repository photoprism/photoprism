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
      <v-tab v-for="(item, index) in tabs" :id="'tab-' + item.name" :key="index" :class="item.class"
             ripple @click.stop.prevent="changePath(item.path)">
        <v-icon v-if="$vuetify.breakpoint.smAndDown" :title="item.label">{{ item.icon }}</v-icon>
        <template v-else>
          <v-icon :size="18" :left="!rtl" :right="rtl">{{ item.icon }}</v-icon>
          <v-badge color="secondary-dark" :left="rtl" :right="!rtl">
            <template #badge>
              <span v-if="item.count">{{ item.count }}</span>
            </template>
            {{ item.label }}
          </v-badge>
        </template>
      </v-tab>

      <v-tabs-items touchless>
        <v-tab-item v-for="(item, index) in tabs" :key="index" :transition="false" class="no-transition" lazy>
          <component :is="item.component" :static-filter="item.filter" :active="active === index"
                     @updateFaceCount="onUpdateFaceCount"></component>
        </v-tab-item>
      </v-tabs-items>
    </v-tabs>
  </div>
</template>

<script>
import Subjects from "pages/people/subjects.vue";
import Faces from "pages/people/faces.vue";

export default {
  name: 'PPagePeople',
  data() {
    const config = this.$config.values;
    const isDemo = this.$config.get("demo");
    const isPublic = this.$config.get("public");
    const isReadOnly = this.$config.get("readonly");

    const tabs = [
      {
        'name': 'people',
        'component': Subjects,
        'filter': {files: 1, type: "person"},
        'label': this.$gettext('Recognized'),
        'class': '',
        'path': '/people',
        'icon': 'people_alt',
      },
      {
        'name': 'people_faces',
        'component': Faces,
        'filter': {markers: true, unknown: true},
        'label': this.$gettext('New'),
        'class': '',
        'path': '/people/new',
        'icon': 'person_add',
        'count': 0,
      },
    ];

    return {
      tabs: tabs,
      demo: isDemo,
      public: isPublic,
      config: config,
      readonly: isReadOnly,
      active: 0,
      rtl: this.$rtl,
    };
  },
  watch: {
    '$route'() {
      this.openTab();
    }
  },
  created() {
    this.openTab();
  },
  methods: {
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
    }
  }
};
</script>
