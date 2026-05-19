import { describe, it, expect, vi } from "vitest";
import { shallowMount } from "@vue/test-utils";
import PSettingsWebdav from "component/settings/webdav.vue";

function mountWebdavDialog({ baseUri = "", userName = "admin", basePath = "", https = false } = {}) {
  return shallowMount(PSettingsWebdav, {
    props: {
      visible: true,
    },
    global: {
      mocks: {
        $session: {
          getUser: () => ({
            Name: userName,
            BasePath: basePath,
          }),
        },
        $config: {
          baseUri,
        },
        $util: {
          isHttps: () => https,
          copyText: vi.fn(),
          openUrl: vi.fn(),
        },
        $view: {
          enter: vi.fn(),
          leave: vi.fn(),
        },
        $gettext: (s) => s,
        $pgettext: (_ctx, s) => s,
      },
    },
  });
}

describe("component/settings/webdav", () => {
  it("shows the root-path WebDAV URL when no base URI is configured", () => {
    const wrapper = mountWebdavDialog({ userName: "user@example.com" });
    const expected = `${window.location.protocol}//${encodeURIComponent("user@example.com")}@${window.location.host}/originals/`;

    expect(wrapper.vm.webdavUrl()).toBe(expected);
  });

  it("includes baseUri in the WebDAV URL and Windows resource", () => {
    const wrapper = mountWebdavDialog({
      baseUri: "/instance/pro-1/",
      userName: "admin",
      basePath: "users/mobile",
      https: true,
    });

    expect(wrapper.vm.webdavUrl()).toBe(`${window.location.protocol}//admin@${window.location.host}/instance/pro-1/originals/users/mobile/`);
    expect(wrapper.vm.windowsUrl()).toContain("\\instance\\pro-1\\originals\\users\\mobile\\");
  });
});
