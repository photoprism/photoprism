package encode

// MovFlags defines default fragmented MP4 flags suitable for streaming:
// - https://developer.mozilla.org/en-US/docs/Web/API/Media_Source_Extensions_API/Transcoding_assets_for_MSE#fragmenting
// - https://nschlia.github.io/ffmpegfs/html/ffmpeg__profiles_8cc.html
// - https://cloudinary.com/glossary/fragmented-mp4
// - https://medium.com/@vlad.pbr/in-browser-live-video-using-fragmented-mp4-3aedb600a07e
// - https://github.com/video-dev/hls.js?tab=readme-ov-file#features
var MovFlags = "use_metadata_tags+faststart"
