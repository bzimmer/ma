# ma - CLI for managing photos locally and at SmugMug

All your media archiving needs!

## Global Flags
|Name|Aliases|Description|
|-|-|-|
|```smugmug-client-key```||smugmug client key|
|```smugmug-client-secret```||smugmug client secret|
|```smugmug-access-token```||smugmug access token|
|```smugmug-token-secret```||smugmug token secret|
|```concurrency```|||
|```json```|```j```||
|```debug```||enable debugging|
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

copy files to a pre-determined directory structure


**Syntax**

```sh
$ ma cp [flags] <file-or-directory> [, <file-or-directory>] <file-or-directory>
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```dryrun```|```n```|||
|```format```||||
|```concurrency```|```c```||the number of concurrent copies|


## *envvars*

**Description**

Useful for creating a .env file for all possible environment variables


**Syntax**

```sh
$ ma envvars [flags]
```



## *export*

**Description**

export images from albums


**Syntax**

```sh
$ ma export [flags] <node id> <directory>
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```concurrency```|||the number of concurrent downloads|
|```force```|||overwrite existing files|


## *find*

**Description**

search for albums or folders by name


**Syntax**

```sh
$ ma find [flags]
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```scope```||||
|```album```|```a```|||
|```node```|```n, f```|||


## *help*

**Description**

Shows a list of commands or help for one command


**Syntax**

```sh
$ ma help [flags] [command]
```



## *ls*

**Description**

list nodes, albums, and/or images




## *ls album*

**Description**

list albums


**Syntax**

```sh
$ ma ls album [flags] <album key> [<album key>, ...]
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```image```|```i, R```|||


## *ls image*

**Description**

list images


**Syntax**

```sh
$ ma ls image [flags] <image key> [<image key>, ...]
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```zero-version```|```z```|||


## *ls node*

**Description**

list nodes


**Syntax**

```sh
$ ma ls node [flags] <node id> [<node id>, ...]
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```album```|```a```|||
|```node```|```n, f```|||
|```image```|```i```|||
|```recurse```|```R```|||
|```depth```||||


## *new*

**Description**

create a new node



**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```parent```||||
|```privacy```||||


## *new album*

**Description**




**Syntax**

```sh
$ ma new album [flags]
```



## *new folder*

**Description**




**Syntax**

```sh
$ ma new folder [flags]
```



## *patch*

**Description**

patch the metadata for albums and images




## *patch album*

**Description**

patch an album (or albums)


**Syntax**

```sh
$ ma patch album [flags] <album key> [<album key>, ...]
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```force```|```f```||force must be specified to apply the patch|
|```auto-urlname```|||if enabled, and an album name provided as a flag, the urlname will be auto-generated from the name|
|```keyword```||||
|```name```||||
|```urlname```||||


## *patch image*

**Description**

patch an image (or images)


**Syntax**

```sh
$ ma patch image [flags] <image key> [<image key>, ...]
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```force```|```f```||force must be specified to apply the patch|
|```keyword```||||
|```caption```||||
|```title```||||
|```latitude```||||
|```longitude```||||
|```altitude```||||


## *up*

**Description**

upload images to SmugMug


**Syntax**

```sh
$ ma up [flags]
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|```album```|```a```|||
|```ext```|```x```|||
|```dryrun```|```n```|||


## *urlname*

**Description**

Create a clean url for the argument by removing "unpleasant" values such as `'s` and `-`


**Syntax**

```sh
$ ma urlname [flags]
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

version information


**Syntax**

```sh
$ ma version [flags]
```


