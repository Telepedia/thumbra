**Thumbra** is a Go application for retrieving and generating thumbnails for files/media stored in S3 buckets. It is aptly named after Thumbor.

It is currently in development, aimed to replace the MediaWiki thumb generator (`thumb.php`) with a more robust implementation - and nicer human-readable URLs. It also allows for generating thumbnails on the fly with different options.

## Why?
MediaWiki's default URL structure - taking into account Telepedia uses S3 - is such that the URLs end up to be `/{wiki}/{hash1}/{hash2}/{fileName}`, this is fine. 

When we get to previous revisions, or thumbs, however, it becomes a bit more "ugly" and we get `{wiki}/thumb/{hash1}/{hash2}/{fileName}/{size}-{fileName}?{revision}`. 

We think that this can be cleaned up with a much nicer syntax, which is more human readable and easier to understand. The API is explained below along with all of the current routes. 

### API

#### Original File (and older revisions)

Getting the original file is simple, the route for this is such:

`/{wiki}/{hash1}/{hash2}/{filename}/revision/{revision}`

The revision accepts two parameters either `latest` or a timestamp in `YYYYMMDDHHSS` format. If latest, the current version of the file will be returned. If you provide a timestamp, the file as existing at that time will be provided - assuming that the file exists at that location. 

This assumes that your MediaWiki instance stores archives in the default place, where archives are placed in the `/archive/{hash1}/{hash2}/` folder, and their filename is `YYYYMMDDHHSS!{filename}`, note that MediaWiki renames these files with the timestamp, an exclamation mark (!) and then the original filename. 

#### Thumbnails

...to do
