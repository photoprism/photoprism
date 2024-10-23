<template>
  <div class="p-page p-page-discover">
    <v-tabs v-model="active" elevation="0" grow class="bg-transparent" background-color="secondary" slider-color="secondary-dark" :height="$vuetify.display.smAndDown ?  48 : 64">
      <v-tab id="tab-discover-colors" ripple @click="changePath('/discover')">
        <translate>Colors</translate>
      </v-tab>

      <v-tab id="tab-discover-similar" ripple @click="changePath('/discover/similar')">
        <translate>Similar</translate>
      </v-tab>

      <v-tab id="tab-discover-season" ripple @click="changePath('/discover/season')">
        <translate>Season</translate>
      </v-tab>

      <v-tab id="tab-discover-random" ripple @click="changePath('/discover/random')">
        <translate>Random</translate>
      </v-tab>

      <v-tabs v-model="active">
        <v-window>
          <v-window-item>
            <p-tab-discover-colors></p-tab-discover-colors>
          </v-window-item>

          <v-window-item>
            <p-tab-discover-todo></p-tab-discover-todo>
          </v-window-item>

          <v-window-item>
            <p-tab-discover-todo></p-tab-discover-todo>
          </v-window-item>

          <v-window-item>
            <p-tab-discover-todo></p-tab-discover-todo>
          </v-window-item>
        </v-window>
      </v-tabs>
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
  methods: {
    changePath: function (path) {
      if (this.$route.path !== path) {
        this.$router.replace(path);
      }
    },
  },
};
</script>
