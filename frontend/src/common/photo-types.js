import { PhotoTypes } from "options/options";

// photoTypeOptions limits the shared media types to the page's search scope.
export function photoTypeOptions(staticFilter = {}, query = "") {
  // enabled recognizes affirmative search filter values.
  const enabled = (value) => value === true || value === "true" || value === 1 || value === "1";
  const fixedSearch = typeof query === "string" ? query.trim().toLowerCase() : "";
  if (
    staticFilter.type ||
    ["video", "live", "animated", "audio", "vector", "document", "raw"].some((key) => enabled(staticFilter[key])) ||
    ["vectors", "documents"].includes(fixedSearch)
  ) {
    return [];
  }
  let allowed;
  if (enabled(staticFilter.photo) || fixedSearch === "mono:true quality:3 photo:true") {
    allowed = ["image", "live", "raw"];
  } else if (enabled(staticFilter.media)) {
    allowed = ["live", "video", "audio", "animated"];
  }
  return PhotoTypes().filter((item) => !allowed || allowed.includes(item.value));
}
