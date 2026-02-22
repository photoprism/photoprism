import { describe, it, expect, vi, afterEach, beforeEach } from "vitest";
import { shallowMount, config as VTUConfig } from "@vue/test-utils";
import PNavigation from "component/navigation.vue";
import { buildNamespace } from "common/storage";
import clientConfig from "../config";

const navigationModeKey = `${buildNamespace(clientConfig.storageNamespace)}navigation.mode`;

function mountNavigation({
  routeName = "photos",
  routeMeta = { hideNav: false },
  isPublic = false,
  isRestricted = false,
  sessionAuth = true,
  featureOverrides = {},
  configValues = {},
  allowMock,
  routerPush,
  vuetifyDisplay = { smAndDown: false },
  eventPublish,
  utilOverrides = {},
  sessionOverrides = {},
} = {}) {
  const baseConfig = VTUConfig.global.mocks.$config || {};
  const baseEvent = VTUConfig.global.mocks.$event || {};
  const baseUtil = VTUConfig.global.mocks.$util || {};
  const baseNotify = VTUConfig.global.mocks.$notify || {};

  const featureFlags = {
    files: true,
    settings: true,
    upload: true,
    account: true,
    logs: true,
    library: true,
    places: true,
    ...featureOverrides,
  };

  const values = {
    siteUrl: "http://localhost:2342/",
    usage: { filesTotal: 1024, filesUsed: 512 },
    legalUrl: configValues.legalUrl ?? null,
    legalInfo: configValues.legalInfo ?? "",
    disable: { settings: false },
    count: {},
    ...configValues,
  };

  const configMock = {
    ...baseConfig,
    getName: baseConfig.getName || vi.fn(() => "PhotoPrism"),
    getAbout: baseConfig.getAbout || vi.fn(() => "About"),
    getIcon: baseConfig.getIcon || vi.fn(() => "/icon.png"),
    getTier: baseConfig.getTier || vi.fn(() => 1),
    isPro: baseConfig.isPro || vi.fn(() => false),
    isSponsor: baseConfig.isSponsor || vi.fn(() => false),
    get: vi.fn((key) => {
      if (key === "demo") return false;
      if (key === "public") return isPublic;
      if (key === "readonly") return false;
      return false;
    }),
    feature: vi.fn((name) => {
      if (name in featureFlags) {
        return !!featureFlags[name];
      }
      return true;
    }),
    allow: allowMock || baseConfig.allow || vi.fn(() => true),
    deny: vi.fn((resource, action) => (resource === "photos" && action === "access_library" ? isRestricted : false)),
    values,
    disconnected: false,
    page: { title: "Photos" },
    test: false,
  };

  const session = {
    auth: sessionAuth,
    isAdmin: vi.fn(() => true),
    isSuperAdmin: vi.fn(() => true),
    hasScope: vi.fn(() => false),
    getUser: vi.fn(() => ({
      getDisplayName: vi.fn(() => "Test User"),
      getAccountInfo: vi.fn(() => "test@example.com"),
      getAvatarURL: vi.fn(() => "/avatar.jpg"),
    })),
    logout: vi.fn(),
    ...sessionOverrides,
  };

  const publish = eventPublish || baseEvent.publish || vi.fn();

  const eventBus = {
    ...baseEvent,
    publish,
    subscribe: baseEvent.subscribe || vi.fn(() => "sub-id"),
    unsubscribe: baseEvent.unsubscribe || vi.fn(),
  };

  const notify = {
    ...baseNotify,
    info: baseNotify.info || vi.fn(),
    blockUI: baseNotify.blockUI || vi.fn(),
  };

  const util = {
    ...baseUtil,
    openExternalUrl: vi.fn(),
    gigaBytes: vi.fn((bytes) => bytes),
    ...utilOverrides,
  };

  const push = routerPush || vi.fn();

  const wrapper = shallowMount(PNavigation, {
    global: {
      mocks: {
        $config: configMock,
        $session: session,
        $router: { push },
        $route: { name: routeName, meta: routeMeta },
        $vuetify: { display: { smAndDown: !!vuetifyDisplay.smAndDown } },
        $event: eventBus,
        $util: util,
        $notify: notify,
        $isRtl: false,
      },
      stubs: {
        "router-link": { template: "<a><slot /></a>" },
        "v-navigation-drawer": true,
      },
    },
  });

  return {
    wrapper,
    configMock,
    session,
    eventBus,
    notify,
    util,
    push,
  };
}

