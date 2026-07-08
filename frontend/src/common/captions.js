import $util from "common/util";

/**
 * PhotoSwipe Dynamic Caption plugin v1.2.7
 * https://github.com/dimsemenov/photoswipe-dynamic-caption-plugin
 *
 * By https://dimsemenov.com
 */

const defaultOptions = {
  captionContent: ".pswp-caption-content",
  type: "below",
  horizontalEdgeThreshold: 20,
  mobileCaptionOverlapRatio: 0.3,
  mobileLayoutBreakpoint: 1024,
  verticallyCenterImage: false,
  // enabled gates the layout-shrinking work in onCalcSlideSize; the caption
  // element is still created so a later toggle-on resumes cleanly.
  enabled: () => true,
};

class PhotoSwipeDynamicCaption {
  constructor(lightbox, options) {
    this.options = {
      ...defaultOptions,
      ...options,
    };

    this.lightbox = lightbox;

    this.lightbox.on("init", () => {
      this.pswp = this.lightbox.pswp;
      this.initCaption();
    });
  }

  initCaption() {
    const { pswp } = this;

    pswp.on("change", () => {
      // Refresh first so a slide preloaded with stale Title/Caption picks
      // up any photos.updated event that landed before activation.
      this.refreshCaption(this.pswp.currSlide);
      // make sure caption is displayed after slides are switched
      this.showCaption(this.pswp.currSlide);
    });

    pswp.on("calcSlideSize", (e) => this.onCalcSlideSize(e));

    pswp.on("slideDestroy", (e) => {
      if (e.slide.dynamicCaption) {
        if (e.slide.dynamicCaption.element) {
          e.slide.dynamicCaption.element.remove();
        }
        delete e.slide.dynamicCaption;
      }
    });

    // hide caption if zoomed
    pswp.on("zoomPanUpdate", ({ slide }) => {
      if (pswp.opener.isOpen && slide.dynamicCaption) {
        if (slide.currZoomLevel > slide.zoomLevels.initial) {
          this.hideCaption(slide);
        } else {
          this.showCaption(slide);
        }

        // move caption on vertical drag
        if (slide.dynamicCaption.element) {
          let captionYOffset = 0;
          if (slide.currZoomLevel <= slide.zoomLevels.initial) {
            const shiftedAmount = slide.pan.y - slide.bounds.center.y;
            if (Math.abs(shiftedAmount) > 1) {
              captionYOffset = shiftedAmount;
            }
          }

          this.setCaptionYOffset(slide.dynamicCaption.element, captionYOffset);
        }

        this.adjustPanArea(slide, slide.currZoomLevel);
      }
    });

    pswp.on("beforeZoomTo", (e) => {
      this.adjustPanArea(pswp.currSlide, e.destZoomLevel);
    });

    // Stop default action of tap when tapping on the caption
    pswp.on("tapAction", (e) => {
      if (e.originalEvent.target.closest(".pswp__dynamic-caption")) {
        e.preventDefault();
      }
    });
  }

  adjustPanArea(slide, zoomLevel) {
    if (slide.dynamicCaption && slide.dynamicCaption.adjustedPanAreaSize) {
      if (zoomLevel > slide.zoomLevels.initial) {
        slide.panAreaSize.x = slide.dynamicCaption.originalPanAreaSize.x;
        slide.panAreaSize.y = slide.dynamicCaption.originalPanAreaSize.y;
      } else {
        // Restore panAreaSize after we zoom back to initial position
        slide.panAreaSize.x = slide.dynamicCaption.adjustedPanAreaSize.x;
        slide.panAreaSize.y = slide.dynamicCaption.adjustedPanAreaSize.y;
      }
    }
  }

  useMobileLayout() {
    const { mobileLayoutBreakpoint } = this.options;

    if (typeof mobileLayoutBreakpoint === "function") {
      return mobileLayoutBreakpoint.call(this);
    } else if (typeof mobileLayoutBreakpoint === "number") {
      if (window.innerWidth < mobileLayoutBreakpoint) {
        return true;
      }
    }

    return false;
  }

