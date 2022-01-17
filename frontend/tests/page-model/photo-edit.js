import { Selector, t } from "testcafe";
import { RequestLogger } from "testcafe";

const logger = RequestLogger(/http:\/\/localhost:2343\/api\/v1\/*/, {
  logResponseHeaders: true,
  logResponseBody: true,
});

export default class Page {
  constructor() {}

  // turn switch of --photo --edit
  async turnSwitchOff(type) {
    await t
      .click("#tab-info")
      .expect(
        Selector(".input-" + type + " input", { timeout: 8000 }).hasAttribute(
          "aria-checked",
          "true"
        )
      )
      .ok()
      .click(Selector(".input-" + type + " input"))
      .expect(
        Selector(".input-" + type + " input", { timeout: 8000 }).hasAttribute(
          "aria-checked",
          "false"
        )
      )
      .ok();
  }

  async turnSwitchOn(type) {
    await t
      .click("#tab-info")
      .expect(
        Selector(".input-" + type + " input", { timeout: 8000 }).hasAttribute(
          "aria-checked",
          "false"
        )
      )
      .ok()
      .click(Selector(".input-" + type + " input"))
      .expect(
        Selector(".input-" + type + " input", { timeout: 8000 }).hasAttribute(
          "aria-checked",
          "true"
        )
      )
      .ok();
  }

  //TODO refactor
  async checkEditFormValues(
    title,
    day,
    month,
    year,
    localTime,
    utcTime,
    timezone,
    country,
    altitude,
    lat,
    lng,
    camera,
    iso,
    exposure,
    lens,
    fnumber,
    flength,
    subject,
    artist,
    copyright,
    license,
    description,
    keywords,
    notes
  ) {
    if (title !== "") {
      await t.expect(Selector(".input-title input").value).eql(title);
    }
    if (day !== "") {
      await t.expect(Selector(".input-day input").value).eql(day);
    }
    if (month !== "") {
      await t.expect(Selector(".input-month input").value).eql(month);
    }
    if (year !== "") {
      await t.expect(Selector(".input-year input").value).eql(year);
    }
    if (timezone !== "") {
      await t.expect(Selector(".input-timezone input").value).eql(timezone);
    }
    if (localTime !== "") {
      await t.expect(Selector(".input-local-time input").value).eql(localTime);
    }
    if (utcTime !== "") {
      await t.expect(Selector(".input-utc-time input").value).eql(utcTime);
    }
    if (altitude !== "") {
      await t.expect(Selector(".input-altitude input").value).eql(altitude);
    }
    if (country !== "") {
      await t.expect(Selector("div").withText(country).visible).ok();
    }
    if (lat !== "") {
      await t.expect(Selector(".input-latitude input").value).eql(lat);
    }
    if (lng !== "") {
      await t.expect(Selector(".input-longitude input").value).eql(lng);
    }
    if (camera !== "") {
      await t.expect(Selector("div").withText(camera).visible).ok();
    }
    if (lens !== "") {
      await t.expect(Selector("div").withText(lens).visible).ok();
    }
    if (iso !== "") {
      await t.expect(Selector(".input-iso input").value).eql(iso);
    }
    if (exposure !== "") {
      await t.expect(Selector(".input-exposure input").value).eql(exposure);
    }
    if (fnumber !== "") {
      await t.expect(Selector(".input-fnumber input").value).eql(fnumber);
    }
    if (flength !== "") {
      await t.expect(Selector(".input-focal-length input").value).eql(flength);
    }
    if (subject !== "") {
      await t.expect(Selector(".input-subject textarea").value).eql(subject);
    }
    if (artist !== "") {
      await t.expect(Selector(".input-artist input").value).eql(artist);
    }
    if (copyright !== "") {
      await t.expect(Selector(".input-copyright input").value).eql(copyright);
    }
    if (license !== "") {
      await t.expect(Selector(".input-license textarea").value).eql(license);
    }
    if (description !== "") {
      await t.expect(Selector(".input-description textarea").value).eql(description);
    }
    if (notes !== "") {
      await t.expect(Selector(".input-notes textarea").value).contains(notes);
    }
    if (keywords !== "") {
      await t.expect(Selector(".input-keywords textarea").value).contains(keywords);
    }
  }

  async editPhoto(
    title,
    timezone,
    day,
    month,
    year,
    localTime,
    altitude,
    lat,
    lng,
    iso,
    exposure,
    fnumber,
    flength,
    subject,
    artist,
    copyright,
    license,
    description,
    keywords,
    notes
  ) {
    await t
      .typeText(Selector(".input-title input"), title, { replace: true })
      .typeText(Selector(".input-timezone input"), timezone, { replace: true })
      .click(Selector("div").withText(timezone).parent('div[role="listitem"]'))
      .typeText(Selector(".input-day input"), day, { replace: true })
      .pressKey("enter")
      .typeText(Selector(".input-month input"), month, { replace: true })
      .pressKey("enter")
      .typeText(Selector(".input-year input"), year, { replace: true })
      .click(Selector("div").withText(year).parent('div[role="listitem"]'))
      .click(Selector(".input-local-time input"))
      .pressKey("ctrl+a delete")
      .typeText(Selector(".input-local-time input"), localTime, { replace: true })
      .pressKey("enter")
      .typeText(Selector(".input-altitude input"), altitude, { replace: true })
      .typeText(Selector(".input-latitude input"), lat, { replace: true })
      .typeText(Selector(".input-longitude input"), lng, { replace: true })
      //.click(Selector('.input-camera input'))
      //.hover(Selector('div').withText('Apple iPhone 6').parent('div[role="listitem"]'))
      //.click(Selector('div').withText('Apple iPhone 6').parent('div[role="listitem"]'))
      //.click(Selector('.input-lens input'))
      //.click(Selector('div').withText('Apple iPhone 5s back camera 4.15mm f/2.2').parent('div[role="listitem"]'))
      .typeText(Selector(".input-iso input"), iso, { replace: true })
      .typeText(Selector(".input-exposure input"), exposure, { replace: true })
      .typeText(Selector(".input-fnumber input"), fnumber, { replace: true })
      .typeText(Selector(".input-focal-length input"), flength, { replace: true })
      .typeText(Selector(".input-subject textarea"), subject, { replace: true })
      .typeText(Selector(".input-artist input"), artist, { replace: true })
      .typeText(Selector(".input-copyright input"), copyright, { replace: true })
      .typeText(Selector(".input-license textarea"), license, { replace: true })
      .typeText(Selector(".input-description textarea"), description, {
        replace: true,
      })
      .typeText(Selector(".input-keywords textarea"), keywords)
      .typeText(Selector(".input-notes textarea"), notes, { replace: true })
      .click(Selector("button.action-approve"));
    await t.expect(Selector(".input-latitude input").visible, { timeout: 5000 }).ok();
    if (t.browser.platform === "mobile") {
      await t.click(Selector("button.action-apply")).click(Selector("button.action-close"));
    } else {
      await t.click(Selector("button.action-done", { timeout: 5000 }));
    }
  }

  async undoPhotoEdit(
    title,
    timezone,
    day,
    month,
    year,
    localTime,
    altitude,
    lat,
    lng,
    country,
    iso,
    exposure,
    fnumber,
    flength,
    subject,
    artist,
    copyright,
    license,
    description,
    keywords,
    notes
  ) {
    if (title.empty || title === "") {
      await t.click(Selector(".input-title input")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-title input"), title, { replace: true });
    }
    await t
      .typeText(Selector(".input-day input"), day, { replace: true })
      .pressKey("enter")
      .typeText(Selector(".input-month input"), month, { replace: true })
      .pressKey("enter")
      .typeText(Selector(".input-year input"), year, { replace: true })
      .pressKey("enter");
    if (localTime.empty || localTime === "") {
      await t.click(Selector(".input-local-time input")).pressKey("ctrl+a delete");
    } else {
      await t
        .click(Selector(".input-local-time input"))
        .pressKey("ctrl+a delete")
        .typeText(Selector(".input-local-time input"), localTime, { replace: true })
        .pressKey("enter");
    }
    if (timezone.empty || timezone === "") {
      await t
        .click(Selector(".input-timezone input"))
        .typeText(Selector(".input-timezone input"), "UTC", { replace: true })
        .pressKey("enter");
    } else {
      await t
        .typeText(Selector(".input-timezone input"), timezone, { replace: true })
        .pressKey("enter");
    }
    if (lat.empty || lat === "") {
      await t.click(Selector(".input-latitude input")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-latitude input"), lat, { replace: true });
    }
    if (lng.empty || lng === "") {
      await t.click(Selector(".input-longitude input")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-longitude input"), lng, { replace: true });
    }
    if (altitude.empty || altitude === "") {
      await t.click(Selector(".input-altitude input")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-altitude input"), altitude, { replace: true });
    }
    if (country.empty || country === "") {
      await t.click(Selector(".input-longitude input")).pressKey("ctrl+a delete");
    } else {
      await t
        .click(Selector(".input-country input"))
        .pressKey("ctrl+a delete")
        .typeText(Selector(".input-country input"), country, { replace: true })
        .pressKey("enter");
    }
    // if (FirstPhotoCamera.empty || FirstPhotoCamera === "")
    //{ await t
    //.click(Selector('.input-camera input'))
    // .hover(Selector('div').withText('Unknown').parent('div[role="listitem"]'))
    //  .click(Selector('div').withText('Unknown').parent('div[role="listitem"]'))}
    //else
    //{await t
    //  .click(Selector('.input-camera input'))
    //   .hover(Selector('div').withText(FirstPhotoCamera).parent('div[role="listitem"]'))
    //    .click(Selector('div').withText(FirstPhotoCamera).parent('div[role="listitem"]'))}
    //if (FirstPhotoLens.empty || FirstPhotoLens === "")
    //{ await t
    //  .click(Selector('.input-lens input'))
    //   .click(Selector('div').withText('Unknown').parent('div[role="listitem"]'))}
    //else
    //{await t
    //   .click(Selector('.input-lens input'))
    //    .click(Selector('div').withText(FirstPhotoLens).parent('div[role="listitem"]'))}
    if (iso.empty || iso === "") {
      await t.click(Selector(".input-iso input")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-iso input"), iso, { replace: true });
    }
    if (exposure.empty || exposure === "") {
      await t.click(Selector(".input-exposure input")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-exposure input"), exposure, { replace: true });
    }
    if (fnumber.empty || fnumber === "") {
      await t.click(Selector(".input-fnumber input")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-fnumber input"), fnumber, { replace: true });
    }
    if (flength.empty || flength === "") {
      await t.click(Selector(".input-focal-length input")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-focal-length input"), flength, {
        replace: true,
      });
    }
    if (subject.empty || subject === "") {
      await t.click(Selector(".input-subject textarea")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-subject textarea"), subject, { replace: true });
    }
    if (artist.empty || artist === "") {
      await t.click(Selector(".input-artist input")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-artist input"), artist, { replace: true });
    }
    if (copyright.empty || copyright === "") {
      await t.click(Selector(".input-copyright input")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-copyright input"), copyright, { replace: true });
    }
    if (license.empty || license === "") {
      await t.click(Selector(".input-license textarea")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-license textarea"), license, { replace: true });
    }
    if (description.empty || description === "") {
      await t.click(Selector(".input-description textarea")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-description textarea"), description, {
        replace: true,
      });
    }
    if (keywords.empty || keywords === "") {
      await t.click(Selector(".input-keywords textarea")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-keywords textarea"), keywords, { replace: true });
    }
    if (notes.empty || notes === "") {
      await t.click(Selector(".input-notes textarea")).pressKey("ctrl+a delete");
    } else {
      await t.typeText(Selector(".input-notes textarea"), notes, { replace: true });
    }
    if (t.browser.platform === "mobile") {
      await t.click(Selector("button.action-apply")).click(Selector("button.action-close"));
    } else {
      await t.click(Selector("button.action-done", { timeout: 5000 }));
    }
  }
}
