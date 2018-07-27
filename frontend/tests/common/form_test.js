import assert from 'assert';
import Form from 'common/form';

describe('common/form', () => {
    it('setDefinition', () => {
        let def = {'foo': {'type': 'string', 'caption': 'Foo'}};
        const form = new Form();

        form.setDefinition(def);

        let result = form.getDefinition();

        assert.equal(result, def);
    });
});