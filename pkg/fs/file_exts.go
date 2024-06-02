package fs

import (
	"path/filepath"
	"strings"
)

// FileExtensions maps file extensions to standard formats
type FileExtensions map[string]Type

// Extensions contains the filename extensions of file formats known to PhotoPrism.
var Extensions = FileExtensions{
	ExtJPEG:     ImageJPEG, // .jpg
	".jpeg":     ImageJPEG,
	".jpe":      ImageJPEG,
	".jif":      ImageJPEG,
	".jfif":     ImageJPEG,
	".jfi":      ImageJPEG,
	".jxl":      ImageJPEGXL,
	ExtTHM:      ImageThumb,
	".tif":      ImageTIFF,
	".tiff":     ImageTIFF,
	".psd":      ImagePSD,
	ExtPNG:      ImagePNG, // .png
	".apng":     ImagePNG,
	".pnga":     ImagePNG,
	".pn":       ImagePNG,
	".gif":      ImageGIF,
	".bmp":      ImageBMP,
	ExtDNG:      ImageDNG, // .dng
	".avif":     ImageAVIF,
	".avis":     ImageAVIFS,
	".avifs":    ImageAVIFS,
	".hif":      ImageHEIC,
	".heif":     ImageHEIC,
	".heic":     ImageHEIC,
	".avci":     ImageHEIC,
	".avcs":     ImageHEIC,
	".heifs":    ImageHEICS,
	".heics":    ImageHEICS,
	".webp":     ImageWebP,
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
	ExtAVC:      VideoAVC,  // .avc
	ExtHEVC:     VideoHEVC, // .hevc
	ExtVVC:      VideoVVC,  // .vvc
	ExtEVC:      VideoEVC,  // .evc
	".mov":      VideoMOV,
	".qt":       VideoMOV,
	".avi":      VideoAVI,
	".av1":      VideoAV1,
	".mpg":      VideoMPG,
	".mpeg":     VideoMPG,
	".mjpg":     VideoMJPG,
	".mjpeg":    VideoMJPG,
	".mp2":      VideoMP2,
	".mpv":      VideoMP2,
	".mp":       VideoMP4,
	ExtMP4:      VideoMP4, // .mp4
	".m4v":      VideoM4V,
	".mxf":      VideoMXF,
	".3gp":      Video3GP,
	".3g2":      Video3G2,
	".flv":      VideoFlash,
	".f4v":      VideoFlash,
	".mkv":      VideoMKV,
	".mts":      VideoAVCHD,
	".m2ts":     VideoBDAV,
	".ogv":      VideoOGV,
	".ogg":      VideoOGV,
	".ogx":      VideoOGV,
	".webm":     VideoWebM,
	".asf":      VideoASF,
	".wmv":      VideoWMV,
	".dv":       VideoDV,
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
	".xmp":      SidecarXMP,
	".aae":      SidecarAAE,
	".xml":      SidecarXML,
	ExtYAML:     SidecarYAML, // .yml
	".yaml":     SidecarYAML,
	".json":     SidecarJSON,
	".txt":      SidecarText,
	".nfo":      SidecarInfo,
	".md":       SidecarMarkdown,
	".markdown": SidecarMarkdown,
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
