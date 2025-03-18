<template>
  <div class="p-page p-page-discover" tabindex="1">
    <v-tabs
      v-model="active"
      elevation="0"
      grow
      class="bg-transparent"
      bg-color="secondary"
      slider-color="surface-variant"
      :height="$vuetify.display.smAndDown ? 48 : 64"
    >
      <v-tab id="tab-discover-colors" ripple @click="changePath('/discover')">
        {{ $gettext(`Colors`) }}
      </v-tab>

      <v-tab id="tab-discover-similar" ripple @click="changePath('/discover/similar')">
        {{ $gettext(`Similar`) }}
      </v-tab>

      <v-tab id="tab-discover-season" ripple @click="changePath('/discover/season')">
        {{ $gettext(`Season`) }}
      </v-tab>

      <v-tab id="tab-discover-random" ripple @click="changePath('/discover/random')">
        {{ $gettext(`Random`) }}
      </v-tab>

      <v-tabs-window v-model="active">
        <v-tabs-window-item>
          <p-tab-discover-colors></p-tab-discover-colors>
        </v-tabs-window-item>

        <v-tabs-window-item>
          <p-tab-discover-todo></p-tab-discover-todo>
        </v-tabs-window-item>

        <v-tabs-window-item>
          <p-tab-discover-todo></p-tab-discover-todo>
        </v-tabs-window-item>

        <v-tabs-window-item>
          <p-tab-discover-todo></p-tab-discover-todo>
        </v-tabs-window-item>
      </v-tabs-window>
    </v-tabs>
  </div>
</template>

<script>
import tabColors from "page/discover/colors.vue";
import tabTodo from "page/discover/todo.vue";

export default {
  name: "PPageDiscover",
  components: {
    "p-tab-discover-colors": tabColors,
    "p-tab-discover-todo": tabTodo,
  },
  props: {
    tab: Number,
  },
  data() {
    return {
      readonly: this.$config.get("readonly"),
      active: this.tab,
    };
  },
  mounted() {
    this.$view.enter(this);
  },
  unmounted() {
    this.$view.leave(this);
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
