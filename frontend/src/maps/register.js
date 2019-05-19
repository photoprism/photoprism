import {LMap, LTileLayer, LMarker, LControl} from "vue2-leaflet";
import {Icon} from "leaflet";

const components = {};

components.install = (Vue) => {
    Vue.component("l-map", LMap);
    Vue.component("l-tile-layer", LTileLayer);
    Vue.component("l-marker", LMarker);
    Vue.component("l-control", LControl);

    delete Icon.Default.prototype._getIconUrl;

    Icon.Default.mergeOptions({
        iconRetinaUrl: require("./marker/marker-icon-2x-red.png"),
        iconUrl: require("./marker/marker-icon-red.png"),
        shadowUrl: require("./marker/marker-shadow.png"),
    });
};

export default components;
