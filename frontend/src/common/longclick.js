/*
    VueJS long click directive

    Based on https://github.com/ittus/vue-long-click (MIT License)
    Author: Thang Minh Vu (https://github.com/ittus)
*/
export default ({delay = 400, interval = 50}) => ({
    bind: function (el, binding, vNode) {
        if (typeof binding.value !== "function") {
            const compName = vNode.context.name;
            let warn = `[longclick:] provided expression '${binding.expression}' is not a function, but has to be`;
            if (compName) {
                warn += `Found in component '${compName}' `;
            }
            console.warn(warn) // eslint-disable-line
        }

        let pressTimer = null;
        let pressInterval = null;

        const start = (e) => {
            if (e.type === "click" && e.button !== 0) {
                return;
            }

            if (pressTimer === null) {
                pressTimer = setTimeout(() => {
                    if (interval && interval > 0) {
                        pressInterval = setInterval(() => {
                            handler();
                        }, interval);
                    }
                    handler();
                }, delay);
            }
        };

        // Cancel Timeout
        const cancel = () => {
            if (pressTimer !== null) {
                clearTimeout(pressTimer);
                pressTimer = null;
            }
            if (pressInterval) {
                clearInterval(pressInterval);
                pressInterval = null;
            }
        };
        // Run Function
        const handler = (e) => {
            binding.value(e);
        }

        ;["mousedown", "touchstart"].forEach(e => el.addEventListener(e, start))
        ;["click", "mouseout", "touchend", "touchcancel"].forEach(e => el.addEventListener(e, cancel));
    },
});
