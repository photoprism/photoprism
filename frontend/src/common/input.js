/*

Copyright (c) 2018 - 2024 PhotoPrism UG. All rights reserved.

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

// import log from "common/log";
export const InputInvalid = 0;
export const ClickShort = 1;
export const ClickLong = 2;

export class Input {
  constructor() {
    this.reset();
    this.preferTouch = false;
  }

  reset() {
    // log.debug("input.reset");
    this.index = -1;
    this.scrollY = window.scrollY;
    this.touches = [];
    this.timeStamp = -1;
  }

  touchStart(ev, index) {
    // log.debug("input.touchStart", [ev, ev.type, ev.touches, index]);
    this.preferTouch = true;
    this.index = index;
    this.scrollY = window.scrollY;
    if (ev.touches) {
      this.touches = ev.touches;
    }
    this.timeStamp = ev.timeStamp;
  }

  mouseDown(ev, index) {
    if (this.preferTouch) {
      return; // Ignore "mousedown" event on touch devices.
    }
    // log.debug("input.mouseDown", [ev, ev.type, ev.touches, index]);
    this.index = index;
    this.scrollY = window.scrollY;
    this.touches = [];
    this.timeStamp = ev.timeStamp;
  }

  clickType(ev, index) {
    // log.debug("input.clickType", [ev, ev.type, index]);
    if (this.timeStamp < 0) {
      return InputInvalid;
    }

    if (ev.changedTouches && ev.changedTouches.length === 1) {
      if (this.touches.length !== 1) {
        return InputInvalid;
      }

      if (Math.abs(this.touches[0].screenX - ev.changedTouches[0].screenX) > 4 || Math.abs(this.touches[0].screenY - ev.changedTouches[0].screenY) > 4) {
        return InputInvalid;
      }
    }

    if (this.index !== index || this.scrollY - window.scrollY !== 0) {
      return InputInvalid;
    }

    const clickDuration = ev.timeStamp - this.timeStamp;

    if (clickDuration > 0 && clickDuration < 333) {
      return ClickShort;
    } else if (clickDuration >= 333) {
      return ClickLong;
    }

    return InputInvalid;
  }

  eval(ev, index) {
    // log.debug("input.eval", [ev, ev.type, index]);
    const result = this.clickType(ev, index);
    this.reset();
    return result;
  }
}

export default Input;
