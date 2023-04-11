# Home

![build](https://github.com/bzimmer/ma/actions/workflows/build.yaml/badge.svg)
[![codecov](https://codecov.io/gh/bzimmer/ma/branch/main/graph/badge.svg?token=J4JYIPRZUC)](https://codecov.io/gh/bzimmer/ma)

## Introduction

Simple tools for managing media files with [SmugMug](https://smugmug.com/)

Uses [smugmug](https://github.com/bzimmer/smugmug) for accessing the SmugMug [API](https://api.smugmug.com)

## Installation

```shell
$ brew install bzimmer/tap/ma
```

## Examples

``` shell title="An example to create a mirror of a top level directory structure"
$ fd --max-depth 1 -t d . /Volumes/Photos00/Scans -x ma new --parent bbHHmQ album {/.}
```
