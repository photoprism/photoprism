import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
import PeopleTab from "component/photo/edit/people.vue";
import Subject from "model/subject";
import { Marker } from "model/marker";

describe("PTabPhotoPeople face actions", () => {
  let wrapper;
  let setCoverSpy;
  let windowOpenSpy;
  let assignedMarker;
  let unassignedMarker;

  const mockPeople = [
    {
      UID: "js6sg6b2h8njw0sx",
      Slug: "john-doe",
      Name: "John Doe",
      Thumb: "",
      ThumbSrc: "",
    },
  ];

  beforeEach(() => {
    setCoverSpy = vi.spyOn(Subject.prototype, "setCover").mockImplementation(function (thumb) {
      this.Thumb = thumb;
      this.ThumbSrc = "manual";
      return Promise.resolve(this);
    });

    windowOpenSpy = vi.spyOn(window, "open").mockImplementation(() => {});

    assignedMarker = new Marker({
      UID: "marker-assigned",
      SubjUID: "js6sg6b2h8njw0sx",
      Invalid: false,
      Thumb: "hash-1234",
      Name: "John Doe",
    });

    unassignedMarker = new Marker({
      UID: "marker-unassigned",
      SubjUID: "",
      Invalid: false,
      Thumb: "",
      Name: "",
    });

    wrapper = mount(PeopleTab, {
      props: {
        uid: "test-photo-uid",
      },
      global: {
        mocks: {
          $gettext: (msg) => msg,
          $notify: {
            blockUI: vi.fn(),
            unblockUI: vi.fn(),
            success: vi.fn(),
            error: vi.fn(),
          },
          $config: {
            values: {
              people: mockPeople,
            },
            feature: vi.fn(() => true),
            get: vi.fn(() => false),
          },
          $router: {
            resolve: vi.fn((route) => ({
              href: `/library/${route.name || "all"}?q=${route.query?.q || ""}`,
            })),
            push: vi.fn(),
          },
          $view: {
            getData: vi.fn(() => ({
              model: {
                getMarkers: vi.fn(() => [assignedMarker, unassignedMarker]),
              },
            })),
          },
        },
        stubs: {
          PConfirmDialog: true,
          PActionMenu: true,
        },
      },
    });

    // Mock loadSubject method
    wrapper.vm.loadSubject = vi.fn(async (uid) => {
      const subject = new Subject({ UID: uid, Slug: "loaded-person", Name: "Loaded Person" });
      subject.getValues = function () {
        return { ...this };
      };
      return subject;
    });
  });

  afterEach(() => {
    setCoverSpy.mockRestore();
    windowOpenSpy.mockRestore();
    if (wrapper) {
      wrapper.unmount();
    }
  });

  it("provides go-to-person and set-cover actions for assigned faces", async () => {
    const marker = { SubjUID: "js6sg6b2h8njw0sx", Invalid: false, Thumb: "hash-1234" };

    const actions = wrapper.vm.getFaceActions(marker);
    const visible = actions.filter((action) => action.visible).map((action) => action.name);

    expect(visible).toEqual(["go-to-person", "set-person-cover"]);
    expect(actions.some((action) => action.name === "remove-face")).toBe(false);

    wrapper.vm.markers = [assignedMarker];
    await wrapper.vm.$nextTick();

    const removeButton = wrapper.find(`[data-id="${assignedMarker.UID}"] .action-reject`);
    expect(removeButton.exists()).toBe(false);
  });

  it("renders quick remove button for unassigned faces", async () => {
    const marker = { SubjUID: "", Invalid: false };

    const actions = wrapper.vm.getFaceActions(marker);
    const visible = actions.filter((action) => action.visible).map((action) => action.name);

    expect(visible).toEqual([]);

    wrapper.vm.markers = [unassignedMarker];
    await wrapper.vm.$nextTick();

    const removeButton = wrapper.find(`[data-id="${unassignedMarker.UID}"] .action-reject`);
    expect(removeButton.exists()).toBe(true);
  });

  it("opens subject route in new window when navigating to person", async () => {
    const marker = { SubjUID: "js6sg6b2h8njw0sx" };

    await wrapper.vm.onGoToPerson(marker);

    expect(wrapper.vm.$router.resolve).toHaveBeenCalledWith({ name: "all", query: { q: "person:john-doe" } });
    expect(windowOpenSpy).toHaveBeenCalledWith("/library/all?q=person:john-doe", "_blank");
  });

  it("sets person cover with manual thumb source", async () => {
    const marker = { SubjUID: "js6sg6b2h8njw0sx", Thumb: "hash-1234" };

    await wrapper.vm.onSetPersonCover(marker);

    expect(wrapper.vm.$notify.blockUI).toHaveBeenCalledWith("busy");
    expect(setCoverSpy).toHaveBeenCalledWith("hash-1234");
    expect(wrapper.vm.$config.values.people[0].Thumb).toBe("hash-1234");
    expect(wrapper.vm.$config.values.people[0].ThumbSrc).toBe("manual");
    expect(wrapper.vm.$notify.success).toHaveBeenCalledWith("Person cover updated");
    expect(wrapper.vm.busy).toBe(false);
  });

  describe("hasFaceMenu", () => {
    it("returns true when at least one action is visible", () => {
      const marker = new Marker({
        UID: "marker1",
        SubjUID: "js6sg6b2h8njw0sx",
        Invalid: false,
        Thumb: "hash-1234",
      });

      expect(wrapper.vm.hasFaceMenu(marker)).toBe(true);
    });

    it("returns false when no actions are visible (invalid face)", () => {
      const marker = new Marker({
        UID: "marker2",
        SubjUID: "xxx",
        Invalid: true,
        Thumb: "hash",
      });

      expect(wrapper.vm.hasFaceMenu(marker)).toBe(false);
    });

    it("returns false for unassigned valid faces", () => {
      const marker = new Marker({
        UID: "marker3",
        SubjUID: "",
        Invalid: false,
      });

      expect(wrapper.vm.hasFaceMenu(marker)).toBe(false);
    });
  });

  describe("Component integration", () => {
    it("renders PActionMenu stub for markers", async () => {
      await wrapper.vm.$nextTick();

      const actionMenus = wrapper.findAllComponents({ name: "PActionMenu" });
      expect(actionMenus.length).toBeGreaterThan(0);
    });

    it("passes correct props to PActionMenu", async () => {
      await wrapper.vm.$nextTick();

      const actionMenu = wrapper.findComponent({ name: "PActionMenu" });
      expect(actionMenu.exists()).toBe(true);

      // Check props
      expect(actionMenu.props("buttonIcon")).toBe("mdi-dots-vertical");
      expect(actionMenu.props("buttonClass")).toBe("input-menu");
      expect(actionMenu.props("items")).toBeInstanceOf(Function);

      // Call items function to verify it returns correct actions
      const actions = actionMenu.props("items")();
      expect(Array.isArray(actions)).toBe(true);
      expect(actions.length).toBe(2); // go-to-person, set-person-cover

      const actionNames = actions.map((a) => a.name);
      expect(actionNames).toContain("go-to-person");
      expect(actionNames).toContain("set-person-cover");
    });
  });
});
