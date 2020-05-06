export default class Util {
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

    static truncate (str, length, ending) {
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
