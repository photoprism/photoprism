package fs

import (
	"path/filepath"
	"strings"
)

// FileExtensions maps file extensions to standard formats
type FileExtensions map[string]Type

// Extensions contains the filename extensions of file formats known to PhotoPrism.
var Extensions = FileExtensions{
	ExtZip:      ArchiveZip,
	ExtPDF:      DocumentPDF, // .pdf
	ExtJpeg:     ImageJpeg,   // .jpg
	".jpeg":     ImageJpeg,
	".jpe":      ImageJpeg,
	".jif":      ImageJpeg,
	".jfif":     ImageJpeg,
	".jfi":      ImageJpeg,
	".jxl":      ImageJpegXL,
	ExtThm:      ImageThumb,
	".tif":      ImageTiff,
	".tiff":     ImageTiff,
	".psd":      ImagePsd,
	ExtPng:      ImagePng, // .png
	".apng":     ImagePng,
	".pnga":     ImagePng,
	".pn":       ImagePng,
	".gif":      ImageGif,
	".bmp":      ImageBmp,
	ExtDng:      ImageDng, // .dng
	".avif":     ImageAvif,
	".avis":     ImageAvifS,
	".avifs":    ImageAvifS,
	".hif":      ImageHeic,
	".heif":     ImageHeic,
	".heic":     ImageHeic,
	".avci":     ImageHeic,
	".avcs":     ImageHeic,
	".heifs":    ImageHeicS,
	".heics":    ImageHeicS,
	".webp":     ImageWebp,
	".mpo":      ImageMPO,
	".3fr":      ImageRaw,
	".ari":      ImageRaw,
	".arw":      ImageRaw,
	".bay":      ImageRaw,
	".cap":      ImageRaw,
	".crw":      ImageRaw,
	".cr2":      ImageRaw,
	".cr3":      ImageRaw,
	".data":     ImageRaw,
	".dcs":      ImageRaw,
	".dcr":      ImageRaw,
	".drf":      ImageRaw,
	".eip":      ImageRaw,
	".erf":      ImageRaw,
	".fff":      ImageRaw,
	".gpr":      ImageRaw,
	".iiq":      ImageRaw,
	".k25":      ImageRaw,
	".kdc":      ImageRaw,
	".mdc":      ImageRaw,
	".mef":      ImageRaw,
	".mos":      ImageRaw,
	".mrw":      ImageRaw,
	".nef":      ImageRaw,
	".nrw":      ImageRaw,
	".obm":      ImageRaw,
	".orf":      ImageRaw,
	".pef":      ImageRaw,
	".ptx":      ImageRaw,
	".pxn":      ImageRaw,
	".r3d":      ImageRaw,
	".raf":      ImageRaw,
	".raw":      ImageRaw,
	".rwl":      ImageRaw,
	".rwz":      ImageRaw,
	".rw2":      ImageRaw,
	".srf":      ImageRaw,
	".srw":      ImageRaw,
	".sr2":      ImageRaw,
	".x3f":      ImageRaw,
	ExtXMP:      SidecarXMP,
	".aae":      SidecarAppleXml,
	ExtXml:      SidecarXml,
	ExtYaml:     SidecarYaml, // .yml
	".yaml":     SidecarYaml,
	ExtJson:     SidecarJson,
	ExtTxt:      SidecarText,
	".nfo":      SidecarInfo,
	ExtMd:       SidecarMarkdown,
	".markdown": SidecarMarkdown,
	".svg":      VectorSVG,
	".ai":       VectorAI,
	".ps":       VectorPS,
	".ps2":      VectorPS,
	".ps3":      VectorPS,
	".eps":      VectorEPS,
	".eps2":     VectorEPS,
	".eps3":     VectorEPS,
	".epi":      VectorEPS,
	".ept":      VectorEPS,
	".epsf":     VectorEPS,
	".epsi":     VectorEPS,
	ExtMov:      VideoMov, // Apple QuickTime Video Container
	ExtQT:       VideoMov, //  .qt
	ExtMp4:      VideoMp4, // MPEG-4 Part 14 Multimedia Container
	ExtH264:     VideoAvc, // ↓ H.264 MPEG-4 Advanced Video Coding (AVC)
	ExtAvc:      VideoAvc, //  .avc
	ExtAvc1:     VideoAvc, //  .avc1
	ExtDva:      VideoAvc, //  .dva
	ExtDva1:     VideoAvc, //  .dva1
	ExtAvc2:     VideoAvc, //  .avc2
	ExtAvc3:     VideoAvc, //  .avc3
	ExtDvav:     VideoAvc, //  .avc3
	ExtAvc10:    VideoAvc, //  .avc10
	ExtH265:     VideoHvc, // ↓ H.265 MPEG-4 HEVC with parameter sets only in the Sample Entry
	ExtHvc:      VideoHvc, //  .hvc
	ExtHvc1:     VideoHvc, //  .hvc1
	ExtDvh:      VideoHvc, //  .dvh
	ExtDvh1:     VideoHvc, //  .dvh1
	ExtHvc2:     VideoHvc, //  .hvc2
	ExtHvc3:     VideoHvc, //  .hvc3
	ExtHvc10:    VideoHvc, //  .hvc10
	ExtHevc:     VideoHvc, //  .hevc
	ExtHevc10:   VideoHvc, //  .hevc10
	ExtHev:      VideoHev, // ↓ H.265 video with parameter sets also in the Samples
	ExtHev1:     VideoHev, //  .hev1
	ExtDvhe:     VideoHev, //  .dvhe
	ExtHev2:     VideoHev, //  .hev2
	ExtHev3:     VideoHev, //  .hev3
	ExtHev10:    VideoHev, //  .hev10
	ExtH266:     VideoVvc, // ↓ H.266 MPEG-4 Versatile Video Coding (VVC)
	ExtVvc:      VideoVvc, //  .vvc
	ExtVvc1:     VideoVvc, //  .vvc1
	ExtEvc:      VideoEvc, // ↓ MPEG-5 Essential Video Coding (EVC)
	ExtEvc1:     VideoEvc, //  .evc1
	".vp8":      VideoVp8,
	".vp9":      VideoVp9,
	".av1":      VideoAv1,
	".av01":     VideoAv1,
	".mpg":      VideoMpeg,
	".mpeg":     VideoMpeg,
	".mjpg":     VideoMjpeg,
	".mjpeg":    VideoMjpeg,
	".mp2":      VideoMp2,
	".mpv":      VideoMp2,
	".mp":       VideoMp4,
	".m4v":      VideoM4v,
	".mxf":      VideoMXF,
	".3gp":      Video3GP,
	".3g2":      Video3G2,
	".flv":      VideoFlash,
	".f4v":      VideoFlash,
	".mkv":      VideoMkv,
	".ts":       VideoM2TS,
	".m2t":      VideoM2TS,
	".m2ts":     VideoM2TS,
	".mp2t":     VideoM2TS,
	".mts":      VideoAVCHD,
	".ogv":      VideoTheora,
	".ogg":      VideoTheora,
	".ogx":      VideoTheora,
	".webm":     VideoWebm,
	".asf":      VideoASF,
	".avi":      VideoAVI,
	".wmv":      VideoWMV,
	".dv":       VideoDV,
}

// Known tests if the file extension is known (supported).
func (m FileExtensions) Known(name string) bool {
	if name == "" {
		return false
	}

	ext := strings.ToLower(filepath.Ext(name))

	if ext == "" {
		return false
	}

	if _, ok := m[ext]; ok {
		return true
	}

	return false
}

// Types returns known extensions by file type.
func (m FileExtensions) Types(noUppercase bool) TypesExt {
	result := make(TypesExt)

	if noUppercase {
		for ext, t := range m {
			if _, ok := result[t]; ok {
				result[t] = append(result[t], ext)
			} else {
				result[t] = []string{ext}
			}
		}
	} else {
		for ext, t := range m {
			extUpper := strings.ToUpper(ext)
			if _, ok := result[t]; ok {
				result[t] = append(result[t], ext, extUpper)
			} else {
				result[t] = []string{ext, extUpper}
			}
		}
	}

	return result
}
