import { PhotoTypes } from "options/options";

// photoTypeOptions limits the shared media types to the page's search scope.
export function photoTypeOptions(staticFilter = {}, query = "") {
  // enabled recognizes affirmative search filter values.
  const enabled = (value) => value === true || value === "true" || value === 1 || value === "1";
  const terms = typeof query === "string" ? query.toLowerCase().match(/"[^"\n]*"|'[^'\n]*'|\S+/g) || [] : [];
  if (
    staticFilter.type ||
    ["video", "live", "animated", "audio", "vector", "document", "raw"].some((key) => enabled(staticFilter[key])) ||
    terms.some((term) => /^(vectors?|documents?|videos?|audio|sounds|live|animated|gifs?|raws?)$/.test(term) || /^(type:.+|(vector|document|video|audio|live|animated|raw):true)$/.test(term))
  ) {
    return [];
  }
  let allowed;
  if (enabled(staticFilter.photo) || terms.some((term) => ["photo:true", "monochrome", "mono", "mono:true"].includes(term))) {
    allowed = ["image", "live", "raw"];
  } else if (enabled(staticFilter.media) || terms.includes("media:true")) {
    allowed = ["live", "video", "animated"];
  }
  return PhotoTypes().filter((item) => !allowed || allowed.includes(item.value));
}
