# ma (media archiver)

![build](https://github.com/bzimmer/ma/actions/workflows/build.yaml/badge.svg)
[![codecov](https://codecov.io/gh/bzimmer/ma/branch/main/graph/badge.svg?token=J4JYIPRZUC)](https://codecov.io/gh/bzimmer/ma)

Simple tools for managing media files with [SmugMug](https://smugmug.com/)

Uses [smugmug](https://github.com/bzimmer/smugmug) for accessing the SmugMug [API](https://api.smugmug.com)

# usage

See the [manual](docs/manual.md) for an overview of all the commands.

```sh
$ ma help
NAME:
   ma - CLI for managing local and Smugmug-hosted photos

USAGE:
   ma [global options] command [command options] [arguments...]

DESCRIPTION:
   CLI for managing local and Smugmug-hosted photos

COMMANDS:
   cp            Copy files to a date-formatted directory structure
   export        Export images from albums
   find, search  Search for albums or folders by name
   ls, list      List nodes, albums, and/or images
   new, create   Create a new node
   patch         patch the metadata of albums and images
   rm            Delete an entity
   similar       Identify similar images
   up, upload    Upload images to SmugMug
   urlname       Create a clean urlname for each argument
   user          Query the authenticated user
   version       Show the version information of the binary
   commands      Print all possible commands
   envvars       Print all the possible environment variables
   help, h       Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug                        enable debugging of http requests (default: false)
   --help, -h                     show help (default: false)
   --json, -j                     encode all results as JSON and print to stdout (default: false)
   --monochrome                   disable colored output (default: false)
   --smugmug-access-token value   smugmug access token (default: "") [$SMUGMUG_ACCESS_TOKEN]
   --smugmug-client-key value     smugmug client key (default: "") [$SMUGMUG_CLIENT_KEY]
   --smugmug-client-secret value  smugmug client secret (default: "") [$SMUGMUG_CLIENT_SECRET]
   --smugmug-token-secret value   smugmug token secret (default: "") [$SMUGMUG_TOKEN_SECRET]
```

## new album creation
An example to create a mirror of a top level directory structure:

```sh
$ fd -t d . /Volumes/Photos00/Scans -x ma new --parent bbHHmQ album {/.}
```
