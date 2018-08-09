import Abstract from 'model/abstract';

class Photo extends Abstract {
    getEntityName() {
        return this.Title;
    }

    getId() {
        return this.ID;
    }

    getGoogleMapsLink() {
        return 'https://www.google.com/maps/place/' + this.Lat + ',' + this.Long;
    }

    static getCollectionResource() {
        return 'photos';
    }

    static getModelName() {
        return 'Photo';
    }
}

export default Photo;
