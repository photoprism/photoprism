class Form {
    constructor(definition) {
        this.definition = definition;
    }

    setValues(values) {
        const def = this.getDefinition();

        for(let prop in def) {
            if(def.hasOwnProperty(prop) && values.hasOwnProperty(prop)) {
                def[prop].value = values[prop];
            }
        }

        return this;
    }

    getValues() {
        const result = {};
        const def = this.getDefinition();
        
        for(let prop in def) {
            if(def.hasOwnProperty(prop)) {
                result[prop] = def[prop].value;
            }
        }
        
        return result;
    }
    
    setDefinition(definition) {
        this.definition = definition;
    }

    getDefinition() {
        return this.definition ? this.definition : {};
    }

    getOptions(fieldName) {
        if(this.definition && this.definition.hasOwnProperty(fieldName) && this.definition[fieldName].hasOwnProperty('options')) {
            return this.definition[fieldName].options;
        }

        return [{option: '', label: ''}];
    }
}

export default Form;