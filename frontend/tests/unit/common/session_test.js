import assert from 'assert';
import Session from 'common/session';

describe('common/session', () => {
    it('should construct session',  () => {
        const storage = window.localStorage;
        const session = new Session(storage);
        assert.equal(session.session_token, null);
    });

    it('should set and get token',  () => {
        const storage = window.localStorage;
        const session = new Session(storage);
        assert.equal(session.session_token, null);
        session.setToken(123421);
        assert.equal(session.session_token, 123421);
        const result = session.getToken();
        assert.equal(result, 123421);
    });

    it('should delete token',  () => {
        const storage = window.localStorage;
        const session = new Session(storage);
        assert.equal(session.session_token, null);
        session.setToken(123421);
        assert.equal(session.session_token, 123421);
        session.deleteToken();
        assert.equal(session.session_token, null);
    });
});