  hideCaption(slide) {
    if (slide.dynamicCaption && !slide.dynamicCaption.hidden) {
      const captionElement = slide.dynamicCaption.element;

      if (!captionElement) {
        return;
      }

      slide.dynamicCaption.hidden = true;
      captionElement.classList.add("pswp__dynamic-caption--faded");

      // Disable caption visibility with the delay, so it's not interactable
      if (slide.captionFadeTimeout) {
        clearTimeout(slide.captionFadeTimeout);
      }

      slide.captionFadeTimeout = setTimeout(() => {
        if (captionElement) {
          captionElement.style.visibility = "hidden";
        }
        delete slide.captionFadeTimeout;
      }, 400);
    }
  }

  setCaptionYOffset(el, y) {
    el.style.transform = `translateY(${y}px)`;
  }

  showCaption(slide) {
    if (slide.dynamicCaption && slide.dynamicCaption.hidden) {
      const captionElement = slide.dynamicCaption.element;

      if (!captionElement) {
        return;
      }

      slide.dynamicCaption.hidden = false;
      captionElement.style.visibility = "visible";

      if (slide.captionFadeTimeout) {
        clearTimeout(slide.captionFadeTimeout);
      }

      slide.captionFadeTimeout = setTimeout(() => {
        if (captionElement) {
          captionElement.classList.remove("pswp__dynamic-caption--faded");
        }
        delete slide.captionFadeTimeout;
      }, 50);
    }
  }

  setCaptionPosition(captionEl, x, y) {
    const isOnHorizontalEdge = x <= this.options.horizontalEdgeThreshold;
    captionEl.classList[isOnHorizontalEdge ? "add" : "remove"]("pswp__dynamic-caption--on-hor-edge");

    if (document.dir === "rtl") {
      captionEl.style.right = x + "px";
    } else {
      captionEl.style.left = x + "px";
    }

    captionEl.style.top = y + "px";
  }

  setCaptionWidth(captionEl, width) {
    if (!width) {
      captionEl.style.removeProperty("width");
    } else {
      captionEl.style.width = width + "px";
    }
  }

  setCaptionType(captionEl, type) {
    const prevType = captionEl.dataset.pswpCaptionType;
    if (type !== prevType) {
      captionEl.classList.add("pswp__dynamic-caption--" + type);
      captionEl.classList.remove("pswp__dynamic-caption--" + prevType);
      captionEl.dataset.pswpCaptionType = type;
    }
  }

  updateCaptionPosition(slide) {
    if (!slide.dynamicCaption || !slide.dynamicCaption.type || !slide.dynamicCaption.element) {
      return;
    }

    if (slide.dynamicCaption.type === "mobile") {
      this.setCaptionType(slide.dynamicCaption.element, slide.dynamicCaption.type);

      if (document.dir === "rtl") {
        slide.dynamicCaption.element.style.removeProperty("right");
      } else {
        slide.dynamicCaption.element.style.removeProperty("left");
      }

      slide.dynamicCaption.element.style.removeProperty("top");
      this.setCaptionWidth(slide.dynamicCaption.element, false);
      return;
    }

    const zoomLevel = slide.zoomLevels.initial;
    const imageWidth = Math.ceil(slide.width * zoomLevel);
    const imageHeight = Math.ceil(slide.height * zoomLevel);

    this.setCaptionType(slide.dynamicCaption.element, slide.dynamicCaption.type);
    if (slide.dynamicCaption.type === "aside") {
      this.setCaptionPosition(slide.dynamicCaption.element, slide.bounds.center.x + imageWidth, slide.bounds.center.y);
      this.setCaptionWidth(slide.dynamicCaption.element, false);
    } else if (slide.dynamicCaption.type === "below") {
      this.setCaptionPosition(slide.dynamicCaption.element, slide.bounds.center.x, slide.bounds.center.y + imageHeight);
      this.setCaptionWidth(slide.dynamicCaption.element, imageWidth);
    }
  }

