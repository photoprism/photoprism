export default {
  props: {
    photos: Array,
    openPhoto: Function,
    editPhoto: Function,
    openLocation: Function,
    context: String,
  },
  data() {
    return {
      clipboard: this.$clipboard,
      showLocation: this.$config.settings().features.places,
      hidePrivate: this.$config.settings().features.private,
      debug: this.$config.get('debug'),
      labels: {
        approve: this.$gettext("Approve"),
        archive: this.$gettext("Archive"),
      },
      mouseDown: {
        index: -1,
        timeStamp: -1,
      },
    };
  },
  methods: {
    onSelect(ev, index) {
      if (ev.shiftKey) {
        this.selectRange(index);
      } else {
        this.$clipboard.toggle(this.photos[index]);
      }

      this.activeIndex = index;
    },

    onMouseDown(ev, index) {
      this.mouseDown.index = index;
      this.mouseDown.timeStamp = ev.timeStamp;
    },

    onClick(ev, index) {
      this.activeIndex = index

      console.log("my click", index, 'activeIndex', this.activeIndex);
      let longClick = (this.mouseDown.index === index && ev.timeStamp - this.mouseDown.timeStamp > 400);

      if (longClick || this.selection.length > 0) {
        if (longClick || ev.shiftKey) {
          this.selectRange(index);
        } else {
          this.$clipboard.toggle(this.photos[index]);
        }
      } else {
        let photo = this.photos[index];

        if (photo.Type === 'video' && photo.isPlayable()) {
          this.openPhoto(index, true);
        } else {
          this.openPhoto(index, false);
        }
      }

    },
    onContextMenu(ev, index) {
      if (this.$isMobile) {
        ev.preventDefault();
        ev.stopPropagation();
        this.selectRange(index);
      }
    },
    selectRange(index, ) {
      this.$clipboard.addRange(index, this.photos);
    }
  }
}