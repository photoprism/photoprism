import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import { nextTick } from "vue";
import PPhotoTimeline from "component/photo/timeline.vue";
import Photo from "model/photo";

function deferred() {
  let resolve;
  let reject;
  const promise = new Promise((res, rej) => {
    resolve = res;
    reject = rej;
  });

  return { promise, resolve, reject };
}

function mountTimeline(props = {}) {
  return mount(PPhotoTimeline, {
    props: {
      filter: {},
      staticFilter: {},
      updateQuery: vi.fn(),
      ...props,
    },
  });
}

describe("component/photo/timeline", () => {
  let originalTimeline;

  beforeEach(() => {
    originalTimeline = Photo.timeline;
    Photo.timeline = vi.fn();
  });

  afterEach(() => {
    if (originalTimeline) {
      Photo.timeline = originalTimeline;
    } else {
      delete Photo.timeline;
    }

    vi.restoreAllMocks();
  });

  it("maps month button clicks to year, month, and day query updates", async () => {
    const updateQuery = vi.fn();
    const searchSpy = vi.spyOn(Photo, "search");

    Photo.timeline.mockResolvedValue({
      buckets: [
        {
          key: "2025-04",
          label: "April 2025",
          year: 2025,
          month: 4,
          photoCount: 3,
        },
      ],
      unknownDateCount: 2,
    });

    const wrapper = mountTimeline({
      filter: { q: "kitten", type: "image", year: 2025, month: 4, day: 12 },
      staticFilter: { type: "raw" },
      updateQuery,
    });

    await flushPromises();
    await nextTick();

    expect(Photo.timeline).toHaveBeenCalledWith({
      q: "kitten",
      type: "raw",
      bucket: "month",
    });
    expect(wrapper.emitted("visibility").at(-1)).toEqual([true]);
    expect(searchSpy).not.toHaveBeenCalled();
    expect(wrapper.find('[data-testid="timeline-unknown"]').text()).toContain("2");
    expect(wrapper.find('[data-testid="timeline-month-2025-4"]').attributes("aria-label")).toBe("Filter photos from April 2025, 3 pictures");
    expect(wrapper.find('[data-testid="timeline-month-2025-4"]').attributes("aria-pressed")).toBe("true");

    await wrapper.find('[data-testid="timeline-month-2025-4"]').trigger("click");

    expect(updateQuery).toHaveBeenCalledWith({ year: 2025, month: 4, day: 0 });
  });

  it("keeps the rail mounted while loading and reloads when refreshToken changes", async () => {
    const first = deferred();
    const second = deferred();

    Photo.timeline.mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);

    const wrapper = mountTimeline({
      filter: { q: "kitten" },
      refreshToken: 0,
    });

    await nextTick();

    expect(wrapper.find('[data-testid="photo-timeline"]').exists()).toBe(true);
    expect(wrapper.find('[data-testid="photo-timeline"]').attributes("aria-busy")).toBe("true");
    expect(wrapper.emitted("visibility").at(-1)).toEqual([true]);

    await wrapper.setProps({ refreshToken: 1 });

    expect(Photo.timeline).toHaveBeenCalledTimes(2);

    second.resolve({
      buckets: [
        {
          key: "2025-04",
          label: "April 2025",
          year: 2025,
          month: 4,
          photoCount: 3,
        },
      ],
      unknownDateCount: 0,
    });

    await flushPromises();
    await nextTick();

    expect(wrapper.find('[data-testid="photo-timeline"]').attributes("aria-busy")).toBe("false");
    expect(wrapper.find('[data-testid="timeline-month-2025-4"]').exists()).toBe(true);
    expect(wrapper.emitted("visibility").at(-1)).toEqual([true]);
  });

  it("collapses parent rail visibility after an empty result or error", async () => {
    Photo.timeline.mockResolvedValue({
      buckets: [],
      unknownDateCount: 0,
    });

    const wrapper = mountTimeline();

    await flushPromises();
    await nextTick();

    expect(wrapper.find('[data-testid="photo-timeline"]').exists()).toBe(true);
    expect(wrapper.find('[data-testid="photo-timeline"]').attributes("aria-busy")).toBe("false");
    expect(wrapper.emitted("visibility").at(-1)).toEqual([false]);

    Photo.timeline.mockRejectedValue(new Error("offline"));
    await wrapper.setProps({ refreshToken: 1 });
    await flushPromises();
    await nextTick();

    expect(wrapper.emitted("visibility").at(-1)).toEqual([false]);
  });

  it("collapses parent rail visibility for unknown-date-only results", async () => {
    Photo.timeline.mockResolvedValue({
      buckets: [],
      unknownDateCount: 5,
    });

    const wrapper = mountTimeline();

    await flushPromises();
    await nextTick();

    expect(wrapper.find('[data-testid="timeline-unknown"]').exists()).toBe(false);
    expect(wrapper.emitted("visibility").at(-1)).toEqual([false]);
  });

  it("keeps stale timeline responses from overwriting newer bucket data", async () => {
    const first = deferred();
    const second = deferred();

    Photo.timeline.mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);

    const wrapper = mountTimeline({
      filter: { q: "first" },
    });

    await nextTick();
    await wrapper.setProps({ filter: { q: "second" } });

    second.resolve({
      buckets: [
        {
          key: "2026-02",
          label: "February 2026",
          year: 2026,
          month: 2,
          photoCount: 8,
        },
      ],
      unknownDateCount: 0,
    });

    await flushPromises();
    await nextTick();

    expect(wrapper.find('[data-testid="timeline-month-2026-2"]').exists()).toBe(true);

    first.resolve({
      buckets: [
        {
          key: "2024-09",
          label: "September 2024",
          year: 2024,
          month: 9,
          photoCount: 4,
        },
      ],
      unknownDateCount: 0,
    });

    await flushPromises();
    await nextTick();

    expect(wrapper.find('[data-testid="timeline-month-2026-2"]').exists()).toBe(true);
    expect(wrapper.find('[data-testid="timeline-month-2024-9"]').exists()).toBe(false);
  });
});
