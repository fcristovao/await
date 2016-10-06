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
    Usage: await [flag] [<dep>...] -- <cmd>
    Await availability of resources.

      -f	Force running the command even after giving up
      -t duration
        	Timeout duration before giving up (default 30s)
      -v	Set verbose output


## Resources

All dependant resources must be specified as URLs or escaped command. Some
resources provided additional functionally encoded as fragment (`#{{Args}}`).

Valid resources are:

- HTTP: `http[s]://[{{USER}}@]{{HOST}}[:{{PORT}}][{{PATH}}[?{{QUERY}}]]`
- Websocket: `ws[s]://[{{USER}}@]{{HOST}}[:{{PORT}}][{{PATH}}[?{{QUERY}}]]`
- TCP: `tcp[4|6]://{{HOST}}[:{{PORT}}]`
- File: `file://{{PATH}}#{{ABSENT}}`
- PostgreSQL: `postgres://{{DSN}}[#{{TABLES}}]`
- MySQL: `mysql://{{DSN}}[#{{TABLES}}]`
- Command: `{{PATH}}[ {{ARG}}...]`


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
