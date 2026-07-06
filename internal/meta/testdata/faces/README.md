# XMP Face-Region Fixtures

Test fixtures for reading face markers and names from XMP metadata
(`meta.Data.Faces`). Both metadata schemas and both delivery paths are covered.

| File               | Schema                      | Path                   | Region(s)           |
|--------------------|-----------------------------|------------------------|---------------------|
| `mwg-single.json`  | MWG-RS (`RegionList`)       | Embedded ExifTool JSON | Alice (scalar)      |
| `mwg-multi.json`   | MWG-RS (`RegionList`)       | Embedded ExifTool JSON | Alice, Bob (arrays) |
| `mp.json`          | Microsoft `MP:RegionInfo`   | Embedded ExifTool JSON | Cara                |
| `adobe-mwg.xmp`    | MWG-RS (attribute form)     | `.xmp` sidecar         | Diana               |
| `digikam-mwg.xmp`  | MWG-RS (child-element form) | `.xmp` sidecar         | Carl                |
| `microsoft-mp.xmp` | Microsoft `MP:RegionInfo`   | `.xmp` sidecar         | Cara                |

The `.json` files are real `exiftool -n -m -api LargeFileSupport -j` dumps
(PhotoPrism's exact arguments), so they preserve the scalar-vs-parallel-array
shape ExifTool emits for one vs. many regions. The `.xmp` sidecars cover both
RDF/XML shapes writers emit for the `mwg-rs:Area` sub-fields: the child-element
form (`<stArea:x>0.6</stArea:x>`, digiKam) and the attribute form
(`stArea:x="0.4"`, Adobe/Lightroom); both are handled.

Regenerate a fixture by writing regions into a JPEG and re-exporting, e.g.:

    exiftool -RegionName="Alice" -RegionType="Face" \
      -RegionAreaX=0.5 -RegionAreaY=0.4 -RegionAreaW=0.1 -RegionAreaH=0.15 \
      -RegionAreaUnit=normalized image.jpg
    exiftool -n -m -api LargeFileSupport -j image.jpg > mwg-single.json