describe("component/navigation", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("routeName", () => {
    it("returns true when current route starts with given name", () => {
      const { wrapper } = mountNavigation({ routeName: "photos_browse" });
      expect(wrapper.vm.routeName("photos")).toBe(true);
      expect(wrapper.vm.routeName("albums")).toBe(false);
    });

    it("returns false when name or route name is missing", () => {
      const { wrapper } = mountNavigation({ routeName: "" });
      expect(wrapper.vm.routeName("photos")).toBe(false);
      expect(wrapper.vm.routeName("")).toBe(false);
    });
  });

  describe("auth and visibility", () => {
    it("auth is true when session is authenticated", () => {
      const { wrapper } = mountNavigation({ sessionAuth: true, isPublic: false });
      expect(wrapper.vm.auth).toBe(true);
    });

    it("auth is true when instance is public even without session", () => {
      const { wrapper } = mountNavigation({ sessionAuth: false, isPublic: true });
      expect(wrapper.vm.auth).toBe(true);
    });

    it("auth is false when neither session nor public access is available", () => {
      const { wrapper } = mountNavigation({ sessionAuth: false, isPublic: false });
      expect(wrapper.vm.auth).toBe(false);
    });

    it("visible is false when route meta.hideNav is true", () => {
      const { wrapper } = mountNavigation({ routeMeta: { hideNav: true } });
      expect(wrapper.vm.visible).toBe(false);
    });
  });

  describe("drawer behavior", () => {
    it("toggleDrawer toggles drawer on small screens", () => {
      const { wrapper } = mountNavigation({
        vuetifyDisplay: { smAndDown: true },
        sessionAuth: true,
      });

      // Force small-screen mode and authenticated session
      wrapper.vm.$vuetify.display.smAndDown = true;
      wrapper.vm.session.auth = true;
      wrapper.vm.isPublic = false;

      wrapper.vm.drawer = false;
      wrapper.vm.toggleDrawer({ target: {} });
      expect(wrapper.vm.drawer).toBe(true);

      wrapper.vm.toggleDrawer({ target: {} });
      expect(wrapper.vm.drawer).toBe(false);
    });

    it("toggleDrawer toggles mini mode on desktop", () => {
      const { wrapper } = mountNavigation({
        vuetifyDisplay: { smAndDown: false },
        isRestricted: false,
      });

      const initial = wrapper.vm.isMini;
      wrapper.vm.toggleDrawer({ target: {} });
      expect(wrapper.vm.isMini).toBe(!initial);
    });

    it("toggleIsMini respects restricted mode and updates localStorage", () => {
      const setItemSpy = vi.spyOn(Storage.prototype, "setItem");

      const { wrapper } = mountNavigation({ isRestricted: false });
      const initial = wrapper.vm.isMini;

      wrapper.vm.toggleIsMini();
      expect(wrapper.vm.isMini).toBe(!initial);
      expect(setItemSpy).toHaveBeenCalledWith(navigationModeKey, `${!initial}`);

      wrapper.vm.isRestricted = true;
      const before = wrapper.vm.isMini;
      wrapper.vm.toggleIsMini();
      expect(wrapper.vm.isMini).toBe(before);
    });
  });

  describe("account and legal navigation", () => {
    it("showAccountSettings routes to account settings when account feature is enabled", () => {
      const { wrapper, push } = mountNavigation({
        featureOverrides: { account: true },
      });

      wrapper.vm.showAccountSettings();
      expect(push).toHaveBeenCalledWith({ name: "settings_account" });
    });

    it("showAccountSettings falls back to general settings when account feature is disabled", () => {
      const { wrapper, push } = mountNavigation({
        featureOverrides: { account: false },
      });

      wrapper.vm.showAccountSettings();
      expect(push).toHaveBeenCalledWith({ name: "settings" });
    });

    it("showLegalInfo opens external URL when legalUrl is configured", () => {
      const { wrapper, util } = mountNavigation({
        configValues: { legalUrl: "https://example.com/legal" },
      });

      wrapper.vm.showLegalInfo();
      expect(util.openExternalUrl).toHaveBeenCalledWith("https://example.com/legal");
    });

    it("showLegalInfo routes to about page when legalUrl is missing", () => {
      const { wrapper, push } = mountNavigation({
        configValues: { legalUrl: null },
      });

      wrapper.vm.showLegalInfo();
      expect(push).toHaveBeenCalledWith({ name: "about" });
    });
  });

  describe("home and upload actions", () => {
    it("onHome toggles drawer on small screens and does not navigate", () => {
      const { wrapper, push } = mountNavigation({
        vuetifyDisplay: { smAndDown: true },
        routeName: "browse",
      });

      // Ensure mobile mode and authenticated session so drawer logic runs
      wrapper.vm.$vuetify.display.smAndDown = true;
      wrapper.vm.session.auth = true;
      wrapper.vm.isPublic = false;
      wrapper.vm.drawer = false;

      wrapper.vm.onHome({ target: {} });
      expect(wrapper.vm.drawer).toBe(true);
      expect(push).not.toHaveBeenCalled();
    });

    it("onHome navigates to home on desktop when not already there", () => {
      const { wrapper, push } = mountNavigation({
        vuetifyDisplay: { smAndDown: false },
        routeName: "albums",
      });

      // Force desktop mode explicitly to avoid relying on Vuetify defaults
      wrapper.vm.$vuetify.display.smAndDown = false;

      wrapper.vm.onHome({ target: {} });
      expect(push).toHaveBeenCalledWith({ name: "home" });
    });

    it("openUpload publishes dialog.upload event", () => {
      const publish = vi.fn();
      const { wrapper, eventBus } = mountNavigation({ eventPublish: publish });

      wrapper.vm.openUpload();
      expect(eventBus.publish).toHaveBeenCalledWith("dialog.upload");
    });
  });

  describe("info and usage actions", () => {
    it("reloadApp shows info notification and blocks UI", () => {
      vi.useFakeTimers();
      const { wrapper, notify } = mountNavigation();
      const setTimeoutSpy = vi.spyOn(global, "setTimeout");

      wrapper.vm.reloadApp();

      expect(notify.info).toHaveBeenCalledWith("Reloading…");
      expect(notify.blockUI).toHaveBeenCalled();
      expect(setTimeoutSpy).toHaveBeenCalled();
      vi.useRealTimers();
    });

    it("showUsageInfo routes to index files", () => {
      const { wrapper, push } = mountNavigation();
      wrapper.vm.showUsageInfo();
      expect(push).toHaveBeenCalledWith({ path: "/index/files" });
    });

    it("showServerConnectionHelp routes to websockets help", () => {
      const { wrapper, push } = mountNavigation();
      wrapper.vm.showServerConnectionHelp();
      expect(push).toHaveBeenCalledWith({ path: "/help/websockets" });
    });
  });

  describe("indexing state", () => {
    it("onIndex sets indexing true for file, folder and indexing events", () => {
      const { wrapper } = mountNavigation();

      wrapper.vm.onIndex("index.file");
      expect(wrapper.vm.indexing).toBe(true);

      wrapper.vm.onIndex("index.folder");
      expect(wrapper.vm.indexing).toBe(true);

      wrapper.vm.onIndex("index.indexing");
      expect(wrapper.vm.indexing).toBe(true);
    });

    it("onIndex sets indexing false when completed", () => {
      const { wrapper } = mountNavigation();

      wrapper.vm.indexing = true;
      wrapper.vm.onIndex("index.completed");
      expect(wrapper.vm.indexing).toBe(false);
    });
  });

  describe("logout", () => {
    it("onLogout calls session.logout", () => {
      const logout = vi.fn();
      const { wrapper, session } = mountNavigation({
        sessionOverrides: { logout },
      });

      wrapper.vm.onLogout();
      expect(session.logout).toHaveBeenCalled();
    });
  });
});