  onCalcSlideSize(e) {
    const { slide } = e;
    let captionSize;
    let useMobileVersion;

    if (!slide.dynamicCaption) {
      slide.dynamicCaption = {
        element: undefined,
        type: false,
        hidden: false,
      };
    }

    // Recompute on every layout pass so edits made via the sidebar (or
    // photos.updated events from other clients) surface when the caption
    // becomes visible again.
    this.refreshCaption(slide);

    if (!slide.dynamicCaption.element) {
      return;
    }

    this.storeOriginalPanAreaSize(slide);

    // Refresh bounds against the current zoomLevels.initial so adjustPanArea
    // positions the photo inside the new box (bounds is this plugin's job).
    slide.bounds.update(slide.zoomLevels.initial);

    // External gate (e.g. the lightbox's Ctrl+H caption toggle): when disabled,
    // skip the panAreaSize shrink so the photo fills the viewport. Drop any
    // stale adjustedPanAreaSize so adjustPanArea is a no-op until re-enabled.
    if (typeof this.options.enabled === "function" && !this.options.enabled()) {
      delete slide.dynamicCaption.adjustedPanAreaSize;
      return;
    }

    if (this.useMobileLayout()) {
      slide.dynamicCaption.type = "mobile";
      useMobileVersion = true;
    } else {
      if (this.options.type === "auto") {
        if (slide.bounds.center.x > slide.bounds.center.y) {
          slide.dynamicCaption.type = "aside";
        } else {
          slide.dynamicCaption.type = "below";
        }
      } else {
        slide.dynamicCaption.type = this.options.type;
      }
    }

    const imageWidth = Math.ceil(slide.width * slide.zoomLevels.initial);
    const imageHeight = Math.ceil(slide.height * slide.zoomLevels.initial);

    this.setCaptionType(slide.dynamicCaption.element, slide.dynamicCaption.type);

    if (slide.dynamicCaption.type === "aside") {
      this.setCaptionWidth(slide.dynamicCaption.element, false);
      captionSize = this.measureCaptionSize(slide.dynamicCaption.element, e.slide);

      const captionWidth = captionSize.x;

      const horizontalEnding = imageWidth + slide.bounds.center.x;
      const horizontalLeftover = slide.panAreaSize.x - horizontalEnding;

      if (horizontalLeftover <= captionWidth) {
        slide.panAreaSize.x -= captionWidth;
        this.recalculateZoomLevelAndBounds(slide);
      } else {
        // do nothing, caption will fit aside without any adjustments
      }
    } else if (slide.dynamicCaption.type === "below" || useMobileVersion) {
      this.setCaptionWidth(slide.dynamicCaption.element, useMobileVersion ? this.pswp.viewportSize.x : imageWidth);

      captionSize = this.measureCaptionSize(slide.dynamicCaption.element, e.slide);
      const captionHeight = captionSize.y;

      if (this.options.verticallyCenterImage) {
        slide.panAreaSize.y -= captionHeight;
        this.recalculateZoomLevelAndBounds(slide);
      } else {
        // Lift the image by the height of the caption only.

        // Get vertical ending of the image.
        const verticalEnding = imageHeight + slide.bounds.center.y;

        // Get height between bottom of the screen and ending of the image,
        // before any adjustments applied.
        const verticalLeftover = slide.panAreaSize.y - verticalEnding;
        const initialPanAreaHeight = slide.panAreaSize.y;

        if (verticalLeftover <= captionHeight) {
          // Lift the image to make more room for the caption.
          slide.panAreaSize.y -= Math.min((captionHeight - verticalLeftover) * 2, captionHeight);

          // we reduce viewport size, thus we need to update zoom level and pan bounds
          this.recalculateZoomLevelAndBounds(slide);

          const maxPositionX = (slide.panAreaSize.x * this.options.mobileCaptionOverlapRatio) / 2;

          // Do not reduce viewport height if too few space available
          if (useMobileVersion && slide.bounds.center.x > maxPositionX) {
            // Restore the default position
            slide.panAreaSize.y = initialPanAreaHeight;
            this.recalculateZoomLevelAndBounds(slide);
          }
        }
      }
    } else {
      // mobile
    }

    this.storeAdjustedPanAreaSize(slide);
    this.updateCaptionPosition(slide);
  }

  measureCaptionSize(captionEl, slide) {
    const rect = captionEl.getBoundingClientRect();
    const event = this.pswp.dispatch("dynamicCaptionMeasureSize", {
      captionEl,
      slide,
      captionSize: {
        x: rect.width,
        y: rect.height,
      },
    });
    return event.captionSize;
  }

  recalculateZoomLevelAndBounds(slide) {
    slide.zoomLevels.update(slide.width, slide.height, slide.panAreaSize);
    slide.bounds.update(slide.zoomLevels.initial);
  }

  storeAdjustedPanAreaSize(slide) {
    if (slide.dynamicCaption) {
      if (!slide.dynamicCaption.adjustedPanAreaSize) {
        slide.dynamicCaption.adjustedPanAreaSize = {};
      }
      slide.dynamicCaption.adjustedPanAreaSize.x = slide.panAreaSize.x;
      slide.dynamicCaption.adjustedPanAreaSize.y = slide.panAreaSize.y;
    }
  }

