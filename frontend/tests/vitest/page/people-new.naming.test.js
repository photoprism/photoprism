// Pins the commit paths of the name field in page/people/new.vue.
//
// Naming a face here creates a person when the typed name matches nobody, and a person created by
// accident is not a typo the user can undo: the matcher then treats one person as two and narrows
// both clusters to keep them apart. So the rule this file pins is that only an explicit choice
// commits - typing does not, and leaving the field does not.
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";

import PPageFaces from "page/people/new.vue";
import PConfirmDialog from "component/confirm/dialog.vue";
import Face from "model/face";
import typeaheadCache from "common/typeahead-cache";

const people = [
  { UID: "js6sg6b2h8njw0sx", Name: "John Doe" },
  { UID: "js6sg6b1h1njaaaa", Name: "Jane Roe" },
];

// Interpolates like vue3-gettext rather than returning the message unchanged. A stub that ignores
// the params cannot tell a matching placeholder key from a mismatched one, which is the whole bug
// the prompt test below exists to catch.
const interpolate = (msg, params) => String(msg).replace(/%\{(\w+)\}/g, (all, key) => (params && key in params ? params[key] : all));

const mocks = () => ({
  $gettext: (msg) => msg,
  $gettextInterpolate: interpolate,
  $notify: { info: vi.fn(), warn: vi.fn(), error: vi.fn(), success: vi.fn(), blockUI: vi.fn(), unblockUI: vi.fn() },
  $config: { values: {}, get: vi.fn(() => false), feature: vi.fn(() => true) },
  $route: { query: {}, name: "people_faces" },
  $router: { replace: vi.fn(), push: vi.fn() },
  $event: { subscribe: vi.fn(() => 1), unsubscribe: vi.fn() },
  $view: { enter: vi.fn(), leave: vi.fn() },
});

