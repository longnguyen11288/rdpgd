# Development

This document explains how to setup a development environment.

## Environment 

To setup your development environment you need the following things,

### Golang

[go](http://golang.org)

```sh
cd /usr/local/src
curl -sOL https://storage.googleapis.com/golang/go1.4.2.src.tar.gz
tar zxvf go1.4.2.src.tar.gz
mv go go1.4.2
cd go1.4.2/src/
./all.bash
```

And then add go to your `PATH`,

```sh
echo 'PATH=${PATH}:/usr/local/go/bin' >> $HOME/.bash_profile
```

### Direnv

[direnv](http://direnv.net) is used for setting up the `GOPATH` via `Godeps` within the project.

Folow the installation instructions on the project website.

### GoConvey

[GoConvey](http://goconvey.co) is used for testing, follow the installation 
instructions on the website.

## Running rdpg-agent

Configuration of `rdpg-agent` is done via environment variables passed into the running process,

```sh
export \
  LOGLEVEL=debug \
  RDPG_SB_PORT=8888 \
  RDPG_SB_USER=cf \
  RDPG_SB_PASS=cf \
  RDPG_ADMIN_PORT=58888 \
  RDPG_PG_URI="postgresql://postgres:admin@127.0.0.1:55432/rdpg?sslmode=disable&connect_timeout=5&fallback_application_name=rdpg-agent" 
```

When running the agent locally, you will need to first deploy the 
`rdpg-boshrelease` and then forward the PostgreSQL from the release to your localhost,

```sh
ssh  -L 5432:127.0.0.1:55432 vcap@10.244.2.2 # BOSH Lite Password: c1oudc0w
```

To run the agent during development,

```sh
go run rdpg-agent.go
```
# Testing
Once GoConvey is running visit [The Web UI](http://127.0.0.1:8080)

Fetch catalog,

```sh
curl -vvv -H "X-Broker-API-Version: 2.4" http://cf:cf@127.0.0.1:8080/v2/catalog
```

For running a health check in development we can run,

```sh
curl -vvv http://admin:admin@127.0.0.1:58888/health/ha_pb_pg
```
