<template>
    <div class="p-pull-refresh">
        <div class="pull-down-header" v-bind:style="{'height': pullDown.height + 'px'}">
            <div class="pull-down-content" :style="pullDownContentStyle">
                <i class="pull-down-icon material-icons">{{iconClass}}</i>
            </div>
        </div>
        <slot></slot>
    </div>
</template>

<script>
    const STATUS_ERROR = -1;
    const STATUS_START = 0;
    const STATUS_READY = 1;
    const STATUS_REFRESH = 2;
    const ANIMATION = 'height .2s ease';

    export default {
        props: {
            onRefresh: {
                type: Function
            },
        },
        data() {
            return {
                config: {
                    errorLabel: "Failed to refresh",
                    startLabel: "Pull to refresh",
                    readyLabel: "Release to refresh",
                    loadingLabel: "Updating...",
                    pullDownHeight: 60,
                },
                pullDown: {
                    status: 0,
                    height: 0,
                    msg: ''
                },
                canPull: false
            };
        },
        computed: {
            label() {
                // label of pull down
                if (this.pullDown.status === STATUS_ERROR) {
                    return this.pullDown.msg;
                }
                return this.customLabels[this.pullDown.status + 1];
            },
            customLabels() {
                return [this.config.errorLabel, this.config.startLabel, this.config.readyLabel, this.config.loadingLabel];
            },
            iconClass() {
                // icon of pull down
                if (this.pullDown.status === STATUS_REFRESH) {
                    return 'refresh';
                } else if (this.pullDown.status === STATUS_ERROR) {
                    return 'error';
                }
                return 'refresh';
            },
            pullDownContentStyle() {
                return {
                    bottom: (this.config.pullDownHeight - 40) / 2 + 'px'
                };
            }
        },
        mounted() {
            this.$nextTick(() => {
                let el = this.$el;
                let pullDownHeader = el.querySelector('.pull-down-header');
                let icon = pullDownHeader.querySelector('.pull-down-icon');
                // set default pullDownHeight
                this.config.pullDownHeight = this.config.pullDownHeight || 60;
                /**
                 * reset the status of pull down
                 * @param {Object} pullDown the pull down
                 * @param {Boolean} withAnimation whether add animation when pull up
                 */
                let resetPullDown = (pullDown, withAnimation) => {
                    if (withAnimation) {
                        pullDownHeader.style.transition = ANIMATION;
                    }
                    pullDown.height = 0;
                    pullDown.status = STATUS_START;
                };
                // store of touch position, include start position and distance
                let touchPosition = {
                    start: 0,
                    distance: 0
                };

                // @see https://www.chromestatus.com/feature/5745543795965952
                // Test via a getter in the options object to see if the passive property is accessed
                let supportsPassive = false;
                try {
                    let opts = Object.defineProperty({}, 'passive', {
                        get: function() {
                            supportsPassive = true;
                        }
                    });
                    /* global window */
                    window.addEventListener("test", null, opts);
                } catch (e) {}

                // bind touchstart event to store start position of touch
                el.addEventListener('touchstart', e => {
                    if (el.scrollTop === 0) {
                        this.canPull = true;
                    } else {
                        this.canPull = false;
                    }
                    touchPosition.start = e.touches.item(0).pageY;
                }, supportsPassive ? {passive: true} : false);

                /**
                 * bind touchmove event, do the following:
                 * first, update the height of pull down
                 * finally, update the status of pull down based on the distance
                 */
                el.addEventListener('touchmove', e => {
                    if (!this.canPull) {
                        return;
                    }

                    let distance = e.touches.item(0).pageY - touchPosition.start;
                    // limit the height of pull down to 180
                    distance = distance > 180 ? 180 : distance;
                    // prevent native scroll
                    if (distance > 0) {
                        el.style.overflow = 'hidden';
                    }
                    // update touchPosition and the height of pull down
                    touchPosition.distance = distance;
                    this.pullDown.height = distance;
                    /**
                     * if distance is bigger than the height of pull down
                     * set the status of pull down to STATUS_READY
                     */
                    if (distance > this.config.pullDownHeight) {
                        this.pullDown.status = STATUS_READY;
                        icon.style.transform = 'rotate(180deg)';
                    } else {
                        /**
                         * else set the status of pull down to STATUS_START
                         * and rotate the icon based on distance
                         */
                        this.pullDown.status = STATUS_START;
                        icon.style.transform = 'rotate(' + distance / this.config.pullDownHeight * 180 + 'deg)';
                    }
                }, supportsPassive ? {passive: true} : false);

                // bind touchend event
                el.addEventListener('touchend', e => {
                    this.canPull = false;
                    el.style.overflowY = 'auto';
                    pullDownHeader.style.transition = ANIMATION;
                    // reset icon rotate
                    icon.style.transform = '';
                    // if distance is bigger than 60
                    if (touchPosition.distance - el.scrollTop > this.config.pullDownHeight) {
                        el.scrollTop = 0;
                        this.pullDown.height = this.config.pullDownHeight;
                        this.pullDown.status = STATUS_REFRESH;
                        // trigger refresh callback
                        if (this.onRefresh && typeof this.onRefresh === 'function') {
                            let res = this.onRefresh();
                            // if onRefresh return promise
                            if (res && res.then && typeof res.then === 'function') {
                                res.then(result => {
                                    resetPullDown(this.pullDown, true);
                                }, error => {
                                    // show error and hide the pull down after 1 second
                                    if (typeof error !== 'string') {
                                        error = false;
                                    }
                                    this.pullDown.msg = error || this.customLabels[0];
                                    this.pullDown.status = STATUS_ERROR;
                                    setTimeout(() => {
                                        resetPullDown(this.pullDown, true);
                                    }, 1000);
                                });
                            } else {
                                resetPullDown(this.pullDown);
                            }
                        } else {
                            resetPullDown(this.pullDown);
                            console.warn('please use :on-refresh to pass onRefresh callback');
                        }
                    } else {
                        resetPullDown(this.pullDown);
                    }
                    // reset touchPosition
                    touchPosition.distance = 0;
                    touchPosition.start = 0;
                });
                // remove transition when transitionend
                pullDownHeader.addEventListener('transitionend', () => {
                    pullDownHeader.style.transition = '';
                });
                pullDownHeader.addEventListener('webkitTransitionEnd', () => {
                    pullDownHeader.style.transition = '';
                });
            });
        }
    };
</script>
