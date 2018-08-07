import Abstract from 'model/abstract';

class Photo extends Abstract {
    getEntityName() {
        return this.Title;
    }

    getId() {
        return this.ID;
    }

    static getCollectionResource() {
        return 'photos';
    }

    static getModelName() {
        return 'Photo';
    }
}

export default Photo;
