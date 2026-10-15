package photoprism

import (
	"crypto/sha1" //nolint:gosec // indexed media hashes use SHA1
	"encoding/hex"
	"errors"
	"io"
	"os"
	"syscall"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/fs/disk"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/media/video"
)

// exportSidecar accepts only a verified AVC at the selected source's full-name destination.
func (w *Convert) exportSidecar(source *MediaFile) (string, ExportOutput, error) {
	name, err := w.avcName(source)
	if err != nil {
		return "", ExportOutput{}, err
	}
	output, err := exportVerifiedOutput(name, "avc")
	if err != nil {
		return "", ExportOutput{}, ErrExportNotReady
	}
	return name, output, nil
}

// exportOpenFile opens only a regular file without blocking on named pipes.
func exportOpenFile(name string, noFollow bool) (*os.File, error) {
	flags := os.O_RDONLY | syscall.O_NONBLOCK
	if noFollow {
		flags |= fs.OpenNoFollow
	}
	file, err := os.OpenFile(name, flags, 0) //nolint:gosec // names derive from admitted indexed sources or configured sidecar roots
	if err != nil {
		return nil, err
	}
	stat, err := file.Stat()
	if err != nil || !stat.Mode().IsRegular() {
		_ = file.Close()
		return nil, ErrExportUnavailable
	}
	return file, nil
}

// exportVerifiedOutput validates a regular container and its actual video stream.
func exportVerifiedOutput(name, kind string) (ExportOutput, error) {
	file, err := exportOpenFile(name, true)
	if err != nil {
		return ExportOutput{}, ErrExportUnavailable
	}
	defer file.Close()
	return exportVerifiedFile(file, kind)
}

// exportVerifiedFile probes the same open file that a download will serve.
func exportVerifiedFile(file *os.File, kind string) (ExportOutput, error) {
	stat, err := file.Stat()
	if err != nil || !stat.Mode().IsRegular() || stat.Size() == 0 {
		return ExportOutput{}, ErrExportUnavailable
	}
	info, err := video.Probe(io.NewSectionReader(file, 0, stat.Size()))
	exportVideoDimensions(&info)
	if err != nil || !info.Compatible || info.Encrypted || info.VideoWidth <= 0 || info.VideoHeight <= 0 || info.VideoCodec == video.CodecUnknown || kind == "avc" && (!exportAvc(string(info.VideoCodec)) || info.VideoFileExt() != ".mp4") {
		return ExportOutput{}, ErrExportUnavailable
	}
	current, err := file.Stat()
	if err != nil || current.Size() != stat.Size() || !current.ModTime().Equal(stat.ModTime()) {
		return ExportOutput{}, ErrExportSourceChanged
	}
	return ExportOutput{Ready: true, Preparable: true, Format: info.VideoFileExt()[1:], Codec: string(info.VideoCodec), Width: info.VideoWidth, Height: info.VideoHeight, Bytes: stat.Size()}, nil
}

// validate checks the selected open source against its indexed hash and captured version.
func (source *exportSource) validate(file *os.File) error {
	stat, err := file.Stat()
	if err != nil || !os.SameFile(stat, source.stat) || stat.Size() != source.stat.Size() || !stat.ModTime().Equal(source.stat.ModTime()) {
		return ErrExportSourceChanged
	}
	verify := func() error {
		hash := sha1.New() //nolint:gosec // indexed media hashes use SHA1
		if _, err = io.Copy(hash, io.NewSectionReader(file, 0, stat.Size())); err != nil || hex.EncodeToString(hash.Sum(nil)) != source.file.FileHash {
			return ErrExportSourceChanged
		}
		current, statErr := file.Stat()
		if statErr != nil || exportFileVersion(current) != exportFileVersion(stat) {
			return ErrExportSourceChanged
		}
		return source.version.check(source.media.FileName())
	}
	if key := exportFileVersion(stat); key != "" {
		err = exportValidation.verify(source.version.name+"\x00"+source.file.FileHash+"\x00"+key, verify)
	} else {
		select {
		case exportValidationSlots <- struct{}{}:
			defer func() { <-exportValidationSlots }()
		default:
			return ErrExportBusy
		}
		err = verify()
	}
	if err != nil {
		return err
	}
	return source.version.check(source.media.FileName())
}

