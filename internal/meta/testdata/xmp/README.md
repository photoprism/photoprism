# XMP Fixture Corpus

Test fixtures for the `internal/meta` XMP sidecar reader rewrite (issue #2260, proposal `specs/proposals/xmp-improvement.md`).

Each fixture is paired with an `.exiftool.txt` reference (output of `exiftool -X <file>`) that captures the canonical interpretation; regression tests compare reader output against the same XMP read through ExifTool.

## Layout

| Directory    | Source                                        | Purpose                                                                                                                                 |
|--------------|-----------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------|
| `adobe/`     | Adobe Bridge / Lightroom / Camera Raw exports | Reference implementation. Covers the full EXIF camera/lens/exposure surface, GPS in 2-component Adobe form, and the `xmpMM:*` triple.   |
| `darktable/` | Darktable lighttable + map mode               | Open-source writer with `<rdf:Bag>` for `dc:subject`, vendor `darktable:*` history, and the 2-component GPS form.                       |
| `digikam/`   | digiKam Edit Metadata dialog                  | Open-source writer with `xmpRights:UsageTerms` (License), `xmpMM:DocumentID/InstanceID/OriginalDocumentID` triple, and section-2 noise. |
| `synthetic/` | Hand-written, validated with `exiftool -X`    | Targeted coverage for tags absent in the tool fixtures and edge cases that real writers don't reliably produce.                         |

## Regenerating Exiftool References

After editing a fixture:

    for f in internal/meta/testdata/xmp/{adobe,darktable,digikam,synthetic}/*.xmp; do
      exiftool -X "$f" > "${f%.xmp}.exiftool.txt"
    done

## Synthetic Fixture Index

| File                        | Purpose                                                                                                                                   |
|-----------------------------|-------------------------------------------------------------------------------------------------------------------------------------------|
| `software-only.xmp`         | `xmp:CreatorTool` (Software). Single-tag fixture for the simplest accessor.                                                               |
| `gps-time-combined.xmp`     | `exif:GPSTimeStamp` as a single combined ISO 8601 / RFC 3339 datetime — the spec-canonical encoding for `TakenGps`.                       |
| `gps-time-split.xmp`        | Legacy split form: `exif:GPSDateStamp` + `exif:GPSTimeStamp`. Some older writers emit this; secondary `TakenGps` fallback.                |
| `time-offsets-subsec.xmp`   | `exif:OffsetTimeOriginal/OffsetTime/OffsetTimeDigitized` cascade plus `exif:SubSecTimeOriginal` joined into `TakenAt`.                    |
| `notes-usercomment.xmp`     | `exif:UserComment` as `lang-alt` (`<rdf:Alt>` with `x-default` + `en` + `de`). Confirms reader prefers `x-default`.                       |
| `xmpdm-creationdate.xmp`    | `xmpDM:CreationDate` — secondary fallback for `CreatedAt` (rare; emitted by Adobe Premiere/After Effects).                                |
| `subject-seq.xmp`           | `dc:subject` as `<rdf:Seq>` instead of `<rdf:Bag>`. Confirms reader handles both list types.                                              |
| `aux-only.xmp`              | Adobe `aux:` namespace — `OwnerName`, `Lens`, `LensID`, `LensSerialNumber`, `SerialNumber`, `Firmware`. Legacy Lightroom/ACR.             |
| `exifex-camera-lens.xmp`    | EXIF 2.3 for XMP `exifEX:` namespace — `LensMake`, `LensModel`, `SerialNumber` (= EXIF BodySerialNumber), `PhotographicSensitivity`, etc. |
| `gpano-360.xmp`             | Google Photo Sphere — `GPano:ProjectionType = equirectangular` plus required dimension fields.                                            |
| `multi-rdf-description.xmp` | Bug 2 demonstration: four sibling `<rdf:Description>` blocks each declaring a different namespace. Tests that XPath walks all blocks.     |
| `alt-edge-cases.xmp`        | `<rdf:Alt>` shapes that real writers don't always produce: no `x-default` (only `de`/`en`), missing `xml:lang`, duplicate `x-default`.    |

## Source Notes

- `digikam/aurora.jpg.xmp` deliberately includes section-2 (out-of-scope) tags such as `xmp:Label`, `digiKam:TagsList`, `lr:hierarchicalSubject`, `MicrosoftPhoto:LastKeywordXMP`, and `xmpMM:History`. These are negative-test bait — the reader must **not** persist them.
- `darktable/aurora.jpg.xmp` includes the `<darktable:history>` develop stack (also section-2 ignored) and uses the 2-component Adobe GPS form (`87,21.291962N`) that today's `GpsToDecimal` silently drops.
- Adobe Bridge fixtures emit `<exif:Flash>` as a struct with `<exif:Fired>` sub-field — the only fixtures in the corpus that exercise the `Flash` composition rule.
