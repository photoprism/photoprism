package raw

// previewUnsafeExt lists RAW extensions whose embedded JPEG preview is unusable (e.g. fails
// thumbnailing), so preview extraction is skipped in favor of a full RAW render.
var previewUnsafeExt = map[string]bool{
	".mos": true, // Leaf .mos files ship a bogus-Huffman embedded preview.
}

// PreviewExtAllowed reports whether the embedded preview may be extracted for the extension,
// which must be lowercase with a leading dot (e.g. ".cr3").
func PreviewExtAllowed(ext string) bool {
	return !previewUnsafeExt[ext]
}