describe("PPageFaces name input", () => {
  let wrapper;
  let setName;

  const mountPage = async (stubs) => {
    wrapper = mount(PPageFaces, {
      props: { staticFilter: { markers: true, unknown: true }, active: true },
      global: {
        mocks: mocks(),
        components: { PConfirmDialog },
        stubs: Object.assign({ VImg: true, PScroll: true, PLoading: true }, stubs),
      },
    });

    await flushPromises();

    return wrapper;
  };

  beforeEach(() => {
    typeaheadCache.clear();
    vi.spyOn(typeaheadCache, "getPeople").mockResolvedValue(people);

    setName = vi.spyOn(Face.prototype, "setName").mockImplementation(function () {
      return Promise.resolve(this);
    });

    vi.spyOn(Face, "search").mockResolvedValue({
      models: [new Face({ ID: "FACE1", SubjUID: "", Name: "", MarkerUID: "ms6sg6b1wowuy666" })],
      count: 1,
      limit: 999,
      offset: 0,
    });
  });

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount();
      wrapper = null;
    }
    vi.restoreAllMocks();
    typeaheadCache.clear();
  });

  it("keeps the typed text off the model", async () => {
    await mountPage({ VCombobox: { name: "VCombobox", template: "<div></div>" }, VTextField: true });

    const m = wrapper.vm.results[0];
    wrapper.vm.onNameInput(m, "k");

    expect(wrapper.vm.nameInput[m.ID]).toBe("k");
    expect(m.Name).toBe("");
    expect(setName).not.toHaveBeenCalled();
  });

  // Driven through the real widget. Vuetify commits free text on blur by itself, so what makes
  // leaving the field safe is that the page ignores that event and resolves the text on its own.
  // An unknown name is neither saved nor thrown away - a modal raised by clicking elsewhere would
  // be its own surprise, and discarding the text would lose work.
  it("neither saves nor prompts when the field loses focus with an unknown name", async () => {
    await mountPage();

    const input = wrapper.find(".input-name input");
    await input.setValue("k");
    await input.trigger("blur");
    await flushPromises();

    expect(setName).not.toHaveBeenCalled();
    expect(wrapper.vm.confirm.visible).toBe(false);
    expect(wrapper.vm.results[0].SubjUID).toBe("");
    expect(wrapper.vm.results[0].Name).toBe("");
    expect(wrapper.vm.nameInput[wrapper.vm.results[0].ID]).toBe("k");
  });

  // A name that already belongs to someone is unambiguous, so blur may commit it.
  it("links on blur when the typed name is an exact match", async () => {
    await mountPage({ VCombobox: { name: "VCombobox", template: "<div></div>" }, VTextField: true });

    const m = wrapper.vm.results[0];
    wrapper.vm.onNameInput(m, "Jane Roe");
    await wrapper.vm.onBlurName(m);
    await flushPromises();

    expect(setName).toHaveBeenCalledWith("Jane Roe");
    expect(m.SubjUID).toBe("js6sg6b1h1njaaaa");
    expect(wrapper.vm.confirm.visible).toBe(false);
  });

  it("asks before creating a person the typed name does not match", async () => {
    await mountPage({ VCombobox: { name: "VCombobox", template: "<div></div>" }, VTextField: true });

    const m = wrapper.vm.results[0];
    wrapper.vm.onNameInput(m, "k");
    await wrapper.vm.onSubmitName(m);
    await flushPromises();

    expect(wrapper.vm.confirm.visible).toBe(true);
    expect(wrapper.vm.confirm.name).toBe("k");
    expect(setName).not.toHaveBeenCalled();
    expect(m.Name).toBe("");
  });

  it("creates the person once the prompt is confirmed", async () => {
    await mountPage({ VCombobox: { name: "VCombobox", template: "<div></div>" }, VTextField: true });

    const m = wrapper.vm.results[0];
    wrapper.vm.onNameInput(m, "k");
    await wrapper.vm.onSubmitName(m);
    await flushPromises();

    wrapper.vm.onConfirmCreatePerson();
    await flushPromises();

    expect(setName).toHaveBeenCalledWith("k");
    expect(m.SubjUID).toBe("");
    expect(wrapper.vm.confirm.visible).toBe(false);
  });

  it("discards the name when the prompt is cancelled", async () => {
    await mountPage({ VCombobox: { name: "VCombobox", template: "<div></div>" }, VTextField: true });

    const m = wrapper.vm.results[0];
    wrapper.vm.onNameInput(m, "k");
    await wrapper.vm.onSubmitName(m);
    await flushPromises();

    wrapper.vm.onCancelCreatePerson();
    await flushPromises();

    expect(setName).not.toHaveBeenCalled();
    expect(wrapper.vm.confirm.visible).toBe(false);
    expect(m.Name).toBe("");
  });

  // Resolved through the cache rather than the component's own list, which is empty until the
  // suggestions load - the window that let a typed name miss a person who does exist.
  it("links a typed name that matches an existing person, whatever its case", async () => {
    await mountPage({ VCombobox: { name: "VCombobox", template: "<div></div>" }, VTextField: true });

    const m = wrapper.vm.results[0];
    wrapper.vm.people = [];
    wrapper.vm.onNameInput(m, "john doe");
    await wrapper.vm.onSubmitName(m);
    await flushPromises();

    expect(wrapper.vm.confirm.visible).toBe(false);
    expect(setName).toHaveBeenCalledWith("John Doe");
    expect(m.SubjUID).toBe("js6sg6b2h8njw0sx");
  });

  it("commits an item picked from the list without asking", async () => {
    await mountPage({ VCombobox: { name: "VCombobox", template: "<div></div>" }, VTextField: true });

    const m = wrapper.vm.results[0];
    await wrapper.vm.onSelectPerson(m, people[1]);
    await flushPromises();

    expect(wrapper.vm.confirm.visible).toBe(false);
    expect(setName).toHaveBeenCalledWith("Jane Roe");
    expect(m.SubjUID).toBe("js6sg6b1h1njaaaa");
  });

  // The prompt has to name the person it would create, and nothing else checks that it does.
  // A params key that does not match the msgid placeholder is not an error in vue3-gettext - the
  // literal token renders and no gate objects: gettext-lint compares placeholders between msgid and
  // msgstr, never against what a caller passes, and jsdom renders both strings equally happily.
  it("names the person in the prompt it renders", async () => {
    await mountPage();

    const input = wrapper.find(".input-name input");
    await input.setValue("Testperson Alpha");
    await input.trigger("keyup.enter");
    await flushPromises();

    expect(wrapper.vm.confirm.visible).toBe(true);

    const dialog = wrapper.findComponent({ name: "PConfirmDialog" });
    const text = dialog.props("text");

    expect(text).toContain("Testperson Alpha");
    expect(text).not.toContain("%{");
  });

  it("acknowledges a save instead of leaving the user to infer it from a list", async () => {
    const page = await mountPage({ VCombobox: { name: "VCombobox", template: "<div></div>" }, VTextField: true });

    const m = page.vm.results[0];
    await page.vm.onSelectPerson(m, people[0]);
    await flushPromises();

    expect(page.vm.$notify.success).toHaveBeenCalled();
    expect(page.vm.nameInput[m.ID]).toBeUndefined();
  });

  // Regression: the prompt is raised by an Enter keydown, and the matching keyup lands on the
  // dialog that just mounted. p-confirm-dialog confirms on Enter by default, so the same keypress
  // created the person the prompt was asking about - which is the whole defect, reintroduced one
  // layer up. Only a full keydown/keyup pair reproduces it, which is why a keydown-only case did
  // not, and why it took a real browser to surface it.
  it("does not let the Enter that raised the prompt also answer it", async () => {
    await mountPage();

    const input = wrapper.find(".input-name input");
    await input.setValue("k");
    await input.trigger("keydown.enter");
    await input.trigger("keyup.enter");
    await flushPromises();

    expect(wrapper.vm.confirm.visible).toBe(true);

    const dialog = wrapper.findComponent({ name: "PConfirmDialog" });
    expect(dialog.exists()).toBe(true);
    expect(dialog.props("confirmOnEnter")).toBe(false);

    dialog.vm.onEnter();
    await flushPromises();

    expect(setName).not.toHaveBeenCalled();
    expect(wrapper.vm.confirm.visible).toBe(true);
  });

  // The load-bearing assumption behind dropping @keyup.enter: the widget itself reports free text
  // through update:modelValue, so one event carries either the chosen person or the typed string
  // and the two cannot race. Driven through a real VCombobox rather than a stub, because a stub
  // would pin what this file assumes rather than what Vuetify does.
  it("routes Enter on free text through the confirmation, using a real combobox", async () => {
    // VTextField is deliberately not stubbed here: VCombobox renders one internally, so stubbing
    // it removes the very input this case has to type into.
    await mountPage();

    const input = wrapper.find(".input-name input");
    expect(input.exists()).toBe(true);

    await input.setValue("k");
    await input.trigger("keyup.enter");
    await flushPromises();

    expect(setName).not.toHaveBeenCalled();
    expect(wrapper.vm.confirm.visible).toBe(true);
    expect(wrapper.vm.confirm.name).toBe("k");
  });

  // Regression: the combobox reports every keystroke through update:model-value, so treating that
  // event as a commit raised the prompt on the first letter, stole focus, and made the suggestion
  // list unreachable - typing "Mic" never got far enough to show "Micha". Typing has to be inert.
  it("stays quiet while the name is being typed", async () => {
    await mountPage();

    const input = wrapper.find(".input-name input");

    for (const value of ["M", "Mi", "Mic"]) {
      await input.setValue(value);
      await flushPromises();

      expect(wrapper.vm.confirm.visible).toBe(false);
      expect(setName).not.toHaveBeenCalled();
    }

    expect(wrapper.vm.nameInput[wrapper.vm.results[0].ID]).toBe("Mic");
  });
});
