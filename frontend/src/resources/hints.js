import {$gettext} from "common/vm";

export default {
    private: $gettext("Exclude photos marked as private from search results, shared albums, labels and places."),
    review: $gettext("Non-photographic and low-quality images require a review before they appear in search results."),
    group: $gettext("Files with sequential names like 'IMG_1234 (2)' or 'IMG_1234 copy 2' belong to the same photo."),
    move: $gettext("Move files from import to originals to save storage. Unsupported file types will never be deleted, they remain in their current location."),
    places: $gettext("Search and display photos on a map."),
    originals: $gettext("Display indexed files in Originals"),
    moments: $gettext("Let PhotoPrism create albums from past events."),
    labels: $gettext("Browse and edit image classification labels."),
    import: $gettext("Imported files will be sorted by date and given a unique name."),
    archive: $gettext("Hide photos that have been moved to archive."),
    upload: $gettext("Add files to your library via Web Upload."),
    download: $gettext("Download single files and zip archives."),
    edit: $gettext("Change photo titles, locations and other metadata."),
    share: $gettext("Upload to WebDAV and other remote services."),
    logs: $gettext("Show server logs in Library."),
    library: $gettext("Show Library in navigation menu."),
    convert: $gettext("File types like RAW might need to be converted so that they can be displayed in a browser. JPEGs will be stored in the same folder next to the original using the best possible quality."),
}
