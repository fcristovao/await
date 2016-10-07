# await

Await availability of resources.

This can be useful in the context of
[Docker Compose](https://docs.docker.com/compose/) where service needs to wait
for other dependant services.

Optionally a timeout can be provided to specify how long to wait for all
dependant resources to become available. On success the command returns code `0`
on failure it returns code `1`.

Additionally a command can be specified which gets executed after all dependant
resources became available.

## Installation

    go get -u github.com/betalo-sweden/await


## Usage

    $ await -h
    Usage: await [options...] <res>... [-- <cmd>]
    Await availability of resources.

      -f	Force running the command even after giving up
      -q	Set quiet mode
      -t duration
        	Timeout duration before giving up (default 1m0s)
      -v	Set verbose output


## Resources

All dependant resources must be specified as URLs or escaped command.

Some resources provided additional functionally encoded as fragment
(`#{{Args}}`). The syntax has to conform to Go's _Struct Field Tag_ syntax:
`[key|key=val,...]` (no quoting supported).

Valid resources are: HTTP, Websocket, TCP, File, PostgreSQL, MySQL, Command.


### HTTP Resource

URL syntax: `http[s]://[{{USER}}@]{{HOST}}[:{{PORT}}][{{PATH}}[?{{QUERY}}]]`


### Websocket Resource

URL syntax: `ws[s]://[{{USER}}@]{{HOST}}[:{{PORT}}][{{PATH}}[?{{QUERY}}]]`


### TCP Resource

URL syntax: `tcp[4|6]://{{HOST}}[:{{PORT}}]`


### File Resource

URL syntax: `file://{{PATH}}[#{{FRAGMENT}}]`

- `absent` key: If present, the resource is defined as available, when the
  specific file is absent, rather than existing.


### PostgreSQL Resource

URL syntax: `postgres://{{DSN}}[#{{FRAGMENT}}]`

- `table[=t1,t2,...]` key-value: If key present and value absent, the resource's
  database scheme must at least contain one table. If key present and value
  present, the resource's database scheme must at least contain the specified
  tables.


### MySQL Resource


URL syntax: `mysql://{{DSN}}[#{{FRAGMENT}}]`

Fragment:

- `table[=t1,t2,...]` key-value: If key present and value absent, the resource's
  database scheme must at least contain one table. If key present and value
  present, the resource's database scheme must at least contain the specified
  tables.


### Command Resource

Does not follow the URL syntax and is used a generic fallback for invalid URLs
(i.e. absent scheme).

URL syntax: `{{PATH}} [{{ARG}}...]`


## Alternatives

Once Docker-Compose support the Docker `HEALTHCHECK` directive for awaiting
services being up, this tool could become deprecated.

Many alternative solutions exist, most of them avoiding to interpret the
resource type, leaving only options like tcp and http, or having a specific
focus on Docker Compose. A few of them listed below:

- [ContainerPilot](https://github.com/joyent/containerpilot)
- [controlled-compose](https://github.com/dansteen/controlled-compose)
- [crane](https://github.com/michaelsauter/crane)
- [docker-wait](https://github.com/aanand/docker-wait)
- [dockerize](https://github.com/jwilder/dockerize)
- [wait-for-it](https://github.com/vishnubob/wait-for-it)
- [wait_for_db](https://gitlab.com/thelabnyc/ci/blob/09504268779acf53d65383b98b76e44ff763ef4d/examples/docker-compose-links/entrypoint.sh)
- [wait_to_start.sh](https://gist.github.com/rochacbruno/bdcad83367593fd52005#file-wait_to_start-sh)
- [waitforservices](https://github.com/Barzahlen/waitforservices)
