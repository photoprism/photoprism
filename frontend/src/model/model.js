class Model {
    constructor(values) {
        this.__originalValues = {};

        if (values) {
            this.setValues(values);
        } else {
            this.setValues(this.getDefaults());
        }
    }

    setValues(values, scalarOnly) {
        if (!values) return;

        for (let key in values) {
            if (values.hasOwnProperty(key) && key !== "__originalValues") {
                this[key] = values[key];

                if (typeof values[key] !== "object") {
                    this.__originalValues[key] = values[key];
                } else if (!scalarOnly) {
                    this.__originalValues[key] = JSON.parse(JSON.stringify(values[key]));
                }

            }
        }

        return this;
    }

    getValues(changed) {
        const result = {};
        const defaults = this.getDefaults();

        for (let key in this.__originalValues) {
            if (this.__originalValues.hasOwnProperty(key) && key !== "__originalValues") {
                let val;
                if (defaults.hasOwnProperty(key)) {
                    switch (typeof defaults[key]) {
                    case "bigint":
                    case "number":
                        val = parseFloat(this[key]);
                        break;
                    case "boolean":
                        val = !!this[key];
                        break;
                    default:
                        val = this[key];
                    }
                } else {
                    val = this[key];
                }

                if (!changed || JSON.stringify(val) !== JSON.stringify(this.__originalValues[key])) {
                    result[key] = val;
                }
            }
        }

        return result;
    }

    getDefaults() {
        return {};
    }
}

export default Model;
