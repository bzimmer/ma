# ma (media archiver)

![build](https://github.com/bzimmer/ma/actions/workflows/build.yaml/badge.svg)
[![codecov](https://codecov.io/gh/bzimmer/ma/branch/main/graph/badge.svg?token=J4JYIPRZUC)](https://codecov.io/gh/bzimmer/ma)

Simple tools for managing media files with [SmugMug](https://smugmug.com/)

Uses [smugmug](https://github.com/bzimmer/smugmug) for accessing the SmugMug [API](https://api.smugmug.com)

# usage

See the [manual](docs/manual.md) for an overview of all the commands.

```sh
~/Development/src/github.com/bzimmer/ma (descriptions) > ma help
NAME:
   ma - CLI for managing photos locally and at SmugMug

USAGE:
   ma [global options] command [command options] [arguments...]

DESCRIPTION:
   CLI for managing photos locally and at SmugMug

COMMANDS:
   cp          copy files to a pre-determined directory structure
   export      export images from albums
   find        search for albums or folders by name
   ls, list    list nodes, albums, and/or images
   new         create a new node
   patch       patch the metadata for albums and images
   up, upload  upload images to SmugMug
   urlname     Create a clean url name for the argument
   user        query the authenticated user
   version     version information
   commands    Print all possible commands
   envvars     Print all the possible environment variables
   help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --smugmug-client-key value     smugmug client key (default: "") [$SMUGMUG_CLIENT_KEY]
   --smugmug-client-secret value  smugmug client secret (default: "") [$SMUGMUG_CLIENT_SECRET]
   --smugmug-access-token value   smugmug access token (default: "") [$SMUGMUG_ACCESS_TOKEN]
   --smugmug-token-secret value   smugmug token secret (default: "") [$SMUGMUG_TOKEN_SECRET]
   --concurrency value            (default: 2)
   --json, -j                     (default: false)
   --debug                        enable debugging (default: false)
   --help, -h                     show help (default: false)
```
