package fs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

type imageFormat int

const (
	imageFormatUnknown imageFormat = iota
	imageFormatJPEG
	imageFormatPNG
	imageFormatGIF
	imageFormatBMP
	imageFormatTIFF
	imageFormatWEBP
)

var errUnsupportedImageFormat = fmt.Errorf("unsupported image format")

// openImageSection opens a file and returns a bounded reader spanning its current size.
func openImageSection(fileName string) (file *os.File, reader *io.SectionReader, err error) {
	file, err = os.Open(fileName) //nolint:gosec // fileName is supplied by the caller and may point to user media
	if err != nil {
		return nil, nil, err
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, nil, err
	}

	return file, io.NewSectionReader(file, 0, info.Size()), nil
}

// DecodeImageFile opens and decodes a natively supported image file using direct format decoders.
func DecodeImageFile(fileName string) (_ image.Image, _ string, err error) {
	file, reader, err := openImageSection(fileName)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	return decodeImage(reader)
}

// DecodeImageConfigFile opens a natively supported image file and decodes only its config.
func DecodeImageConfigFile(fileName string) (_ image.Config, _ string, err error) {
	file, reader, err := openImageSection(fileName)
	if err != nil {
		return image.Config{}, "", err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	return decodeImageConfig(reader)
}

// DecodeImageData decodes a natively supported image from a bounded in-memory buffer.
func DecodeImageData(data []byte) (image.Image, string, error) {
	reader := bytes.NewReader(data)
	return decodeImage(io.NewSectionReader(reader, 0, int64(len(data))))
}

// DecodeImageConfigData decodes the config for a natively supported in-memory image buffer.
func DecodeImageConfigData(data []byte) (image.Config, string, error) {
	reader := bytes.NewReader(data)
	return decodeImageConfig(io.NewSectionReader(reader, 0, int64(len(data))))
}

// decodeImage decodes a bounded image reader by dispatching to the matching direct decoder.
func decodeImage(reader *io.SectionReader) (image.Image, string, error) {
	format, err := detectImageFormat(reader)
	if err != nil {
		return nil, "", err
	}

	switch format {
	case imageFormatJPEG:
		img, err := jpeg.Decode(reader)
		return img, "jpeg", err
	case imageFormatPNG:
		img, err := png.Decode(reader)
		return img, "png", err
	case imageFormatGIF:
		img, err := gif.Decode(reader)
		return img, "gif", err
	case imageFormatBMP:
		img, err := bmp.Decode(reader)
		return img, "bmp", err
	case imageFormatTIFF:
		img, err := tiff.Decode(reader)
		return img, "tiff", err
	case imageFormatWEBP:
		img, err := webp.Decode(reader)
		return img, "webp", err
	default:
		return nil, "", errUnsupportedImageFormat
	}
}

// decodeImageConfig decodes only the image config from a bounded reader.
func decodeImageConfig(reader *io.SectionReader) (image.Config, string, error) {
	format, err := detectImageFormat(reader)
	if err != nil {
		return image.Config{}, "", err
	}

	switch format {
	case imageFormatJPEG:
		cfg, err := jpeg.DecodeConfig(reader)
		return cfg, "jpeg", err
	case imageFormatPNG:
		cfg, err := png.DecodeConfig(reader)
		return cfg, "png", err
	case imageFormatGIF:
		cfg, err := gif.DecodeConfig(reader)
		return cfg, "gif", err
	case imageFormatBMP:
		cfg, err := bmp.DecodeConfig(reader)
		return cfg, "bmp", err
	case imageFormatTIFF:
		cfg, err := tiff.DecodeConfig(reader)
		return cfg, "tiff", err
	case imageFormatWEBP:
		cfg, err := webp.DecodeConfig(reader)
		return cfg, "webp", err
	default:
		return image.Config{}, "", errUnsupportedImageFormat
	}
}

// detectImageFormat inspects the header bytes and validates TIFF offsets before decode.
func detectImageFormat(reader *io.SectionReader) (imageFormat, error) {
	headerSize := reader.Size()
	if headerSize <= 0 {
		return imageFormatUnknown, io.ErrUnexpectedEOF
	}

	header := make([]byte, min(16, int(headerSize)))
	if _, err := reader.ReadAt(header, 0); err != nil && err != io.EOF {
		return imageFormatUnknown, err
	}

	switch {
	case isJPEGHeader(header):
		return imageFormatJPEG, nil
	case isPNGHeader(header):
		return imageFormatPNG, nil
	case isGIFHeader(header):
		return imageFormatGIF, nil
	case isBMPHeader(header):
		return imageFormatBMP, nil
	case isWEBPHeader(header):
		return imageFormatWEBP, nil
	case isTIFFHeader(header):
		if err := validateTIFFOffset(header, reader.Size()); err != nil {
			return imageFormatUnknown, err
		}
		return imageFormatTIFF, nil
	default:
		return imageFormatUnknown, errUnsupportedImageFormat
	}
}

// isJPEGHeader reports whether the header matches the JPEG SOI marker.
func isJPEGHeader(header []byte) bool {
	return len(header) >= 2 && header[0] == 0xff && header[1] == 0xd8
}

// isPNGHeader reports whether the header matches the PNG signature.
func isPNGHeader(header []byte) bool {
	return len(header) >= 8 && bytes.Equal(header[:8], []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a})
}

// isGIFHeader reports whether the header matches a GIF signature.
func isGIFHeader(header []byte) bool {
	return len(header) >= 6 && (bytes.Equal(header[:6], []byte("GIF87a")) || bytes.Equal(header[:6], []byte("GIF89a")))
}

// isBMPHeader reports whether the header matches a BMP signature.
func isBMPHeader(header []byte) bool {
	return len(header) >= 2 && header[0] == 'B' && header[1] == 'M'
}

// isWEBPHeader reports whether the header matches a RIFF WEBP container.
func isWEBPHeader(header []byte) bool {
	return len(header) >= 12 && bytes.Equal(header[:4], []byte("RIFF")) && bytes.Equal(header[8:12], []byte("WEBP"))
}

// isTIFFHeader reports whether the header matches a classic TIFF signature.
func isTIFFHeader(header []byte) bool {
	return len(header) >= 8 && (bytes.Equal(header[:4], []byte{0x49, 0x49, 0x2a, 0x00}) || bytes.Equal(header[:4], []byte{0x4d, 0x4d, 0x00, 0x2a}))
}

// validateTIFFOffset checks the initial TIFF IFD offset against the bounded reader size.
func validateTIFFOffset(header []byte, size int64) error {
	if len(header) < 8 {
		return io.ErrUnexpectedEOF
	}

	var order binary.ByteOrder
	switch {
	case bytes.Equal(header[:4], []byte{0x49, 0x49, 0x2a, 0x00}):
		order = binary.LittleEndian
	case bytes.Equal(header[:4], []byte{0x4d, 0x4d, 0x00, 0x2a}):
		order = binary.BigEndian
	default:
		return fmt.Errorf("invalid TIFF header")
	}

	ifdOffset := int64(order.Uint32(header[4:8]))
	if ifdOffset < 0 || ifdOffset >= size {
		return fmt.Errorf("invalid TIFF: IFD offset %d exceeds file size %d", ifdOffset, size)
	}

	return nil
}
