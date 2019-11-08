import Abstract from "model/abstract";
import Form from "common/form";
import Api from "common/api";

class User extends Abstract {
    getEntityName() {
        return this.FirstName + " " + this.LastName;
    }

    getId() {
        return this.ID;
    }

    getRegisterForm() {
        return Api.options(this.getEntityResource() + "/register").then(response => Promise.resolve(new Form(response.data)));
    }

    getProfileForm() {
        return Api.options(this.getEntityResource() + "/profile").then(response => Promise.resolve(new Form(response.data)));
    }

    changePassword(oldPassword, newPassword) {
        return Api.put(this.getEntityResource() + "/password", {
            password: oldPassword,
            new_password: newPassword,
        }).then((response) => Promise.resolve(response.data));
    }

    saveProfile() {
        return Api.post(this.getEntityResource() + "/profile", this.getValues()).then((response) => Promise.resolve(this.setValues(response.data)));
    }

    static getCollectionResource() {
        return "users";
    }

    static getModelName() {
        return "User";
    }
}

export default User;