  storeOriginalPanAreaSize(slide) {
    if (slide.dynamicCaption) {
      if (!slide.dynamicCaption.originalPanAreaSize) {
        slide.dynamicCaption.originalPanAreaSize = {};
      }
      slide.dynamicCaption.originalPanAreaSize.x = slide.panAreaSize.x;
      slide.dynamicCaption.originalPanAreaSize.y = slide.panAreaSize.y;
    }
  }

  // refreshCaption rebuilds the slide's caption HTML in place so the lightbox
  // picks up edits applied through the sidebar or pushed via photos.updated.
  // Creates the element lazily on first non-empty HTML; once created, the
  // element stays in the DOM and is just updated, so the layout adjustments
  // in onCalcSlideSize can keep their pan-area bookkeeping consistent.
  refreshCaption(slide) {
    if (!slide || !slide.dynamicCaption) {
      return;
    }

    const captionHTML = this.getCaptionHTML(slide);

    if (!slide.dynamicCaption.element) {
      if (!captionHTML) {
        return;
      }
      slide.dynamicCaption.element = document.createElement("div");
      slide.dynamicCaption.element.className = "pswp__dynamic-caption pswp__hide-on-close";
      slide.dynamicCaption.element.innerHTML = captionHTML; // security-reviewed: captionHTML is sanitized or HTML-encoded.
      slide.holderElement?.appendChild(slide.dynamicCaption.element);
      this.pswp.dispatch("dynamicCaptionUpdateHTML", {
        captionElement: slide.dynamicCaption.element,
        slide,
      });
      return;
    }

    if (slide.dynamicCaption.element.innerHTML === captionHTML) {
      return;
    }

    slide.dynamicCaption.element.innerHTML = captionHTML; // security-reviewed: captionHTML is sanitized or HTML-encoded.
    this.pswp.dispatch("dynamicCaptionUpdateHTML", {
      captionElement: slide.dynamicCaption.element,
      slide,
    });
  }

  // refreshCurrentCaption refreshes the caption for the slide currently shown.
  refreshCurrentCaption() {
    if (this.pswp && this.pswp.currSlide) {
      this.refreshCaption(this.pswp.currSlide);
    }
  }

  getCaptionHTML(slide) {
    if (!slide) {
      return "";
    }

    // Prefer a model-aware lookup when the host provides one — the plugin
    // owns the HTML format so callers don't reimplement encoding/sanitizing.
    if (typeof this.options.getModel === "function") {
      const model = this.options.getModel.call(this, slide);
      if (model) {
        return this.formatCaption(model);
      }
    }

    if (typeof this.options.captionContent === "function") {
      return this.options.captionContent.call(this, slide);
    }

    const currSlideElement = slide.data?.element;
    let captionHTML = "";
    if (currSlideElement) {
      const hiddenCaption = currSlideElement.querySelector(this.options.captionContent);
      if (hiddenCaption) {
        // get caption from element with class pswp-caption-content
        captionHTML = $util.sanitizeHtml(hiddenCaption.innerHTML);
      } else {
        const img = currSlideElement.querySelector("img");
        if (img) {
          // get caption from alt attribute
          captionHTML = $util.encodeHTML(img.getAttribute("alt") || "");
        }
      }
    }
    return captionHTML;
  }

  // formatCaption returns the PhotoPrism caption HTML (sanitized) for a slide
  // model. Title renders as <h4>; caption text renders as <h4> when it stands
  // alone on a single line, otherwise as <p>. Falls back to Description when
  // Caption is empty. Pure: does not mutate the model.
  formatCaption(model) {
    if (!model) {
      return "";
    }

    let html = "";

    if (model.Title) {
      html += `<h4>${$util.encodeHTML(model.Title.trim())}</h4>`;
    }

    let text = typeof model.Caption === "string" ? model.Caption.trim() : "";
    if (!text && typeof model.Description === "string") {
      text = model.Description.trim();
    }

    if (text) {
      if (!html && text.split("\n").length < 2) {
        // Single-line, no title — promote to <h4> so it reads as the headline.
        html += `<h4>${$util.encodeHTML(text)}</h4>`;
      } else {
        html += `<p>${$util.encodeHTML(text)}</p>`;
      }
    }

    return $util.sanitizeHtml(html);
  }
}

export default PhotoSwipeDynamicCaption;
