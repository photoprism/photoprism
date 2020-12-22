export default {
  props: {
    staticFilter: Object,
    view: String,
  },
    
  watch: {
    '$route'() {
      const query = this.$route.query;

      this.filter.q = query['q'] ? query['q'] : '';
      this.filter.all = query['all'] ? query['all'] : '';
      this.lastFilter = {};
      this.routeName = this.$route.name;
      this.search();
    }
  },

  data() {
      const routeName = this.$route.name;
      const settings = {};

      return {
        subscriptions: [],
        listen: false,
        dirty: false,
        results: [],
        scrollDisabled: true,
        loading: true,
        pageSize: 24,
        offset: 0,
        page: 0,
        selection: [],
        settings: settings,
        lastFilter: {},
        routeName: routeName,
        titleRule: v => v.length <= this.$config.get('clip') || this.$gettext("Name too long"),
        mouseDown: {
          index: -1,
          timeStamp: -1,
        },
        lastId: "",
      }
  }

}