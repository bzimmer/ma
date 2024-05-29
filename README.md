# ma (media archiver)

![build](https://github.com/bzimmer/ma/actions/workflows/build.yaml/badge.svg)
[![codecov](https://codecov.io/gh/bzimmer/ma/branch/main/graph/badge.svg?token=J4JYIPRZUC)](https://codecov.io/gh/bzimmer/ma)

Simple tools for managing media files with [SmugMug](https://smugmug.com/)

Uses [smugmug](https://github.com/bzimmer/smugmug) for accessing the SmugMug [API](https://api.smugmug.com)

# usage

See the [manual](https://bzimmer.github.io/ma/commands) for an overview of all the commands.

```sh
NAME:
   ma - CLI for managing local and Smugmug-hosted photos

USAGE:
   ma [global options] command [command options] 

DESCRIPTION:
   CLI for managing local and Smugmug-hosted photos

COMMANDS:
   export        Export images from albums
   find, search  Search for albums or folders by name
   ls, list      List nodes, albums, and/or images
   new, create   Create a new node
   patch         patch the metadata of albums and images
   rm            Delete an entity
   similar       Identify similar images
   title         Create a title following the specified convention
   up, upload    Upload images to SmugMug
   urlname       Create a clean urlname for each argument
   user          Query the authenticated user
   version       Show the version information of the binary
   envvars       Print all the possible environment variables
   help, h       Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --smugmug-client-key value     smugmug client key [$SMUGMUG_CLIENT_KEY]
   --smugmug-client-secret value  smugmug client secret [$SMUGMUG_CLIENT_SECRET]
   --smugmug-access-token value   smugmug access token [$SMUGMUG_ACCESS_TOKEN]
   --smugmug-token-secret value   smugmug token secret [$SMUGMUG_TOKEN_SECRET]
   --json, -j                     emit all results as JSON and print to stdout (default: false)
   --monochrome                   disable colored loggingoutput (default: false)
   --debug                        enable verbose debugging (default: false)
   --trace                        enable debugging of http requests (default: false)
   --help, -h                     show help
```

## new album creation
An example to create a mirror of a top level directory structure:

```sh
$ fd -t d . /Volumes/Photos00/Scans -x ma new --parent bbHHmQ album {/.}
```