// ExportFile opens a verified sidecar or embedded video segment without preparing files.
func (w *Convert) ExportFile(f *entity.File, kind string) (*os.File, int64, ExportOutput, error) {
	source, err := w.exportSource(f)
	if err != nil {
		return nil, 0, ExportOutput{}, err
	}
	if source.captureRole == "secondary" {
		return nil, 0, ExportOutput{}, ErrExportUnavailable
	}
	sourceFile, err := exportOpenFile(source.media.FileName(), false)
	if err != nil {
		return nil, 0, ExportOutput{}, ErrExportSourceChanged
	}
	if err = source.validate(sourceFile); err != nil {
		_ = sourceFile.Close()
		return nil, 0, ExportOutput{}, err
	}
	if kind == "video-part" {
		info, probeErr := video.Probe(io.NewSectionReader(sourceFile, 0, source.stat.Size()))
		info.FileSize = source.stat.Size()
		exportVideoDimensions(&info)
		if probeErr != nil || f.MediaType != entity.MediaLive || !info.Compatible || info.Encrypted || info.VideoOffset <= 0 || info.VideoSize() <= 0 || info.VideoWidth <= 0 || info.VideoHeight <= 0 {
			_ = sourceFile.Close()
			return nil, 0, ExportOutput{}, ErrExportUnavailable
		}
		if err = source.version.check(source.media.FileName()); err != nil {
			_ = sourceFile.Close()
			return nil, 0, ExportOutput{}, err
		}
		output := ExportOutput{Ready: true, Preparable: true, Format: info.VideoFileExt()[1:], Codec: string(info.VideoCodec), Width: info.VideoWidth, Height: info.VideoHeight, Bytes: info.VideoSize()}
		return sourceFile, info.VideoOffset, output, nil
	}
	_ = sourceFile.Close()
	if kind != "avc" {
		return nil, 0, ExportOutput{}, ErrExportUnavailable
	}
	name, err := w.avcName(source.media)
	if err != nil {
		return nil, 0, ExportOutput{}, err
	}
	file, err := exportOpenFile(name, true)
	if err != nil {
		return nil, 0, ExportOutput{}, ErrExportNotReady
	}
	output, err := exportVerifiedFile(file, "avc")
	if err == nil {
		err = source.version.check(source.media.FileName())
	}
	if err != nil {
		_ = file.Close()
		return nil, 0, ExportOutput{}, ErrExportNotReady
	}
	return file, 0, output, nil
}

// PrepareExport reuses standard AVC sidecars and returns embedded video without storing a copy.
func (w *Convert) PrepareExport(f *entity.File, kind string) (ExportOutput, error) {
	probe, err := w.ProbeExport(f)
	if err != nil {
		return ExportOutput{}, err
	}
	if kind == "video-part" {
		if probe.VideoPart == nil {
			return ExportOutput{}, ErrExportUnavailable
		}
		file, _, output, fileErr := w.ExportFile(f, kind)
		if file != nil {
			_ = file.Close()
		}
		return output, fileErr
	}
	if kind != "avc" || probe.Avc == nil {
		return ExportOutput{}, ErrExportUnavailable
	}
	source, err := w.exportSource(f)
	if err != nil {
		return ExportOutput{}, err
	}
	file, err := exportOpenFile(source.media.FileName(), false)
	if err != nil {
		return ExportOutput{}, ErrExportSourceChanged
	}
	err = source.validate(file)
	_ = file.Close()
	if err != nil {
		return ExportOutput{}, err
	}
	if source.companion != nil {
		companionFile, openErr := exportOpenFile(source.companion.media.FileName(), false)
		if openErr != nil {
			return ExportOutput{}, ErrExportSourceChanged
		}
		err = source.companion.validate(companionFile)
		_ = companionFile.Close()
		if err != nil {
			return ExportOutput{}, err
		}
	}
	if _, output, readyErr := w.exportSidecar(source.media); readyErr == nil {
		return output, nil
	}
	if !probe.Avc.Preparable {
		return ExportOutput{}, ErrExportUnavailable
	}
	if w.conf.InsufficientStorage() {
		return ExportOutput{}, status.ErrInsufficientStorage
	}
	_, err = w.coordinatedAvc(source.media, w.conf.FFmpegEncoder(), false, false, true, source.version)
	if err != nil {
		if disk.IsNoSpace(err) || errors.Is(err, status.ErrInsufficientStorage) {
			return ExportOutput{}, status.ErrInsufficientStorage
		}
		return ExportOutput{}, err
	}
	_, output, err := w.exportSidecar(source.media)
	return output, err
}
