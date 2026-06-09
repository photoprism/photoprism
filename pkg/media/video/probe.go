package video

import (
	"errors"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"github.com/abema/go-mp4"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/media"
)

// HeadScanLimit caps how far sample-entry chunk scans read into the file.
// The sample entry sits in the stsd box inside moov; faststart videos place
// moov at the head, and Motion Photos put it just past the JPEG (a few MB
// at most), so 16 MiB comfortably covers both layouts.
const HeadScanLimit = 16 * 1024 * 1024

// HeadScanChunks lists every video sample-entry code searched for in a single
// pass over the file head when go-mp4 does not report the codec directly.
var HeadScanChunks = append(append(Chunks{}, HevcChunks...), MagicYuvChunks...)

// ProbeFile returns information for the given filename.
func ProbeFile(fileName string) (info Info, err error) {
	if fileName == "" {
		return info, errors.New("filename missing")
	}

	var stat os.FileInfo
	var file *os.File

	// Ensure the file exists and is not a directory.
	if stat, err = os.Stat(fileName); err != nil {
		return info, err
	} else if stat.IsDir() {
		return info, errors.New("invalid filename")
	}

	// Open the file for reading.
	if file, err = os.Open(fileName); err != nil { //nolint:gosec // fileName validated by caller
		return info, err
	}

	defer func() {
		err = errors.Join(err, file.Close())
	}()

	// Get video information.
	info, err = Probe(file)

	// Return if failed.
	if err != nil {
		return info, err
	}

	// Add file name.
	info.FileName = fileName
	info.FileSize = stat.Size()

	// Set file type based on filename?
	if info.FileType == fs.TypeUnknown {
		info.FileType = fs.FileType(fileName)
	}

	// Set media type based on filename?
	if info.MediaType == media.Unknown {
		info.MediaType = media.Formats[info.FileType]
	}

	return info, err
}

// Probe returns information on the provided video file.
// see https://pkg.go.dev/github.com/abema/go-mp4#ProbeInfo
func Probe(file io.ReadSeeker) (info Info, err error) {
	// Set defaults.
	info = NewInfo()

	if file == nil {
		return info, errors.New("file is nil")
	}

	// Probe file from the beginning.
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return info, err
	}

	// Find file type start offset.
	if offset, findErr := CompatibleBrands.FileTypeOffset(file); findErr != nil {
		return info, findErr
	} else if offset < 0 {
		return info, nil
	} else {
		info.VideoOffset = int64(offset)
	}

	// Ignore any data before the video offset.
	videoFile := NewReadSeeker(file, info.VideoOffset)

	var video *mp4.ProbeInfo

	// Detect Mp4 video metadata.
	if video, err = mp4.Probe(videoFile); err != nil {
		return info, err
	}

	// Check compatibility.
	if CompatibleBrands.ContainsAny(video.CompatibleBrands) {
		info.Compatible = true
		info.VideoType = Mp4
		info.VideoMimeType = header.ContentTypeMp4
		info.FPS = 30.0 // TODO: Detect actual frames per second!

		if info.VideoOffset > 0 {
			info.MediaType = media.Live
			info.ThumbOffset = 0
		} else {
			info.MediaType = media.Video
			info.FileType = fs.VideoMp4
		}
	}

	// Check major brand.
	if video.MajorBrand == ChunkQT.Get() {
		info.VideoType = Mov
		info.VideoMimeType = header.ContentTypeMov
		if info.MediaType == media.Video {
			info.FileType = fs.VideoMov
		}
	}

	// Additional properties.
	info.FastStart = video.FastStart
	info.Tracks = len(video.Tracks)

	// Check tracks for codec, resolution and encryption.
	for _, track := range video.Tracks {
		if track.Encrypted && !info.Encrypted {
			info.Encrypted = track.Encrypted
		}

		if track.Codec == mp4.CodecAVC1 {
			info.VideoCodec = CodecAvc1
		}

		if avc := track.AVC; avc != nil {
			if info.VideoMimeType == "" {
				info.VideoMimeType = header.ContentTypeMp4
			}
			if w := int(avc.Width); w > info.VideoWidth {
				info.VideoWidth = w
			}
			if h := int(avc.Height); h > info.VideoHeight {
				info.VideoHeight = h
			}
		}
	}

	// If no AVC video was found, search the file head for other recognizable sample-entry
	// codes (HEVC, see https://stackoverflow.com/questions/63468587, and MagicYUV) in a single
	// pass. Each hit is validated as a real visual sample entry so a four-byte code colliding
	// with raw payload bytes (e.g. DV streams in a QuickTime container) is not misread as a
	// codec. The Codecs lookup normalizes each hit, e.g. HVC1/HVC2/HVC3/DVH1 → CodecHvc1 and
	// M8RG/M8Y2/… → CodecMagicYUV; the lookup is lowercased because MagicYUV codes are uppercase.
	if info.VideoCodec == "" {
		if pos, hit, fileErr := HeadScanChunks.SampleEntryOffset(file, HeadScanLimit); pos > 0 && fileErr == nil {
			info.VideoCodec = Codecs[strings.ToLower(hit.String())]
		}
	}

	// Calculate video duration in seconds.
	if video.Duration > 0 {
		s := float64(video.Duration) / float64(video.Timescale)

		info.Duration = time.Duration(s * float64(time.Second))

		if info.FPS > 0 {
			info.Frames = int(math.Round(info.FPS * s))
		}
	}

	return info, err
}
