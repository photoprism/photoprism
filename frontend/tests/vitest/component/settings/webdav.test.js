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

  it("keeps the Windows resource host for host names and IPv4 addresses", () => {
    const wrapper = mountWebdavDialog();

    expect(wrapper.vm.windowsUrl()).toBe(`\\\\${window.location.host}\\originals\\`);
    expect(wrapper.vm.uncHost("photos.example.com")).toBe("photos.example.com");
    expect(wrapper.vm.uncHost("192.0.2.1")).toBe("192.0.2.1");
  });

  it("writes IPv6 addresses as ipv6-literal.net names for Windows", () => {
    const wrapper = mountWebdavDialog();

    expect(wrapper.vm.uncHost("[2001:db8::1]")).toBe("2001-db8--1.ipv6-literal.net");
    expect(wrapper.vm.uncHost("[fe80::1%4]")).toBe("fe80--1s4.ipv6-literal.net");
    expect(wrapper.vm.windowsUrl({ hostname: "[2001:db8::1]", port: "2342" })).toBe("\\\\2001-db8--1.ipv6-literal.net:2342\\originals\\");
  });

  it("writes IPv6 addresses as ipv6-literal.net names in HTTPS Windows paths", () => {
    const wrapper = mountWebdavDialog({ https: true });

    expect(wrapper.vm.windowsUrl({ hostname: "[2001:db8::1]", port: "8443" })).toBe("\\\\2001-db8--1.ipv6-literal.net@SSL@8443\\originals\\");
    expect(wrapper.vm.windowsUrl({ hostname: "[2001:db8::1]", port: "" })).toBe("\\\\2001-db8--1.ipv6-literal.net@SSL\\originals\\");
    expect(wrapper.vm.windowsUrl({ hostname: "photos.example.com", port: "" })).toBe("\\\\photos.example.com@SSL\\originals\\");
  });
});
