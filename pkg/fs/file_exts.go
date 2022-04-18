package fs

import (
	"path/filepath"
	"strings"
)

// FileExtensions maps file extensions to standard formats
type FileExtensions map[string]Type

// Extensions contains the filename extensions of file formats known to PhotoPrism.
var Extensions = FileExtensions{
	".jpg":      ImageJPEG,
	".jpeg":     ImageJPEG,
	".jpe":      ImageJPEG,
	".jif":      ImageJPEG,
	".jfif":     ImageJPEG,
	".jfi":      ImageJPEG,
	".thm":      ImageJPEG,
	".heif":     ImageHEIF,
	".heic":     ImageHEIF,
	".heifs":    ImageHEIF,
	".heics":    ImageHEIF,
	".avci":     ImageHEIF,
	".avcs":     ImageHEIF,
	".avif":     ImageHEIF,
	".avifs":    ImageHEIF,
	".webp":     ImageWebP,
	".tif":      ImageTIFF,
	".tiff":     ImageTIFF,
	".png":      ImagePNG,
	".pn":       ImagePNG,
	".mpo":      ImageMPO,
	".gif":      ImageGIF,
	".bmp":      ImageBMP,
	".3fr":      RawImage,
	".ari":      RawImage,
	".arw":      RawImage,
	".bay":      RawImage,
	".cap":      RawImage,
	".crw":      RawImage,
	".cr2":      RawImage,
	".cr3":      RawImage,
	".data":     RawImage,
	".dcs":      RawImage,
	".dcr":      RawImage,
	".dng":      RawImage,
	".drf":      RawImage,
	".eip":      RawImage,
	".erf":      RawImage,
	".fff":      RawImage,
	".gpr":      RawImage,
	".iiq":      RawImage,
	".k25":      RawImage,
	".kdc":      RawImage,
	".mdc":      RawImage,
	".mef":      RawImage,
	".mos":      RawImage,
	".mrw":      RawImage,
	".nef":      RawImage,
	".nrw":      RawImage,
	".obm":      RawImage,
	".orf":      RawImage,
	".pef":      RawImage,
	".ptx":      RawImage,
	".pxn":      RawImage,
	".r3d":      RawImage,
	".raf":      RawImage,
	".raw":      RawImage,
	".rwl":      RawImage,
	".rwz":      RawImage,
	".rw2":      RawImage,
	".srf":      RawImage,
	".srw":      RawImage,
	".sr2":      RawImage,
	".x3f":      RawImage,
	".hevc":     VideoHEVC,
	".mov":      VideoMOV,
	".qt":       VideoMOV,
	".avi":      VideoAVI,
	".av1":      VideoAV1,
	".avc":      VideoAVC,
	".vvc":      VideoVVC,
	".mpg":      VideoMPG,
	".mpeg":     VideoMPG,
	".mjpg":     VideoMJPG,
	".mjpeg":    VideoMJPG,
	".mp2":      VideoMP2,
	".mpv":      VideoMP2,
	".mp":       VideoMP4,
	".mp4":      VideoMP4,
	".m4v":      VideoMP4,
	".3gp":      Video3GP,
	".3g2":      Video3G2,
	".flv":      VideoFlash,
	".f4v":      VideoFlash,
	".mkv":      VideoMKV,
	".mts":      VideoAVCHD,
	".ogv":      VideoOGV,
	".ogg":      VideoOGV,
	".ogx":      VideoOGV,
	".webm":     VideoWebM,
	".asf":      VideoASF,
	".wmv":      VideoWMV,
	".xmp":      XmpFile,
	".aae":      AaeFile,
	".xml":      XmlFile,
	".yml":      YamlFile,
	".yaml":     YamlFile,
	".json":     JsonFile,
	".txt":      TextFile,
	".md":       MarkdownFile,
	".markdown": MarkdownFile,
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

// TypesExt returns known extensions by file type.
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
