import assert from 'assert';
import Api from 'common/api';
import MockAdapter from 'axios-mock-adapter';

const mock = new MockAdapter(Api);

const getCollectionResponse = [
    {id: 1, name: 'John Smith'},
    {id: 1, name: 'John Smith'}
];

const getEntityResponse = {
    id: 1, name: 'John Smith'
};

const postEntityResponse = {
    users: [
        {id: 1, name: 'John Smith'}
    ]
};

const putEntityResponse = {
    users: [
        {id: 2, name: 'John Foo'}
    ]
};

const deleteEntityResponse = null;

mock.onGet('foo').reply(200, getCollectionResponse);
mock.onGet('foo/123').reply(200, getEntityResponse);
mock.onPost('foo').reply(201, postEntityResponse);
mock.onPut('foo/2').reply(200, putEntityResponse);
mock.onDelete('foo/2').reply(204, deleteEntityResponse);

describe('common/api', () => {
    it('get("foo") should return a list of results and return with HTTP code 200', (done) => {
        Api.get('foo').then(
            (response) => {
                assert.equal(200, response.status);
                assert.deepEqual(getCollectionResponse, response.data);
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it('get("foo/123") should return one item and return with HTTP code 200', (done) => {
        Api.get('foo/123').then(
            (response) => {
                assert.equal(200, response.status);
                assert.deepEqual(getEntityResponse, response.data);
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it('post("foo") should create one item and return with HTTP code 201', (done) => {
        Api.post('foo', postEntityResponse).then(
            (response) => {
                assert.equal(201, response.status);
                assert.deepEqual(postEntityResponse, response.data);
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it('put("foo/2") should update one item and return with HTTP code 200', (done) => {
        Api.put('foo/2', putEntityResponse).then(
            (response) => {
                assert.equal(200, response.status);
                assert.deepEqual(putEntityResponse, response.data);
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });

    it('delete("foo/2") should delete one item and return with HTTP code 204', (done) => {
        Api.delete('foo/2', deleteEntityResponse).then(
            (response) => {
                assert.equal(204, response.status);
                assert.deepEqual(deleteEntityResponse, response.data);
                done();
            }
        ).catch(
            (error) => {
                done(error);
            }
        );
    });
});