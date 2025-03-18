/**
 * PhotoSwipe Dynamic Caption plugin v1.2.7
 * https://github.com/dimsemenov/photoswipe-dynamic-caption-plugin
 *
 * By https://dimsemenov.com
 */

const defaultOptions = {
  captionContent: ".pswp-caption-content",
  type: "auto",
  horizontalEdgeThreshold: 20,
  mobileCaptionOverlapRatio: 0.3,
  mobileLayoutBreakpoint: 600,
  verticallyCenterImage: false,
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

      const captionHTML = this.getCaptionHTML(slide);

      if (!captionHTML) {
        return;
      }

      slide.dynamicCaption.element = document.createElement("div");
      slide.dynamicCaption.element.className = "pswp__dynamic-caption pswp__hide-on-close";
      slide.dynamicCaption.element.innerHTML = captionHTML;

      this.pswp.dispatch("dynamicCaptionUpdateHTML", {
        captionElement: slide.dynamicCaption.element,
        slide,
      });

      slide.holderElement.appendChild(slide.dynamicCaption.element);
    }

    if (!slide.dynamicCaption.element) {
      return;
    }

    this.storeOriginalPanAreaSize(slide);

    slide.bounds.update(slide.zoomLevels.initial);

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

  getCaptionHTML(slide) {
    if (typeof this.options.captionContent === "function") {
      return this.options.captionContent.call(this, slide);
    }

    const currSlideElement = slide.data.element;
    let captionHTML = "";
    if (currSlideElement) {
      const hiddenCaption = currSlideElement.querySelector(this.options.captionContent);
      if (hiddenCaption) {
        // get caption from element with class pswp-caption-content
        captionHTML = hiddenCaption.innerHTML;
      } else {
        const img = currSlideElement.querySelector("img");
        if (img) {
          // get caption from alt attribute
          captionHTML = img.getAttribute("alt");
        }
      }
    }
    return captionHTML;
  }
}

export default PhotoSwipeDynamicCaption;
