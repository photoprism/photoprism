const Nanosecond = 1;
const Microsecond = 1000 * Nanosecond;
const Millisecond = 1000 * Microsecond;
const Second = 1000 * Millisecond;
const Minute = 60 * Second;
const Hour = 60 * Minute;

export default class Util {
    static duration(d) {
        let u = d;

        let neg = d < 0;

        if (neg) {
            u = -u;
        }

        if (u < Second) {
            // Special case: if duration is smaller than a second,
            // use smaller units, like 1.2ms
            if (!u) {
                return "0s";
            }

            if (u < Microsecond) {
                return u + "ns";
            }

            if (u < Millisecond) {
                return Math.round(u / Microsecond) + "µs";
            }

            return Math.round(u / Millisecond) + "ms";
        }

        let result = [];

        let h = Math.floor(u / Hour);
        let min = Math.floor(u / Minute)%60;
        let sec = Math.ceil(u / Second)%60;

        result.push(h.toString().padStart(2, "0"));
        result.push(min.toString().padStart(2, "0"));
        result.push(sec.toString().padStart(2, "0"));

        // return `${h}h${min}m${sec}s`

        return result.join(":");
    }

    static arabicToRoman(number) {
        let roman = "";
        const romanNumList = {
            M: 1000,
            CM: 900,
            D: 500,
            CD: 400,
            C: 100,
            XC: 90,
            L: 50,
            XV: 40,
            X: 10,
            IX: 9,
            V: 5,
            IV: 4,
            I: 1,
        };
        let a;
        if (number < 1 || number > 3999)
            return "";
        else {
            for (let key in romanNumList) {
                a = Math.floor(number / romanNumList[key]);
                if (a >= 0) {
                    for (let i = 0; i < a; i++) {
                        roman += key;
                    }
                }
                number = number % romanNumList[key];
            }
        }

        return roman;
    }

    static truncate(str, length, ending) {
        if (length == null) {
            length = 100;
        }
        if (ending == null) {
            ending = "...";
        }
        if (str.length > length) {
            return str.substring(0, length - ending.length) + ending;
        } else {
            return str;
        }
    }
}
