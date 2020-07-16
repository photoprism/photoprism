# Metadata Test Files

You may use [ImageMagick](http://www.imagemagick.org/Usage/resize/#resample) 
to create smaller versions of images to be stored in this repository:

```
convert original.JPG -resize 500 small.JPG
```

JSON sidecar files for testing can be created with [exiftool](https://exiftool.org/):

```
exiftool -j example.jpg > example.json
```
