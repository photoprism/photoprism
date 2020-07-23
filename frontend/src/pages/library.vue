<template>
  <div class="p-page p-page-library">
    <v-tabs
            v-model="active"
            flat
            grow
            color="secondary"
            slider-color="secondary-dark"
            height="64"
    >
      <v-tab id="tab-index" ripple @click="changePath('/library')">
        <translate key="Index">Index</translate>
      </v-tab>

      <v-tab id="tab-import" :disabled="readonly || !$config.feature('import')" ripple @click="changePath('/library/import')">
        <template v-if="config.settings.import.move">
          <translate key="Move">Move</translate>
        </template>
        <template v-else>
          <translate key="Copy">Copy</translate>
        </template>
      </v-tab>

      <v-tab id="tab-logs" ripple @click="changePath('/library/logs')" v-if="$config.feature('logs')">
        <translate key="Logs">Logs</translate>
      </v-tab>

      <v-tabs-items touchless>
        <v-tab-item lazy>
          <p-tab-index></p-tab-index>
        </v-tab-item>

        <v-tab-item :disabled="readonly" lazy>
          <p-tab-import></p-tab-import>
        </v-tab-item>

        <v-tab-item  v-if="$config.feature('logs')">
          <p-tab-logs></p-tab-logs>
        </v-tab-item>
      </v-tabs-items>
    </v-tabs>
  </div>
</template>

<script>
    import tabImport from "pages/library/import.vue";
    import tabIndex from "pages/library/index.vue";
    import tabLogs from "pages/library/logs.vue";

    export default {
        name: 'p-page-library',
        props: {
            tab: Number
        },
        components: {
            'p-tab-index': tabIndex,
            'p-tab-import': tabImport,
            'p-tab-logs': tabLogs,
        },
        data() {
            return {
                config: this.$config.values,
                readonly: this.$config.get("readonly"),
                active: this.tab,
            }
        },
        methods: {
            changePath: function(path) {
                if (this.$route.path !== path) {
                    this.$router.replace(path)
                }
            }
        }
    };
</script>
