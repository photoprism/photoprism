// TODO: find a better name for this mixin

export default {
    props: {
        selection: Array,
        album: Object,
        filter: Object,
    },
    data() {
        return {
            activeIndex: -1,
        }
    },
    mounted() {
        this.$shortcuts.activate(this, 'photo')
        this.$shortcuts.app.$el.focus()

        if (this.photos.length > 0) {
            this.setActiveIndex(this.$shortcuts.activeIndex)
        }
    },
    unmounted() {
        this.$shortcuts.deactivate('photo')
    },
    computed: {
        activePhoto: {
            get: function() {
                return this.photos[this.activeIndex]
            }
        }
    },
    methods: {
        setActiveIndex(index) {
            if (index < 0) {
                index = 0
            }
            this.activeIndex = index
            this.$shortcuts.activeIndex = this.activeIndex
            this.updateActiveElement()
        },

        updateActiveIndex(index) {
            this.setActiveIndex(this.activeIndex + index);
        },

        updateActiveElement() {
            this.$nextTick(() => {
                let activeElement = this.$el.querySelector('.active')
                console.log("activeElement", activeElement)
                if (!this.isElementInView(activeElement, true))
                    activeElement.scrollIntoView(false);
            })
        },

        next() {
            console.log("next!")
            if (this.activeIndex < (this.photos.length-1))
            {
                this.updateActiveIndex(1);
            }
        },
        prev() {
            if (this.activeIndex > 0) {
                this.updateActiveIndex(-1);
            }
        },

        up() {
            const numPerRow = this.getNumPerRow('.p-results')
            if (this.activeIndex >= numPerRow) {
                this.updateActiveIndex(-numPerRow)
            }
        },
        down() {
            const numPerRow = this.getNumPerRow('.p-results')
            if (this.activeIndex < this.photos.length - numPerRow - 1) {
                this.updateActiveIndex(numPerRow)
            }
        },
        open() {
            if (this.activeIndex < 0) return;

            this.openPhoto(this.activeIndex, false);
        },
        edit() {
            if (this.activeIndex < 0) return;

            this.editPhoto(this.activeIndex);
        },
        toggleSelection() {
            if (this.activeIndex < 0) return;

            this.$clipboard.toggle(this.activePhoto);
        },

        // https://stackoverflow.com/questions/49043684/how-to-calculate-the-amount-of-flexbox-items-in-a-row/49046973
        getNumPerRow(gridSelector) {
            console.log("gridSelector", gridSelector)
            window.$elem = this.$el
            console.log("gridSelector", gridSelector)
            const grid = this.$el.querySelector(gridSelector);
            console.log("grid", grid)
            const gridChildren = Array.from(grid.children);
            const gridNum = gridChildren.length;
            const baseOffset = gridChildren[0].offsetTop;
            const breakIndex = gridChildren.findIndex(item => item.offsetTop > baseOffset);
            const numPerRow = (breakIndex === -1 ? gridNum : breakIndex);
            return numPerRow
        },

        // https://stackoverflow.com/questions/487073/how-to-check-if-element-is-visible-after-scrolling
        isElementInView: function (element, fullyInView) {
            //var pageTop = window.scrollY;
            //var pageBottom = pageTop + window.innerHeight;
            var pageTop = 0
            var pageBottom = pageTop + window.innerHeight;
            var rect = element.getBoundingClientRect()
            var elementTop = rect.top
            var elementBottom = rect.bottom

            console.log("pageTop", pageTop, "pageBottom", pageBottom)
            console.log("elementTop", elementTop, "elementBottom", elementBottom)
    
            if (fullyInView === true) {
                return ((pageTop < elementTop) && (pageBottom > elementBottom));
            } else {
                return ((elementTop <= pageBottom) && (elementBottom >= pageTop));
            }
        }


    }

}