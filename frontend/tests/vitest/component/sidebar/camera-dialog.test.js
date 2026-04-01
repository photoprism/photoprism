import { describe, it, expect, vi } from "vitest";
import { mount } from "@vue/test-utils";
import PCameraDialog from "component/sidebar/camera-dialog.vue";

describe("PCameraDialog component", () => {
  const mockPhoto = {
    CameraID: 5,
    LensID: 3,
    Iso: "200",
    Exposure: "1/250",
    FNumber: "2.8",
    FocalLength: "50",
  };

  it("should load values from photo via loadFromPhoto", () => {
    const w = mount(PCameraDialog, {
      props: { visible: false, photo: mockPhoto },
    });

    expect(w.vm.cameraID).toBe(0);

    w.vm.loadFromPhoto();

    expect(w.vm.cameraID).toBe(5);
    expect(w.vm.lensID).toBe(3);
    expect(w.vm.iso).toBe("200");
    expect(w.vm.exposure).toBe("1/250");
    expect(w.vm.fNumber).toBe("2.8");
    expect(w.vm.focalLength).toBe("50");
  });

  it("should handle null photo gracefully", () => {
    const w = mount(PCameraDialog, {
      props: { visible: false, photo: null },
    });

    w.vm.loadFromPhoto();

    expect(w.vm.cameraID).toBe(0);
    expect(w.vm.lensID).toBe(0);
    expect(w.vm.iso).toBe("");
  });

  it("should emit close event", () => {
    const onClose = vi.fn();
    const w = mount(PCameraDialog, {
      props: { visible: false, photo: mockPhoto, onClose },
    });
    w.vm.close();
    expect(onClose).toHaveBeenCalledOnce();
  });

  it("should emit confirm with edited values", () => {
    const onConfirm = vi.fn();
    const w = mount(PCameraDialog, {
      props: { visible: false, photo: mockPhoto, onConfirm },
    });

    w.vm.loadFromPhoto();

    w.vm.cameraID = 10;
    w.vm.lensID = 7;
    w.vm.iso = "400";
    w.vm.exposure = "1/125";
    w.vm.fNumber = "1.4";
    w.vm.focalLength = "35";

    w.vm.confirm();

    expect(onConfirm).toHaveBeenCalledOnce();
    expect(onConfirm).toHaveBeenCalledWith({
      CameraID: 10,
      LensID: 7,
      Iso: "400",
      Exposure: "1/125",
      FNumber: "1.4",
      FocalLength: "35",
    });
  });

  it("should use defaults for missing photo fields", () => {
    const w = mount(PCameraDialog, {
      props: { visible: false, photo: { CameraID: 2 } },
    });

    w.vm.loadFromPhoto();

    expect(w.vm.cameraID).toBe(2);
    expect(w.vm.lensID).toBe(0);
    expect(w.vm.iso).toBe("");
    expect(w.vm.exposure).toBe("");
  });
});
