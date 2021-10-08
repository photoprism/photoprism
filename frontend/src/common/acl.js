export const Constants = {
  roles: {
    RoleDefault: "*", // used if no role matches
    RoleAdmin: "admin",
    RolePartner: "partner",
    RoleFamily: "family",
    RoleSibling: "sibling",
    RoleParent: "parent",
    RoleGrandparent: "grandparent",
    RoleChild: "child",
    RoleFriend: "friend",
    RoleBestFriend: "best-friend",
    RoleClassmate: "classmate",
    RoleWorkmate: "workmate",
    RoleGuest: "guest",
    RoleMember: "member",
  },
  actions: {
    ActionDefault: "*", // allows a subject/role to execute all other actions
    ActionSearch: "search",
    ActionCreate: "create",
    ActionRead: "read",
    ActionUpdate: "update",
    ActionUpdateSelf: "update-self",
    ActionDelete: "delete",
    ActionPrivate: "private",
    ActionUpload: "upload",
    ActionDownload: "download",
    ActionShare: "share",
    ActionLike: "like",
    ActionComment: "comment",
    ActionExport: "export",
    ActionImport: "import",
  },
  resources: {
    ResourceDefault: "*",
    ResourceConfig: "config",
    ResourceConfigOptions: "config_options",
    ResourceSettings: "settings",
    ResourceLogs: "logs",
    ResourceAccounts: "accounts",
    ResourceSubjects: "subjects",
    ResourceAlbums: "albums",
    ResourceCameras: "cameras",
    ResourceCategories: "categories",
    ResourceCountries: "countries",
    ResourceFiles: "files",
    ResourceFolders: "folders",
    ResourceLabels: "labels",
    ResourceLenses: "lenses",
    ResourceLinks: "links",
    ResourceGeo: "geo",
    ResourcePasswords: "passwords",
    ResourceUsers: "users",
    ResourcePhotos: "photos",
    ResourcePlaces: "places",
    ResourceFeedback: "feedback",
    ResourceReview: "review",
    ResourceArchive: "archive",
    ResourcePrivate: "private",
    ResourceLibrary: "library",
  },
};

export default class Acl {
  constructor(acl) {
    this.acl = acl;
  }
  accessAllowed(role, resource, action) {
    if (!this.acl) return false;
    console.log("resource: ", resource);
    console.log("role: ", role);
    console.log("action: ", action);
    let res;
    if (!this.acl[resource]) {
      console.log("resource not found");
      if (!this.acl[Constants.resources.ResourceDefault]) return false;
      console.log("using default resource");
      res = this.acl[Constants.resources.ResourceDefault];
    } else {
      res = this.acl[resource];
    }

    let rol;
    if (!res[role]) {
      console.log("role not found");
      if (!res[Constants.roles.RoleDefault]) return false;
      console.log("using default role");
      rol = res[Constants.roles.RoleDefault];
    } else {
      rol = res[role];
    }

    let act;
    if (!rol[action]) {
      console.log("action not found");
      if (!rol[Constants.actions.ActionDefault]) return false;
      console.log("using default action");
      act = rol[Constants.actions.ActionDefault];
    } else {
      act = rol[action];
    }
    console.log("Result: ", act);
    return act;
  }
  accessAllowedAny(role, resource, ...actions) {
    // let result = false;
    // for (const a in actions) {
    //   result = result || this.accessAllowed(role, resource, a);
    // }
    // return result;
    // return actions.reduce((accumulator, action) => {
    //   return accumulator || this.accessAllowed(role, resource, action);
    // });
    // for (const a in actions) {
    //   if (this.accessAllowed(role, resource, a)) return true;
    // }
    // return false;
    return actions.some((action) => this.accessAllowed(role, resource, action));
  }
  getConstants() {
    return Constants;
  }
}
