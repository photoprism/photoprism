import Api from 'common/api';
import User from 'model/user';

class Session {
    /**
     * @param {Storage} storage
     */
    constructor(storage) {
        this.storage = storage;
        this.session_token = this.storage.getItem('session_token');

        const userJson = this.storage.getItem('user');

        this.user = userJson !== 'undefined' ? new User(JSON.parse(userJson)) : null;
    }

    setToken(token) {
        this.session_token = token;
        this.storage.setItem('session_token', token);
        Api.defaults.headers.common['X-Session-Token'] = token;
    }

    getToken() {
        return this.session_token;
    }

    deleteToken() {
        this.session_token = null;
        this.storage.removeItem('session_token');
        Api.defaults.headers.common['X-Session-Token'] = '';
        this.deleteUser();
    }

    setUser(user) {
        this.user = user;
        this.storage.setItem('user', JSON.stringify(user.getValues()));
    }

    getUser() {
        return this.user;
    }

    getEmail() {
        if (this.isUser()) {
            return this.user.userEmail;
        }

        return '';
    }

    getFullName() {
        if (this.isUser()) {
            return this.user.userFirstName + ' ' + this.user.userLastName;
        }

        return '';
    }

    getFirstName() {
        if (this.isUser()) {
            return this.user.userFirstName;
        }

        return '';
    }

    isUser() {
        return this.user.hasId();
    }

    isAdmin() {
        return this.user.hasId() && this.user.userRole === 'admin';
    }

    isAnonymous() {
        return !this.user.hasId();
    }

    deleteUser() {
        this.user = null;
        this.storage.removeItem('user');
    }

    login(email, password) {
        this.deleteToken();

        return Api.post('session', { email: email, password: password }).then(
            (result) => {
                this.setToken(result.data.token);
                this.setUser(new User(result.data.user));
           }
        );
    }

    logout() {
        const token = this.getToken();

        this.deleteToken();

        Api.delete('session/' + token).then(
            () => {
                window.location = '/';
            }
        );
    }
}

export default Session;
