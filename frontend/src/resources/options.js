import {$gettext} from "common/vm";
import moment from "moment-timezone";

export const TimeZones = moment.tz.names();

export const Languages = [
    {
        "text": $gettext("English"),
        "value": "en"
    },
    {
        "text": $gettext("German"),
        "value": "de"
    },
    {
        "text": $gettext("Dutch"),
        "value": "nl"
    },
    {
        "text": $gettext("Russian"),
        "value": "ru"
    }
];

export const Themes = [
    {
        "text": $gettext("Default"),
        "value": "default"
    },
    {
        "text": $gettext("Cyano"),
        "value": "cyano"
    },
    {
        "text": $gettext("Lavender"),
        "value": "lavender"
    },
    {
        "text": $gettext("Moonlight"),
        "value": "moonlight"
    },
    {
        "text": $gettext("Onyx"),
        "value": "onyx"
    },
    {
        "text": $gettext("Raspberry"),
        "value": "raspberry"
    },
    {
        "text": $gettext("Seaweed"),
        "value": "seaweed"
    }
];
export const MapsAnimate = [
    {
        "text": $gettext("None"),
        "value": 0
    },
    {
        "text": $gettext("Fast"),
        "value": 2500
    },
    {
        "text": $gettext("Medium"),
        "value": 6250
    },
    {
        "text": $gettext("Slow"),
        "value": 10000
    }
];

export const MapsStyle = [
    {
        "text": $gettext("Offline"),
        "value": "offline"
    },
    {
        "text": $gettext("Streets"),
        "value": "streets"
    },
    {
        "text": $gettext("Hybrid"),
        "value": "hybrid"
    },
    {
        "text": $gettext("Topographic"),
        "value": "topo"
    },
    {
        "text": $gettext("Moonlight"),
        "value": "darkmatter"
    }
];
