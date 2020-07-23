<template>
  <div class="p-page p-page-settings">
    <v-tabs
            v-model="active"
            flat
            grow
            touchless
            color="secondary"
            slider-color="secondary-dark"
            height="64"
    >
      <v-tab id="tab-settings-general" ripple @click="changePath('/settings')">
        <translate key="General">General</translate>
      </v-tab>

      <v-tab id="tab-settings-sync" ripple @click="changePath('/settings/sync')">
        <translate key="Backup">Backup</translate>
      </v-tab>

      <v-tab id="tab-settings-account" ripple @click="changePath('/settings/account')" v-if="!public">
        <translate key="Account">Account</translate>
      </v-tab>

      <v-tabs-items touchless>
        <v-tab-item lazy>
          <p-settings-general></p-settings-general>
        </v-tab-item>
        <v-tab-item lazy>
          <p-settings-sync></p-settings-sync>
        </v-tab-item>
        <v-tab-item lazy v-if="!public">
          <p-settings-account></p-settings-account>
        </v-tab-item>
      </v-tabs-items>
    </v-tabs>
  </div>
</template>

<script>
    import tabGeneral from "pages/settings/general.vue";
    import tabSync from "pages/settings/sync.vue";
    import tabAccount from "pages/settings/account.vue";

    export default {
        name: 'p-page-settings',
        props: {
            tab: Number
        },
        components: {
            'p-settings-general': tabGeneral,
            'p-settings-sync': tabSync,
            'p-settings-account': tabAccount,
        },
        data() {
            return {
                public: this.$config.get("public"),
                readonly: this.$config.get("readonly"),
                active: this.tab,
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
