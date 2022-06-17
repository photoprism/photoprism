import Albums from "share/albums.vue";
import AlbumPhotos from "share/photos.vue";

const c = window.__CONFIG__;
const siteTitle = c.siteAuthor ? c.siteAuthor : c.name;
const shareTitle = c.settings.share.title ? c.settings.share.title : siteTitle;

export default [
  {
    name: "home",
    path: "/",
    redirect: { name: "albums" },
  },
  {
    name: "albums",
    path: "/s/:token",
    component: Albums,
    meta: { title: shareTitle, auth: true, hideNav: true },
    props: { view: "album", staticFilter: { type: "" } },
  },
  {
    name: "album",
    path: "/s/:token/:uid",
    component: AlbumPhotos,
    meta: { title: shareTitle, auth: true, hideNav: true },
  },
  {
    path: "*",
    redirect: { name: "albums" },
  },
];
