import {$gettext} from "common/vm";
import moment from "moment-timezone";
import {Info} from "luxon";
import {config} from "../session";

export const TimeZones = () => moment.tz.names();

export const Days = () => {
    let result = [];

    for (let i = 1; i <= 31; i++) {
        result.push({"value": i, "text": i.toString().padStart(2, "0")});
    }

    result.push({"value": -1, "text": $gettext("Unknown")});

    return result;
};

export const Years = () => {
    let result = [];

    const currentYear = new Date().getUTCFullYear();

    for (let i = currentYear; i >= 1750; i--) {
        result.push({"value": i, "text": i.toString().padStart(4, "0")});
    }

    result.push({"value": -1, "text": $gettext("Unknown")});

    return result;
};

export const IndexedYears = () => {
    let result = [];

    if (config.values.years) {
        for (let i = 0; i < config.values.years.length; i++) {
            result.push({"value": parseInt(config.values.years[i]), "text": config.values.years[i].toString()});
        }
    }

    result.push({"value": -1, "text": $gettext("Unknown")});

    return result;
};

export const Months = () => {
    let result = [];

    const months = Info.months("long");

    for (let i = 0; i < months.length; i++) {
        result.push({"value": i + 1, "text": months[i]});
    }

    result.push({"value": -1, "text": $gettext("Unknown")});

    return result;
};

export const MonthsShort = () => {
    let result = [];

    for (let i = 1; i <= 12; i++) {
        result.push({"value": i + 1, "text": i.toString().padStart(2, "0")});
    }

    result.push({"value": -1, "text": $gettext("Unknown")});

    return result;
};

export const Languages = () => [
    {
        "text": $gettext("English"),
        "value": "en",
    },
    {
        "text": $gettext("German"),
        "value": "de",
    },
    {
        "text": $gettext("French"),
        "value": "fr",
    },
    {
        "text": $gettext("Spanish"),
        "value": "es",
    },
    {
        "text": $gettext("Dutch"),
        "value": "nl",
    },
    {
        "text": $gettext("Polish"),
        "value": "pl",
    },
    {
        "text": $gettext("Russian"),
        "value": "ru",
    },
];

export const Themes = () => [
    {
        "text": $gettext("Default"),
        "value": "default",
    },
    {
        "text": $gettext("Cyano"),
        "value": "cyano",
    },
    {
        "text": $gettext("Lavender"),
        "value": "lavender",
    },
    {
        "text": $gettext("Moonlight"),
        "value": "moonlight",
    },
    {
        "text": $gettext("Onyx"),
        "value": "onyx",
    },
    {
        "text": $gettext("Raspberry"),
        "value": "raspberry",
    },
    {
        "text": $gettext("Seaweed"),
        "value": "seaweed",
    },
];
export const MapsAnimate = () => [
    {
        "text": $gettext("None"),
        "value": 0,
    },
    {
        "text": $gettext("Fast"),
        "value": 2500,
    },
    {
        "text": $gettext("Medium"),
        "value": 6250,
    },
    {
        "text": $gettext("Slow"),
        "value": 10000,
    },
];

export const MapsStyle = () => [
    {
        "text": $gettext("Offline"),
        "value": "offline",
    },
    {
        "text": $gettext("Streets"),
        "value": "streets",
    },
    {
        "text": $gettext("Hybrid"),
        "value": "hybrid",
    },
    {
        "text": $gettext("Topographic"),
        "value": "topo",
    },
    {
        "text": $gettext("Moonlight"),
        "value": "darkmatter",
    },
];

export const PhotoTypes = () => [
    {
        "text": $gettext("Image"),
        "value": "image",
    },
    {
        "text": $gettext("Raw"),
        "value": "raw",
    },
    {
        "text": $gettext("Live"),
        "value": "live",
    },
    {
        "text": $gettext("Video"),
        "value": "video",
    },
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

export const Colors = () => [
    {"Example": "#AB47BC", "Name": $gettext("Purple"), "Slug": "purple"},
    {"Example": "#FF00FF", "Name": $gettext("Magenta"), "Slug": "magenta"},
    {"Example": "#EC407A", "Name": $gettext("Pink"), "Slug": "pink"},
    {"Example": "#EF5350", "Name": $gettext("Red"), "Slug": "red"},
    {"Example": "#FFA726", "Name": $gettext("Orange"), "Slug": "orange"},
    {"Example": "#D4AF37", "Name": $gettext("Gold"), "Slug": "gold"},
    {"Example": "#FDD835", "Name": $gettext("Yellow"), "Slug": "yellow"},
    {"Example": "#CDDC39", "Name": $gettext("Lime"), "Slug": "lime"},
    {"Example": "#66BB6A", "Name": $gettext("Green"), "Slug": "green"},
    {"Example": "#009688", "Name": $gettext("Teal"), "Slug": "teal"},
    {"Example": "#00BCD4", "Name": $gettext("Cyan"), "Slug": "cyan"},
    {"Example": "#2196F3", "Name": $gettext("Blue"), "Slug": "blue"},
    {"Example": "#A1887F", "Name": $gettext("Brown"), "Slug": "brown"},
    {"Example": "#F5F5F5", "Name": $gettext("White"), "Slug": "white"},
    {"Example": "#9E9E9E", "Name": $gettext("Grey"), "Slug": "grey"},
    {"Example": "#212121", "Name": $gettext("Black"), "Slug": "black"},
];
