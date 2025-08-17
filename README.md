**Thumbra** is a Go application for retrieving and generating thumbnails for files/media stored in S3 buckets.

It is currently in development, aimed to replace the MediaWiki thumb generator (`thumb.php`) with a more robust implementation - and nicer human-readable URLs. It also allows for generating thumbnails on the fly with different options.

## Why?
MediaWiki's default URL structure - taking into account Telepedia uses S3 - is such that the URLs end up to be `/{wiki}/{hash1}/{hash2}/{fileName}`, this is fine. 

When we get to previous revisions, or thumbs, however, it becomes a bit more "ugly" and we get `{wiki}/thumb/{hash1}/{hash2}/{fileName}/{size}-{fileName}?{revision}`. 

...more to come