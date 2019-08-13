import Form, { FormPropertyType } from 'common/form';

let chai = require('../../../node_modules/chai/chai');
let assert = chai.assert;

describe('common/form', () => {
    it('setting and getting definition', () => {
        const def = { foo: { type: FormPropertyType.String, caption: 'Foo' } };
        const form = new Form();

        form.setDefinition(def);

        const result = form.getDefinition();
        assert.equal(result, def);
    });

    it('setting and getting a value according to type', () => {
        const def = {
            foo: { type: FormPropertyType.String, caption: 'Foo' },
        };
        const form = new Form();

        form.setDefinition(def);
        form.setValue('foo', 'test');

        const result = form.getValue('foo');
        assert.equal(result, 'test');
    });

    it('setting a value not according to type', () => {
        const def = {
            foo: { type: FormPropertyType.String, caption: 'Foo' },
        };
        const form = new Form();

        form.setDefinition(def);

        assert.throws(() => {
            form.setValue('foo', 3);
        });
    });

    it('setting and getting a value for missing property throws exception', () => {
        const def = {
            foo: { type: FormPropertyType.String, caption: 'Foo' },
        };
        const form = new Form();

        form.setDefinition(def);

        assert.throws(() => {
            form.setValue('bar', 3);
        });

        assert.throws(() => {
            form.getValue('bar');
        });
    });

    it('setting and getting a complex value', () => {
        const complexValue = {
            something: 'abc',
            another: 'def',
        };
        const def = {
            foo: {
                type: FormPropertyType.Object,
                caption: 'Foo',
            },
        };
        const form = new Form();

        form.setDefinition(def);
        form.setValue('foo', complexValue);

        const result = form.getValue('foo');
        assert.deepEqual(result, complexValue);
    });

    it('setting and getting more values at once', () => {
        const def = {
            foo: { type: FormPropertyType.String, caption: 'Foo' },
            baz: { type: FormPropertyType.String, caption: 'XX' },
        };
        const form = new Form();

        form.setDefinition(def);
        form.setValues({ foo: 'test', baz: 'yyy'});

        const result = form.getValues();
        assert.equal(result.foo, 'test');
        assert.equal(result.baz, 'yyy');
    });

    it('getting options of fieldname', () => {
        const def = {
            search: {
                type: FormPropertyType.String,
                caption: 'Search',
                label: {options: 'tiles', text: 'Tiles'},
                options: [
                    {value: 'tiles', text: 'Tiles'},
                    {value: 'mosaic', text: 'Mosaic'},
                ],
            },
        };
        const form = new Form();

        form.setDefinition(def);

        const result = form.getOptions("search");
        assert.equal(result[0].value, "tiles");
        assert.equal(result[1].text, "Mosaic");
    });

    it('getting not existing options returns empty object', () => {
        const def = {
            foo: {
                type: FormPropertyType.Object,
                caption: 'Foo',
            },
        };
        const form = new Form();

        form.setDefinition(def);

        const result = form.getOptions("foo");
        assert.equal(result[0].option, "");
        assert.equal(result[0].label, "");
    });
});
