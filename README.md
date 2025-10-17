**Thumbra** is a Go application for retrieving and generating thumbnails for files/media stored in S3 buckets. It is aptly named after Thumbor.

It is currently in development, aimed to replace the MediaWiki thumb generator (`thumb.php`) with a more robust implementation - and nicer human-readable URLs. It also allows for generating thumbnails on the fly with different options.

**Note:** Thumbra does not handle deleted files. These should never be publicly visible - MediaWiki handles moving these to-and-from the `deleted` zone when they are deleted or undeleted. They should not be publicly visible.

## Why?
MediaWiki's default URL structure - taking into account Telepedia uses S3 - is such that the URLs end up to be `/{wiki}/{hash1}/{hash2}/{fileName}`, this is fine. 

When we get to previous revisions, or thumbs, however, it becomes a bit more "ugly" and we get `{wiki}/thumb/{hash1}/{hash2}/{fileName}/{size}-{fileName}?{revision}`. 

We think that this can be cleaned up with a much nicer syntax, which is more human readable and easier to understand. The API is explained below along with all of the current routes. 

### API

#### Full size file

Getting the original file is simple, the route for this is such:

`/{wiki}/{hash1}/{hash2}/{filename}/revision/{revision}`

The revision accepts two parameters either `latest` or a timestamp in `YYYYMMDDHHSS` format. If latest, the current version of the file will be returned. If you provide a timestamp, the file as existing at that time will be provided - assuming that the file exists at that location. 

This assumes that your MediaWiki instance stores archives in the default place, where archives are placed in the `/archive/{hash1}/{hash2}/` folder, and their filename is `YYYYMMDDHHSS!{filename}`, note that MediaWiki renames these files with the timestamp, an exclamation mark (!) and then the original filename. 

#### Thumbnails

### Scaled to width

To get a thumbnail of a file, the route is as such 

`/{wiki}/{hash1}/{hash2}/{filename}/revision/{revision}/scale-to-width/{width}`

The revision, as before, accepts either `latest` or the timestamp as `YYYYMMDDHHSS`. The API will first try and return the file at that width, or if it does not exist, will thumbnail the file to that width, return it, and store it in S3. 

Note: this route will refuse to upscale an image (for obvious reasons, also becasue that is what MediaWiki does natively). However, unlike MediaWiki, if the width > the original width, instead of returning an error, like MediaWiki does, the original image will be returned at the full size. This prevents broken display of images in wikis - in any case, the size of the image returned will be at least =< the size requested, so will not appear larger than requested.

#### Passthrough/Supported types

The API currently supports thumbnailing the following media types:
* PNG
* JPEG/JPG
* WEBP
* GIF
