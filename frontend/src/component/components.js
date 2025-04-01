// Global site components.
import PNotify from "component/notify.vue";
import PScroll from "component/scroll.vue";
import PNavigation from "component/navigation.vue";
import PUpdate from "component/update.vue";
import PLoading from "component/loading.vue";
import PLoadingBar from "component/loading-bar.vue";
import PLightboxMenu from "component/lightbox/menu.vue";
import PSidebarMetadata from "component/sidebar/metadata.vue";
import PLightbox from "component/lightbox.vue";

// Icons.
import IconLivePhoto from "component/icon/live-photo.vue";
import IconSponsor from "component/icon/sponsor.vue";
import IconPrism from "component/icon/prism.vue";

// User account management.
import PSettingsApps from "component/settings/apps.vue";
import PSettingsPasscode from "component/settings/passcode.vue";
import PSettingsPassword from "component/settings/password.vue";

// Albums.
import PAlbumClipboard from "component/album/clipboard.vue";
import PAlbumToolbar from "component/album/toolbar.vue";
import PAlbumEditDialog from "component/album/edit/dialog.vue";
import PAlbumDeleteDialog from "component/album/delete/dialog.vue";

// Login.
import PAuthHeader from "component/auth/header.vue";
import PAuthFooter from "component/auth/footer.vue";

// Sharing.
import PShareDialog from "component/share/dialog.vue";

// Settings.
import PSettingsWebdav from "component/settings/webdav.vue";

// Confirm.
import PConfirmDialog from "component/confirm/dialog.vue";
import PConfirmSponsor from "component/confirm/sponsor.vue";

// Originals.
import PFileClipboard from "component/file/clipboard.vue";
import PFileDeleteDialog from "component/file/delete/dialog.vue";

// Labels.
import PLabelClipboard from "component/label/clipboard.vue";
import PLabelDeleteDialog from "component/label/delete/dialog.vue";
import PLabelEditDialog from "component/label/edit/dialog.vue";

// People.
import PPeopleMergeDialog from "component/people/merge/dialog.vue";
import PPeopleEditDialog from "component/people/edit/dialog.vue";
import PPeopleClipboard from "component/people/clipboard.vue";

// Photos.
import PPhotoViewCards from "component/photo/view/cards.vue";
import PPhotoViewMosaic from "component/photo/view/mosaic.vue";
import PPhotoViewList from "component/photo/view/list.vue";
import PPhotoClipboard from "component/photo/clipboard.vue";
import PPhotoToolbar from "component/photo/toolbar.vue";
import PPhotoPreview from "component/photo/preview.vue";
import PPhotoArchiveDialog from "component/photo/archive/dialog.vue";
import PPhotoAlbumDialog from "component/photo/album/dialog.vue";
import PPhotoEditDialog from "component/photo/edit/dialog.vue";
import PPhotoUploadDialog from "component/photo/upload/dialog.vue";

// Services.
import PServiceAdd from "component/service/add.vue";
import PServiceRemove from "component/service/remove.vue";
import PServiceEdit from "component/service/edit.vue";
import PServiceUpload from "component/service/upload.vue";

// Installs the components imported above.
export function install(app) {
  app.component("PNotify", PNotify);
  app.component("PScroll", PScroll);
  app.component("PNavigation", PNavigation);
  app.component("PLoading", PLoading);
  app.component("PLoadingBar", PLoadingBar);
  app.component("PLightboxMenu", PLightboxMenu);
  app.component("PSidebarMetadata", PSidebarMetadata);
  app.component("PLightbox", PLightbox);
  app.component("PUpdate", PUpdate);

  app.component("IconLivePhoto", IconLivePhoto);
  app.component("IconSponsor", IconSponsor);
  app.component("IconPrism", IconPrism);

  app.component("PSettingsApps", PSettingsApps);
  app.component("PSettingsPasscode", PSettingsPasscode);
  app.component("PSettingsPassword", PSettingsPassword);
  app.component("PSettingsWebdav", PSettingsWebdav);

  app.component("PAlbumClipboard", PAlbumClipboard);
  app.component("PAlbumToolbar", PAlbumToolbar);
  app.component("PAlbumEditDialog", PAlbumEditDialog);
  app.component("PAlbumDeleteDialog", PAlbumDeleteDialog);

  app.component("PAuthHeader", PAuthHeader);
  app.component("PAuthFooter", PAuthFooter);

  app.component("PShareDialog", PShareDialog);

  app.component("PConfirmDialog", PConfirmDialog);
  app.component("PConfirmSponsor", PConfirmSponsor);

  app.component("PFileClipboard", PFileClipboard);
  app.component("PFileDeleteDialog", PFileDeleteDialog);

  app.component("PLabelClipboard", PLabelClipboard);
  app.component("PLabelDeleteDialog", PLabelDeleteDialog);
  app.component("PLabelEditDialog", PLabelEditDialog);

  app.component("PPeopleMergeDialog", PPeopleMergeDialog);
  app.component("PPeopleEditDialog", PPeopleEditDialog);
  app.component("PPeopleClipboard", PPeopleClipboard);

  app.component("PPhotoViewCards", PPhotoViewCards);
  app.component("PPhotoViewMosaic", PPhotoViewMosaic);
  app.component("PPhotoViewList", PPhotoViewList);
  app.component("PPhotoClipboard", PPhotoClipboard);
  app.component("PPhotoToolbar", PPhotoToolbar);
  app.component("PPhotoPreview", PPhotoPreview);
  app.component("PPhotoArchiveDialog", PPhotoArchiveDialog);
  app.component("PPhotoAlbumDialog", PPhotoAlbumDialog);
  app.component("PPhotoEditDialog", PPhotoEditDialog);
  app.component("PPhotoUploadDialog", PPhotoUploadDialog);

  app.component("PServiceAdd", PServiceAdd);
  app.component("PServiceRemove", PServiceRemove);
  app.component("PServiceEdit", PServiceEdit);
  app.component("PServiceUpload", PServiceUpload);
}
