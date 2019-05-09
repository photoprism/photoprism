export const FormPropertyType = Object.freeze({
    String: "string",
    Number: "number",
    Object: "object",
});

export default class Form {
    constructor(definition) {
        this.definition = definition;
    }

    setValues(values) {
        const def = this.getDefinition();

        for (let prop in def) {
            if (values.hasOwnProperty(prop)) {
                this.setValue(prop, values[prop]);
            }
        }

        return this;
    }

    getValues() {
        const result = {};
        const def = this.getDefinition();

        for (let prop in def) {
            result[prop] = this.getValue(prop);
        }

        return result;
    }

    setValue(name, value) {
        const def = this.getDefinition();

        if (!def.hasOwnProperty(name)) {
            throw `Property ${name} not found`;
        } else if (typeof value != def[name].type) {
            throw `Property ${name} must be ${def[name].type}`;
        } else {
            def[name].value = value;
        }

        return this;
    }

    getValue(name) {
        const def = this.getDefinition();

        if (def.hasOwnProperty(name)) {
            return def[name].value;
        } else {
            throw `Property ${name} not found`;
        }
    }

    setDefinition(definition) {
        this.definition = definition;
    }

    getDefinition() {
        return this.definition ? this.definition : {};
    }

    getOptions(fieldName) {
        if (
            this.definition &&
            this.definition.hasOwnProperty(fieldName) &&
            this.definition[fieldName].hasOwnProperty("options")
        ) {
            return this.definition[fieldName].options;
        }

        return [{ option: "", label: "" }];
    }
}
