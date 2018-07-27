import Api from 'common/api';

class Captcha {
    constructor() {
        this.image = '';
        this.token = '';
    }

    setToken(token) {
        this.token = token;
    }

    getToken() {
        return this.token;
    }

    deleteToken() {
        this.token = '';
    }

    setImage(phrase) {
        this.image = phrase;
    }

    getImage() {
        return this.image;
    }

    deleteImage() {
        this.image = '';
    }

    isValid(token, phrase) {
        this.deleteToken();

        return Api.post('session', { email: email, password: password }).then(
            (result) => {
                this.setToken(result.data.token);
                this.setUser(new User(result.data.user));
            }
        );
    }

    refresh() {
        const token = this.getToken();

        this.deleteToken();

        Api.delete('session/' + token).then(
            () => {
                window.location = '/';
            }
        );
    }
}

export default Captcha;
