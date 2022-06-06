# ma - CLI for managing photos locally and at SmugMug

All your media archiving needs!

## Global Flags
|Name|Aliases|Description|
|-|-|-|
|```smugmug-client-key```||smugmug client key|
|```smugmug-client-secret```||smugmug client secret|
|```smugmug-access-token```||smugmug access token|
|```smugmug-token-secret```||smugmug token secret|
|```json```|```j```|encode all results as JSON and print to stdout|
|```monochrome```||disable colored output|
|```debug```||enable debugging of http requests|
|```help```|```h```|show help|

## Commands
* [commands](#commands)
* [cp](#cp)
* [envvars](#envvars)
* [export](#export)
* [find](#find)
* [help](#help)
* [ls](#ls)
* [ls album](#ls-album)
* [ls image](#ls-image)
* [ls node](#ls-node)
* [new](#new)
* [new album](#new-album)
* [new folder](#new-folder)
* [patch](#patch)
* [patch album](#patch-album)
* [patch image](#patch-image)
* [rm](#rm)
* [rm image](#rm-image)
* [similar](#similar)
* [up](#up)
* [urlname](#urlname)
* [user](#user)
* [version](#version)

## *commands*

**Description**

Print all possible commands


**Syntax**

```sh
$ ma commands [flags]
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```description```|```d```||Print the command description as a comment|
|```relative```|```r```||Specify the command relative to the current working directory|


## *cp*

**Description**

copy files from a source(s) to a destination using the image date to layout the directory structure


**Syntax**

```sh
$ ma cp [flags] <file-or-directory> [, <file-or-directory>] <file-or-directory>
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```dryrun```|```n```||prepare to copy but don't actually do it|
|```format```|||the date format used for the destination directory|
|```concurrency```|```c```||the number of concurrent copy operations|


## *envvars*

**Description**

Useful for creating a .env file for all possible environment variables


**Syntax**

```sh
$ ma envvars [flags]
```



## *export*

**Description**

export images from albums to local disk


**Syntax**

```sh
$ ma export [flags] <node id> <directory>
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```concurrency```|```c```||the number of concurrent downloads|
|```force```|||overwrite existing files|


## *find*

**Description**

find albums or folders by name (if `--album` or `--node` is not specified, both will be searched)


**Syntax**

```sh
$ ma find [flags]
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```scope```|||root the search at the scope, if not specified the authenticated user's scope will be used|
|```album```|```a```||search only for albums|
|```node```|```n, f```||search only for nodes|

**Example**

`find` will look nodes for the specified query. If `--scope` is not specified, the currently
authenticated user's scope is used.

```sh
$ ma find --scope "/api/v2/user/cmac" Event
2021-10-24T18:38:52-07:00 INF find name=Events nodeID=kTR76 type=Folder
2021-10-24T18:38:52-07:00 INF find albumKey=mjXDhW imageCount=79 name="Harley Pan America event" nodeID=sBt5dp type=Album
2021-10-24T18:38:52-07:00 INF find albumKey=dDMKTz imageCount=6 name="SmugMug camera straps!" nodeID=vGtbt type=Album
2021-10-24T18:38:52-07:00 INF counters count=2 metric=ma.find.album
2021-10-24T18:38:52-07:00 INF counters count=1 metric=ma.find.node
```

```sh
$ ma find --scope "/api/v2/user/cmac" SmugMug
2021-10-24T18:40:21-07:00 INF find name=SmugMug nodeID=XWx8t type=Folder
2021-10-24T18:40:23-07:00 INF find albumKey=fqH3fM imageCount=16 name="SmugMug videos" nodeID=99SCh type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=TrBCmb imageCount=99 name="SmugMug Heroes" nodeID=Bpj6s type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=bKfV9Z imageCount=5 name="SmugMug Stickers!" nodeID=XWhVs type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=fk58cR imageCount=6 name="Powered by smugmug" nodeID=CxKrQ type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=Qp6Vk5 imageCount=35 name="SmugMug Halloween 2012" nodeID=NzST7 type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=dDMKTz imageCount=6 name="SmugMug camera straps!" nodeID=vGtbt type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=Q4JCgb imageCount=64 name="SmugMug Tahoe Soiree" nodeID=gQX9t type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=wKjKNd imageCount=1 name="SmugMug August 2012" nodeID=kDCZZ type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=9nmSBj imageCount=104 name="Fark/smugmug photoshop contest" nodeID=tZgNn type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=GxmcTJ imageCount=24 name="SmugMug homepage slide show" nodeID=SH5wj type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=bMh8jp imageCount=81 name="SmugMug Photo booth 2013" nodeID=kdCn8 type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=dgfgxg imageCount=4 name="SmugMug on an iPhone" nodeID=Q6qvZ type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=pXPZ4k imageCount=90 name="SmugMug Christmas party 2012" nodeID=wWSwG type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=D4nm6Z imageCount=6 name="Past smugmug home pages" nodeID=zLWT2 type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=XqxN46 imageCount=44 name="Tahoe Soiree" nodeID=VC5j4 type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=pZ9XFH imageCount=7 name="smugmug 1st annual Christmas brekkie" nodeID=KK9Dx type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=TJstLc imageCount=37 name="Possible smugmug home page thumbs" nodeID=cpVJr type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=243FWz imageCount=1 name="SmugMug Gym Shoot 2016 with Von Wong" nodeID=Kwz5kc type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=4hxRpd imageCount=0 name="My SmugMug Site Files (Do Not Delete)" nodeID=fPR7p type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=bT2KLC imageCount=129 name="Green hair!" nodeID=4n2G7 type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=WJvpCp imageCount=19 name="SmugMug's green hair" nodeID=7CMcv type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=NjqVwL imageCount=20 name="SmugMug's Windows 7 party with seven 7x7 burgers" nodeID=HcHRq type=Album
2021-10-24T18:40:23-07:00 INF counters count=1 metric=ma.find.node
2021-10-24T18:40:23-07:00 INF counters count=22 metric=ma.find.album
```


## *help*

**Description**

Shows a list of commands or help for one command


**Syntax**

```sh
$ ma help [flags] [command]
```



## *ls*

**Description**

list the deails of albums, nodes, and/or images



**Overview**

`ls` returns all the nodes under the specified parent node.

```sh
$ ma ls node -R kTR76
2021-10-24T18:41:37-07:00 INF ls name=Events nodeID=kTR76 parentID=zx4Fx type=Folder
2021-10-24T18:41:38-07:00 INF ls albumKey=jbBNhR imageCount=16 name="Black lives matter protest" nodeID=q2qP7F parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF ls albumKey=GNpNRf imageCount=11 name="47th annual electric vehicle show silicon valley" nodeID=6zFLgD parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF ls albumKey=LPZThB imageCount=21 name="Humans melt and pets soak up love at the Bay Area Pet Fair" nodeID=MG5c5r parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF ls albumKey=MGxZXq imageCount=36 name="San Francisco Reptile Expo" nodeID=kvh7p7 parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF ls albumKey=KJXpSD imageCount=25 name="Muttville adoption ceremony" nodeID=JbnGdJ parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF ls albumKey=pttBvK imageCount=18 name="SF Body Art Expo 2018" nodeID=xH9vPF parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF ls albumKey=gJ7d6r imageCount=35 name="Maker Faire 2008" nodeID=BjcPR parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF ls albumKey=2XrGxm imageCount=33 name="Santa Cruz fungus festival 2019" nodeID=jM739r parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF ls albumKey=B8BLsX imageCount=30 name="Golden Gate dog show 2019" nodeID=LtqS2s parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF ls albumKey=wqmpff imageCount=22 name="Spanish Fork High 50th reunion 2019" nodeID=5FD6k8 parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF ls albumKey=qH5CG3 imageCount=35 name="Spanish Fork Fiesta Days parade 2019" nodeID=J3wNMj parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF counters count=12 metric=ma.ls.node
```

If `ls` queries an album, only the album details are returned.

```sh
$ ma ls node -R q2qP7F
2021-10-24T18:42:52-07:00 INF ls albumKey=jbBNhR imageCount=16 name="Black lives matter protest" nodeID=q2qP7F parentID=kTR76 type=Album
2021-10-24T18:42:52-07:00 INF counters count=1 metric=ma.ls.node
```

To list the images for an album, use the `-i` flag.

```sh
$ ma ls node -R -i q2qP7F
2021-10-24T18:44:30-07:00 INF ls albumKey=jbBNhR imageCount=16 name="Black lives matter protest" nodeID=q2qP7F parentID=kTR76 type=Album
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--3.jpg imageKey=pjVHggG imageURI=/api/v2/album/jbBNhR/image/pjVHggG-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--6.jpg imageKey=Wzk7Lk5 imageURI=/api/v2/album/jbBNhR/image/Wzk7Lk5-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--7.jpg imageKey=mQRcX2V imageURI=/api/v2/album/jbBNhR/image/mQRcX2V-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--8.jpg imageKey=Vhhbw4j imageURI=/api/v2/album/jbBNhR/image/Vhhbw4j-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--9.jpg imageKey=3MgjDCT imageURI=/api/v2/album/jbBNhR/image/3MgjDCT-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--10.jpg imageKey=jfChpW8 imageURI=/api/v2/album/jbBNhR/image/jfChpW8-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--11.jpg imageKey=VDt5z9K imageURI=/api/v2/album/jbBNhR/image/VDt5z9K-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--12.jpg imageKey=S9jS4PW imageURI=/api/v2/album/jbBNhR/image/S9jS4PW-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--13.jpg imageKey=f6hWww5 imageURI=/api/v2/album/jbBNhR/image/f6hWww5-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--14.jpg imageKey=JdL4NbN imageURI=/api/v2/album/jbBNhR/image/JdL4NbN-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--15.jpg imageKey=F8QQhnN imageURI=/api/v2/album/jbBNhR/image/F8QQhnN-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--16.jpg imageKey=bvh5hgQ imageURI=/api/v2/album/jbBNhR/image/bvh5hgQ-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--17.jpg imageKey=g6F7kfS imageURI=/api/v2/album/jbBNhR/image/g6F7kfS-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--18.jpg imageKey=wd626B2 imageURI=/api/v2/album/jbBNhR/image/wd626B2-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--19.jpg imageKey=DLp2Mkt imageURI=/api/v2/album/jbBNhR/image/DLp2Mkt-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--20.jpg imageKey=B9C6jjx imageURI=/api/v2/album/jbBNhR/image/B9C6jjx-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF counters count=16 metric=ma.ls.image
2021-10-24T18:44:31-07:00 INF counters count=1 metric=ma.ls.node
```


## *ls album*

**Description**

list the contents of an album(s)


**Syntax**

```sh
$ ma ls album [flags] <album key> [<album key>, ...]
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```image```|```i, R```||include images in the query|


## *ls image*

**Description**

list the details of an image(s)


**Syntax**

```sh
$ ma ls image [flags] <image key> [<image key>, ...]
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```zero-version```|```z, 0```||if no version is specified, append `-0`|


## *ls node*

**Description**

list the contents of a node(s)


**Syntax**

```sh
$ ma ls node [flags] <node id> [<node id>, ...]
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```album```|```a```||include albums in the query|
|```node```|```n, f```||include nodes in the query|
|```image```|```i```||include images in the query|
|```recurse```|```R```||walk the node tree|
|```depth```|||walk the node tree to the specified depth|


## *new*

**Description**

create a new album or folder



**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```parent```|||the parent node at which the new node will be rooted|
|```privacy```|||the privacy settings for the new album|


## *new album*

**Description**

create a new album for images


**Syntax**

```sh
$ ma new album [flags]
```



## *new folder*

**Description**

create a new folder for albums


**Syntax**

```sh
$ ma new folder [flags]
```



## *patch*

**Description**

patch enables updating the metadata of both albums and images




## *patch album*

**Description**

patch the metadata of a single album


**Syntax**

```sh
$ ma patch album [flags] <album key> [<album key>, ...]
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```force```|```f```||force must be specified to apply the patch|
|```auto-urlname```|||if enabled, and an album name provided as a flag, the urlname will be auto-generated from the name|
|```keyword```|||a set of keywords describing the album|
|```name```|||the name of the album|
|```urlname```|||the urlname of the album (see `--auto-urlname` to set this automatically based on the album name)|


## *patch image*

**Description**

patch the metadata of an image (not the image itself though)


**Syntax**

```sh
$ ma patch image [flags] <image key> [<image key>, ...]
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```force```|```f```||force must be specified to apply the patch|
|```keyword```|||specifies keywords describing the image|
|```caption```|||the caption of the image|
|```title```|||the title of the image|
|```latitude```|||the latitude of the image location|
|```longitude```|||the longitude of the image location|
|```altitude```|||the altitude of the image location|


## *rm*

**Description**

delete an entity




## *rm image*

**Description**

delete an image from an album


**Syntax**

```sh
$ ma rm image [flags] IMAGE_KEY [, IMAGE-KEY, ...]
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```album```|||the album from which the image is to be deleted|
|```zero-version```|```z, 0```||if no version is specified, append `-0`|


## *similar*

**Description**

identify similar images


**Syntax**

```sh
$ ma similar [flags] FILE-OR-DIRECTORY, [FILE-OR-DIRECTORY, ...]
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```concurrency```|```c```||the number of concurrent image reads|


## *up*

**Description**

upload image files to the specified album, selectively including specific file extensions


**Syntax**

```sh
$ ma up [flags]
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```album```|```a```||the album to which image files will be uploaded|
|```ext```|```x```||the set of file extensions suitable for uploading|
|```dryrun```|```n```||prepare to upload but don't actually do it|

**Example**

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


## *urlname*

**Description**

create a clean urlname for each argument


**Syntax**

```sh
$ ma urlname [flags]
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```validate```|```a```||validate the url name|

**Example**

`urlname` displays an automatically generated `UrlName` suitable for SmugMug.

```sh
$ ma urlname "2021-10-31 Halloween Party"
2021-10-24T18:35:38-07:00 INF urlname name="2021-10-31 Halloween Party" url=2021-10-31-Halloween-Party
2021-10-24T18:35:38-07:00 INF counters count=1 metric=ma.urlname.urlname
```


## *user*

**Description**

query the authenticated user


**Syntax**

```sh
$ ma user [flags]
```



## *version*

**Description**

show the version information of the binary


**Syntax**

```sh
$ ma version [flags]
```


