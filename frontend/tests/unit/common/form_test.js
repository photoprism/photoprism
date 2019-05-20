import assert from 'assert';
import Form, { FormPropertyType } from 'common/form';

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
        };
        const form = new Form();

        form.setDefinition(def);
        form.setValues({ foo: 'test' });

        const result = form.getValues();
        assert.equal(result.foo, 'test');
    });
});
