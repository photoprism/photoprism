import {$gettext} from "common/vm";
import moment from "moment-timezone";

export const TimeZones = () => moment.tz.names();

export const Languages = () => [
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

export const Themes = () => [
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
export const MapsAnimate = () => [
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

export const MapsStyle = () => [
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

export const Intervals = () => [
    {"value": 0, "text": $gettext("Never")},
    {"value": 3600, "text": $gettext("1 hour")},
    {"value": 3600 * 4, "text": $gettext("4 hours")},
    {"value": 3600 * 12, "text": $gettext("12 hours")},
    {"value": 86400, "text": $gettext("Daily")},
    {"value": 86400 * 2, "text": $gettext("Every two days")},
    {"value": 86400 * 7, "text": $gettext("Once a week")},
];

export const Expires = () => [
    {"value": 0, "text": $gettext("Never")},
    {"value": 86400, "text": $gettext("After 1 day")},
    {"value": 86400 * 3, "text": $gettext("After 3 days")},
    {"value": 86400 * 7, "text": $gettext("After 7 days")},
    {"value": 86400 * 14, "text": $gettext("After two weeks")},
    {"value": 86400 * 31, "text": $gettext("After one month")},
    {"value": 86400 * 60, "text": $gettext("After two months")},
    {"value": 86400 * 365, "text": $gettext("After one year")},
];
