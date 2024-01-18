const clientConfig = {
  mode: "user",
  name: "PhotoPrism",
  about: "PhotoPrism® CE",
  edition: "ce",
  version: "210710-bae1f2d7-Linux-x86_64-DEBUG",
  copyright: "(c) 2018-2024 PhotoPrism UG. All rights reserved.",
  flags: "public debug experimental settings",
  baseUri: "",
  staticUri: "/static",
  apiUri: "/api/v1",
  contentUri: "/api/v1",
  siteUrl: "http://localhost:2342/",
  sitePreview: "http://localhost:2342/static/img/preview.jpg",
  siteTitle: "PhotoPrism",
  siteCaption: "AI-Powered Photos App",
  siteDescription: "Open-Source Photo Management",
  siteAuthor: "@photoprism_app",
  debug: false,
  readonly: false,
  uploadNSFW: false,
  public: false,
  experimental: true,
  disableSettings: false,
  test: true,
  demo: false,
  sponsor: true,
  albumCategories: ["Animal", "Holiday"],
  albums: [
    {
      ID: 69,
      UID: "aqw0vmr32zb4560f",
      CoverUID: "",
      FolderUID: "",
      Slug: "test-album-1",
      Path: "",
      Type: "album",
      Title: "Test Album 1",
      Location: "",
      Category: "",
      Caption: "",
      Description: "",
      Notes: "",
      Filter: "",
      Order: "oldest",
      Template: "",
      Country: "zz",
      Year: 0,
      Month: 0,
      Day: 0,
      Favorite: true,
      Private: false,
      CreatedAt: "2021-07-10T09:28:03Z",
      UpdatedAt: "2021-07-10T09:28:03Z",
      DeletedAt: null,
    },
    {
      ID: 70,
      UID: "aqw0vmzrkc202vty",
      CoverUID: "",
      FolderUID: "",
      Slug: "test-album-2",
      Path: "",
      Type: "album",
      Title: "Test Album 2",
      Location: "",
      Category: "",
      Caption: "",
      Description: "",
      Notes: "",
      Filter: "",
      Order: "oldest",
      Template: "",
      Country: "zz",
      Year: 0,
      Month: 0,
      Day: 0,
      Favorite: true,
      Private: false,
      CreatedAt: "2021-07-10T09:28:12Z",
      UpdatedAt: "2021-07-10T09:28:12Z",
      DeletedAt: null,
    },
  ],
  cameras: [
    {
      ID: 7,
      Slug: "apple-iphone-se",
      Name: "Apple iPhone SE",
      Make: "Apple",
      Model: "iPhone SE",
    },
    {
      ID: 2,
      Slug: "canon-eos-6d",
      Name: "Canon EOS 6D",
      Make: "Canon",
      Model: "EOS 6D",
    },
    {
      ID: 3,
      Slug: "canon-eos-7d",
      Name: "Canon EOS 7D",
      Make: "Canon",
      Model: "EOS 7D",
    },
    {
      ID: 6,
      Slug: "hmd-global-nokia-x71",
      Name: "HMD Global Nokia X71",
      Make: "HMD Global",
      Model: "Nokia X71",
    },
    {
      ID: 4,
      Slug: "huawei-mate-20-lite",
      Name: "HUAWEI Mate 20 lite",
      Make: "HUAWEI",
      Model: "Mate 20 lite",
    },
    {
      ID: 5,
      Slug: "huawei-p30",
      Name: "HUAWEI P30",
      Make: "HUAWEI",
      Model: "P30",
    },
    {
      ID: 1,
      Slug: "zz",
      Name: "Unknown",
      Make: "",
      Model: "Unknown",
    },
  ],
  lenses: [
    {
      ID: 6,
      Slug: "apple-iphone-se-back-camera-4-15mm-f-2-2",
      Name: "Apple iPhone SE back camera 4.15mm f/2.2",
      Make: "Apple",
      Model: "iPhone SE back camera 4.15mm f/2.2",
      Type: "",
    },
    {
      ID: 3,
      Slug: "ef100mm-f-2-8l-macro-is-usm",
      Name: "EF100mm f/2.8L Macro IS USM",
      Make: "",
      Model: "EF100mm f/2.8L Macro IS USM",
      Type: "",
    },
    {
      ID: 5,
      Slug: "ef16-35mm-f-2-8l-ii-usm",
      Name: "EF16-35mm f/2.8L II USM",
      Make: "",
      Model: "EF16-35mm f/2.8L II USM",
      Type: "",
    },
    {
      ID: 2,
      Slug: "ef24-105mm-f-4l-is-usm",
      Name: "EF24-105mm f/4L IS USM",
      Make: "",
      Model: "EF24-105mm f/4L IS USM",
      Type: "",
    },
    {
      ID: 4,
      Slug: "ef70-200mm-f-4l-is-usm",
      Name: "EF70-200mm f/4L IS USM",
      Make: "",
      Model: "EF70-200mm f/4L IS USM",
      Type: "",
    },
    {
      ID: 1,
      Slug: "zz",
      Name: "Unknown",
      Make: "",
      Model: "Unknown",
      Type: "",
    },
  ],
  countries: [
    {
      ID: "bw",
      Slug: "botswana",
      Name: "Botswana",
    },
    {
      ID: "fr",
      Slug: "france",
      Name: "France",
    },
    {
      ID: "de",
      Slug: "germany",
      Name: "Germany",
    },
    {
      ID: "gr",
      Slug: "greece",
      Name: "Greece",
    },
    {
      ID: "za",
      Slug: "south-africa",
      Name: "South Africa",
    },
    {
      ID: "gb",
      Slug: "united-kingdom",
      Name: "United Kingdom",
    },
    {
      ID: "zz",
      Slug: "zz",
      Name: "Unknown",
    },
  ],
  people: [
    {
      UID: "jr0jgyx2viicdnf7",
      Name: "Andrea Sander",
      Keywords: ["andrea"],
    },
    {
      UID: "jr0jgyx2viicdn88",
      Name: "Otto Sander",
      Keywords: ["andrea"],
    },
    {
      UID: "jr0jgzi2qmp5wt97",
      Name: "Otto Sander",
      Keywords: ["otto", "sander"],
    },
  ],
  thumbs: [
    { size: "fit_720", usage: "SD TV, Mobile", w: 720, h: 720 },
    { size: "fit_1280", usage: "HD TV, SXGA", w: 1280, h: 1024 },
    { size: "fit_1920", usage: "Full HD", w: 1920, h: 1200 },
    { size: "fit_2048", usage: "DCI 2K, Tablets", w: 2048, h: 2048 },
    { size: "fit_2560", usage: "Quad HD, Notebooks", w: 2560, h: 1600 },
    { size: "fit_3840", usage: "4K Ultra HD", w: 3840, h: 2400 },
    { size: "fit_4096", usage: "DCI 4K, Retina 4K", w: 4096, h: 4096 },
    { size: "fit_7680", usage: "8K Ultra HD 2", w: 7680, h: 4320 },
  ],
  status: "unregistered",
  mapKey: "D9ve6edlcVR2mEsNvCXa",
  downloadToken: "2lbh9x09",
  previewToken: "public",
  cssUri: "/static/build/app.2259c0edcc020e7af593.css",
  jsUri: "/static/build/app.9bd7132eaee8e4c7c7e3.js",
  manifestUri: "/manifest.json",
  settings: {
    ui: {
      scrollbar: true,
      zoom: false,
      theme: "default",
      language: "en",
    },
    search: {
      batchSize: 90,
    },
    maps: {
      animate: 0,
      style: "streets",
    },
    features: {
      upload: true,
      download: true,
      private: true,
      review: false,
      files: true,
      videos: true,
      folders: true,
      albums: true,
      moments: true,
      estimates: true,
      people: true,
      labels: true,
      places: true,
      edit: true,
      archive: true,
      delete: false,
      share: true,
      library: true,
      import: true,
      logs: true,
    },
    import: {
      path: "/",
      move: false,
    },
    index: {
      path: "/",
      convert: true,
      rescan: false,
    },
    stack: {
      uuid: true,
      meta: true,
      name: false,
    },
    share: {
      title: "",
    },
    download: {
      name: "file",
    },
    templates: {
      default: "index.gohtml",
    },
  },
  disable: {
    backups: false,
    webdav: false,
    settings: false,
    places: false,
    exiftool: false,
    darktable: false,
    rawtherapee: false,
    sips: true,
    heifconvert: false,
    ffmpeg: false,
    tensorflow: false,
  },
  count: {
    all: 133,
    photos: 132,
    videos: 1,
    cameras: 6,
    lenses: 5,
    countries: 6,
    hidden: 0,
    favorites: 1,
    private: 0,
    private_albums: 0,
    private_folders: 0,
    private_moments: 0,
    private_months: 0,
    private_states: 0,
    review: 22,
    stories: 0,
    albums: 2,
    moments: 4,
    months: 27,
    folders: 23,
    files: 136,
    places: 17,
    states: 8,
    people: 5,
    labels: 22,
    labelMaxPhotos: 118,
  },
  pos: {
    uid: "pqu0xswtrlixbcjp",
    cid: "s2:149c947fca4c",
    utc: "2021-06-01T09:46:52Z",
    lat: 35.2847,
    lng: 23.8122,
  },
  years: [2021, 2020, 2019, 2018, 2017, 2015, 2013, 2012],
  colors: [
    {
      Example: "#AB47BC",
      Name: "Purple",
      Slug: "purple",
    },
    {
      Example: "#FF00FF",
      Name: "Magenta",
      Slug: "magenta",
    },
    {
      Example: "#EC407A",
      Name: "Pink",
      Slug: "pink",
    },
    {
      Example: "#EF5350",
      Name: "Red",
      Slug: "red",
    },
    {
      Example: "#FFA726",
      Name: "Orange",
      Slug: "orange",
    },
    {
      Example: "#D4AF37",
      Name: "Gold",
      Slug: "gold",
    },
    {
      Example: "#FDD835",
      Name: "Yellow",
      Slug: "yellow",
    },
    {
      Example: "#CDDC39",
      Name: "Lime",
      Slug: "lime",
    },
    {
      Example: "#66BB6A",
      Name: "Green",
      Slug: "green",
    },
    {
      Example: "#009688",
      Name: "Teal",
      Slug: "teal",
    },
    {
      Example: "#00BCD4",
      Name: "Cyan",
      Slug: "cyan",
    },
    {
      Example: "#2196F3",
      Name: "Blue",
      Slug: "blue",
    },
    {
      Example: "#A1887F",
      Name: "Brown",
      Slug: "brown",
    },
    {
      Example: "#F5F5F5",
      Name: "White",
      Slug: "white",
    },
    {
      Example: "#9E9E9E",
      Name: "Grey",
      Slug: "grey",
    },
    {
      Example: "#212121",
      Name: "Black",
      Slug: "black",
    },
  ],
  categories: [
    {
      UID: "lqw0teu1kndplci9",
      Slug: "animal",
      Name: "Animal",
    },
    {
      UID: "lqw0tfrbx6e6flcx",
      Slug: "bird",
      Name: "Bird",
    },
    {
      UID: "lqw0tfw28lz7hcpq",
      Slug: "food",
      Name: "Food",
    },
    {
      UID: "lqw0tfqhgq2fr0ga",
      Slug: "insect",
      Name: "Insect",
    },
    {
      UID: "lqw0tfr144mh3jrd",
      Slug: "nature",
      Name: "Nature",
    },
    {
      UID: "lqw0tf72t04mgecr",
      Slug: "outdoor",
      Name: "Outdoor",
    },
    {
      UID: "lqw0teu1jpuk8310",
      Slug: "people",
      Name: "People",
    },
    {
      UID: "lqw0teufc81nxvqt",
      Slug: "portrait",
      Name: "Portrait",
    },
    {
      UID: "lqw0tft3e5qjlfcz",
      Slug: "vehicle",
      Name: "Vehicle",
    },
    {
      UID: "lqw0tft315xza8bk",
      Slug: "water",
      Name: "Water",
    },
    {
      UID: "lqw0tfs1dfgra72o",
      Slug: "wildlife",
      Name: "Wildlife",
    },
  ],
  clip: 160,
  server: {
    cores: 16,
    routines: 26,
    memory: {
      used: 81586008,
      reserved: 148459544,
      info: "Used 82 MB / Reserved 148 MB",
    },
  },
};

window.__CONFIG__ = clientConfig;

export default clientConfig;
