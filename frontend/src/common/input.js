/*

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.app>

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

    PhotoPrism® is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.app/developer-guide/

*/

export const InputInvalid = 0;
export const ClickShort = 1;
export const ClickLong = 2;

export class Input {
  constructor() {
    this.reset();
  }

  reset() {
    this.index = -1;
    this.scrollY = window.scrollY;
    this.touches = [];
    this.timeStamp = -1;
  }

  touchStart(ev, index) {
    this.index = index;
    this.scrollY = window.scrollY;
    if (ev.touches) {
      this.touches = ev.touches;
    }
    this.timeStamp = ev.timeStamp;
  }

  mouseDown(ev, index) {
    this.index = index;
    this.scrollY = window.scrollY;
    this.touches = [];
    this.timeStamp = ev.timeStamp;
  }

  clickType(ev, index) {
    if (this.timeStamp < 0) {
      return InputInvalid;
    }

    if (ev.changedTouches && ev.changedTouches.length === 1) {
      if (this.touches.length !== 1) {
        return InputInvalid;
      }

      if (
        Math.abs(this.touches[0].screenX - ev.changedTouches[0].screenX) > 4 ||
        Math.abs(this.touches[0].screenY - ev.changedTouches[0].screenY) > 4
      ) {
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
    const result = this.clickType(ev, index);
    this.reset();
    return result;
  }
}

export default Input;
