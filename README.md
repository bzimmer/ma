# ma (media archiver)

![build](https://github.com/bzimmer/ma/actions/workflows/build.yaml/badge.svg)

Simple tools for managing media files with [SmugMug](https://smugmug.com/)

Uses [smugmug](https://github.com/bzimmer/smugmug) for accessing the SmugMug [API](https://api.smugmug.com)

# usage

```sh
NAME:
   ma - CLI for managing photos locally and at SmugMug

USAGE:
   ma [global options] command [command options] [arguments...]

COMMANDS:
   user      query the authenticated user
   find      search for album or folders by name
   ls, list  list albums and/or folders
   up        upload images to SmugMug
   cp        copy files to a pre-determined directory structure
   help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --smugmug-client-key value     smugmug client key (default: "xxx") [$SMUGMUG_CLIENT_KEY]
   --smugmug-client-secret value  smugmug client secret (default: "xxx") [$SMUGMUG_CLIENT_SECRET]
   --smugmug-access-token value   smugmug access token (default: "xxx") [$SMUGMUG_ACCESS_TOKEN]
   --smugmug-token-secret value   smugmug token secret (default: "xxx") [$SMUGMUG_TOKEN_SECRET]
   --debug                        enable debugging (default: false)
   --help, -h                     show help (default: false)
```

## find

`find` will look nodes for the specified query. If `--scope` is not specified, the currently
authenticated user's scope is used.

```sh
$ ma find --scope "/api/v2/user/cmac" Event
2021-08-02T18:55:02-07:00 INF find name=Events nodeID=kTR76 parentID=zx4Fx type=Folder
2021-08-02T18:55:02-07:00 INF find albumKey=mjXDhW imageCount=79 name="Harley Pan America event" nodeID=sBt5dp
2021-08-02T18:55:02-07:00 INF find albumKey=dDMKTz imageCount=6 name="SmugMug camera straps!" nodeID=vGtbt
```

```sh
$ ma find --json --scope "/api/v2/user/cmac" Event
{"name":"Events","nodeID":"kTR76","parentID":"zx4Fx","type":"Folder"}
{"albumKey":"mjXDhW","imageCount":79,"name":"Harley Pan America event","nodeID":"sBt5dp"}
{"albumKey":"dDMKTz","imageCount":6,"name":"SmugMug camera straps!","nodeID":"vGtbt"}
```

## ls (list)

`ls` returns all the nodes under the specified parent node.

```sh
$ ma ls -R --json kTR76
{"name":"Events","nodeID":"kTR76","parentID":"zx4Fx","type":"Folder"}
{"albumKey":"jbBNhR","imageCount":16,"name":"Black lives matter protest","nodeID":"q2qP7F","parentID":"kTR76","type":"Album"}
{"albumKey":"GNpNRf","imageCount":11,"name":"47th annual electric vehicle show silicon valley","nodeID":"6zFLgD","parentID":"kTR76","type":"Album"}
{"albumKey":"LPZThB","imageCount":21,"name":"Humans melt and pets soak up love at the Bay Area Pet Fair","nodeID":"MG5c5r","parentID":"kTR76","type":"Album"}
{"albumKey":"MGxZXq","imageCount":36,"name":"San Francisco Reptile Expo","nodeID":"kvh7p7","parentID":"kTR76","type":"Album"}
{"albumKey":"KJXpSD","imageCount":25,"name":"Muttville adoption ceremony","nodeID":"JbnGdJ","parentID":"kTR76","type":"Album"}
{"albumKey":"pttBvK","imageCount":18,"name":"SF Body Art Expo 2018","nodeID":"xH9vPF","parentID":"kTR76","type":"Album"}
{"albumKey":"gJ7d6r","imageCount":35,"name":"Maker Faire 2008","nodeID":"BjcPR","parentID":"kTR76","type":"Album"}
{"albumKey":"2XrGxm","imageCount":33,"name":"Santa Cruz fungus festival 2019","nodeID":"jM739r","parentID":"kTR76","type":"Album"}
{"albumKey":"B8BLsX","imageCount":30,"name":"Golden Gate dog show 2019","nodeID":"LtqS2s","parentID":"kTR76","type":"Album"}
{"albumKey":"wqmpff","imageCount":22,"name":"Spanish Fork High 50th reunion 2019","nodeID":"5FD6k8","parentID":"kTR76","type":"Album"}
{"albumKey":"qH5CG3","imageCount":35,"name":"Spanish Fork Fiesta Days parade 2019","nodeID":"J3wNMj","parentID":"kTR76","type":"Album"}
```

If `ls` queries an album, only the album details are returned (for now).

```sh
$ ma ls -R --json q2qP7F
task: [ma] go run cmd/ma/*.go ls -R --json q2qP7F
{"albumKey":"jbBNhR","imageCount":16,"name":"Black lives matter protest","nodeID":"q2qP7F","parentID":"kTR76","type":"Album"}
```

## up (upload)

`up` uploads files to the specified gallery. The `up` command uses multiple goroutines to concurrently upload images.

The `up` command queries the gallery for existing images and uses the filename and MD5 to compare against local files
to determine if an image should be uploaded. In this example the four files do not exist in the gallery.

```sh
$ ma up --album 7dXUSm $HOME/Pictures/_Export
2021-08-02T19:14:08-07:00 INF querying existing gallery images
2021-08-02T19:14:09-07:00 INF existing gallery images count=0
2021-08-02T19:14:09-07:00 INF skipping path=/Users/bzimmer/Pictures/_Export/.DS_Store reason=unsupported
2021-08-02T19:14:09-07:00 INF uploadable path=/Users/bzimmer/Pictures/c/2021-07-23/DSCF6020.jpg
2021-08-02T19:14:09-07:00 INF upload album=7dXUSm name=DSCF6020.jpg replaces= status=uploading
2021-08-02T19:14:09-07:00 INF uploadable path=/Users/bzimmer/Pictures/_Export/2021-07-24/DSCF6052.jpg
2021-08-02T19:14:09-07:00 INF upload album=7dXUSm name=DSCF6052.jpg replaces= status=uploading
2021-08-02T19:14:09-07:00 INF uploadable path=/Users/bzimmer/Pictures/_Export/2021-07-24/DSCF6113.jpg
2021-08-02T19:14:09-07:00 INF upload album=7dXUSm name=DSCF6113.jpg replaces= status=uploading
2021-08-02T19:14:09-07:00 INF uploadable path=/Users/bzimmer/Pictures/_Export/2021-07-24/DSCF6125.jpg
2021-08-02T19:14:09-07:00 INF upload album=7dXUSm name=DSCF6125.jpg replaces= status=uploading
2021-08-02T19:14:14-07:00 INF upload album=7dXUSm elapsed=4207.259194 name=DSCF6113.jpg status=success uri=/api/v2/image/c7bk872-0
2021-08-02T19:14:18-07:00 INF upload album=7dXUSm elapsed=8141.954387 name=DSCF6125.jpg status=success uri=/api/v2/image/8m66Wsb-0
2021-08-02T19:14:18-07:00 INF upload album=7dXUSm elapsed=8169.71444 name=DSCF6052.jpg status=success uri=/api/v2/image/93qT6kX-0
2021-08-02T19:14:18-07:00 INF upload album=7dXUSm elapsed=8176.553802 name=DSCF6020.jpg status=success uri=/api/v2/image/BMWQ6pL-0
2021-08-02T19:14:18-07:00 INF counters count=1 metric=ma.fsUploadable.skip.unsupported
2021-08-02T19:14:18-07:00 INF counters count=4 metric=ma.fsUploadable.open
2021-08-02T19:14:18-07:00 INF counters count=4 metric=ma.upload.attempt
2021-08-02T19:14:18-07:00 INF counters count=4 metric=ma.upload.success
2021-08-02T19:14:18-07:00 INF samples count=4 max=8.176553726196289 mean=7.173870325088501 metric=ma.upload.upload min=4.207259178161621 stddev=1.9777973515749452
```

A second upload attempt will not upload anything because the filename and MD5 match.

```sh
$ ma up --album 7dXUSm $HOME/Pictures/_Export
2021-08-02T19:09:06-07:00 INF querying existing gallery images
2021-08-02T19:09:09-07:00 INF existing gallery images count=95
2021-08-02T19:09:09-07:00 INF skipping path=/Users/bzimmer/Pictures/_Export/.DS_Store reason=unsupported
2021-08-02T19:09:09-07:00 INF skipping path=/Users/bzimmer/Pictures/_Export/2021-07-23/DSCF6020.jpg reason=md5
2021-08-02T19:09:09-07:00 INF skipping path=/Users/bzimmer/Pictures/_Export/2021-07-24/DSCF6052.jpg reason=md5
2021-08-02T19:09:09-07:00 INF skipping path=/Users/bzimmer/Pictures/_Export/2021-07-24/DSCF6113.jpg reason=md5
2021-08-02T19:09:09-07:00 INF skipping path=/Users/bzimmer/Pictures/_Export/2021-07-24/DSCF6125.jpg reason=md5
2021-08-02T19:09:09-07:00 INF complete
2021-08-02T19:09:09-07:00 INF counters count=4 metric=ma.fsUploadable.open
2021-08-02T19:09:09-07:00 INF counters count=4 metric=ma.fsUploadable.skip.md5
2021-08-02T19:09:09-07:00 INF counters count=1 metric=ma.fsUploadable.skip.unsupported
```

Updating existing images and uploading again will result in new versions at SmugMug (noted by the `uri` suffix).

```sh
$ ma up --album 7dXUSm $HOME/Pictures/_Export
2021-08-02T19:08:52-07:00 INF querying existing gallery images
2021-08-02T19:08:54-07:00 INF existing gallery images count=95
2021-08-02T19:08:54-07:00 INF skipping path=/Users/bzimmer/Pictures/_Export/.DS_Store reason=unsupported
2021-08-02T19:08:54-07:00 INF uploadable path=/Users/bzimmer/Pictures/_Export/2021-07-23/DSCF6020.jpg
2021-08-02T19:08:54-07:00 INF upload album=7dXUSm name=DSCF6020.jpg replaces=/api/v2/image/sMbr3QX-1 status=uploading
2021-08-02T19:08:54-07:00 INF uploadable path=/Users/bzimmer/Pictures/_Export/2021-07-24/DSCF6052.jpg
2021-08-02T19:08:54-07:00 INF upload album=7dXUSm name=DSCF6052.jpg replaces=/api/v2/image/7HP3GhG-1 status=uploading
2021-08-02T19:08:54-07:00 INF uploadable path=/Users/bzimmer/Pictures/_Export/2021-07-24/DSCF6113.jpg
2021-08-02T19:08:54-07:00 INF upload album=7dXUSm name=DSCF6113.jpg replaces=/api/v2/image/Pj4DHR6-1 status=uploading
2021-08-02T19:08:54-07:00 INF uploadable path=/Users/bzimmer/Pictures/_Export/2021-07-24/DSCF6125.jpg
2021-08-02T19:08:54-07:00 INF upload album=7dXUSm name=DSCF6125.jpg replaces=/api/v2/image/DFTxbKB-1 status=uploading
2021-08-02T19:08:58-07:00 INF upload album=7dXUSm elapsed=3970.703549 name=DSCF6052.jpg status=success uri=/api/v2/image/7HP3GhG-2
2021-08-02T19:09:00-07:00 INF upload album=7dXUSm elapsed=5182.417988 name=DSCF6020.jpg status=success uri=/api/v2/image/sMbr3QX-2
2021-08-02T19:09:01-07:00 INF upload album=7dXUSm elapsed=6976.05742 name=DSCF6125.jpg status=success uri=/api/v2/image/DFTxbKB-2
2021-08-02T19:09:03-07:00 INF upload album=7dXUSm elapsed=8587.274661 name=DSCF6113.jpg status=success uri=/api/v2/image/Pj4DHR6-2
2021-08-02T19:09:03-07:00 INF counters count=4 metric=ma.upload.attempt
2021-08-02T19:09:03-07:00 INF counters count=4 metric=ma.upload.success
2021-08-02T19:09:03-07:00 INF counters count=1 metric=ma.fsUploadable.skip.unsupported
2021-08-02T19:09:03-07:00 INF counters count=4 metric=ma.fsUploadable.open
2021-08-02T19:09:03-07:00 INF samples count=4 max=8.587274551391602 mean=6.179113388061523 metric=ma.upload.upload min=3.970703601837158 stddev=2.0252436802798743
```
