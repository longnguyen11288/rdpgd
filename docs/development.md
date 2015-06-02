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
firstNode=10.244.2.2
export \
  RDPG_SB_PORT=8888 \
  RDPGAPI_SB_USER=cf \
  RDPGAPI_SB_PASS=cf \
  RDPG_ADMIN_PORT=58888 \
  RDPGAPI_PG_URI="postgresql://postgres:admin@${firstNode}:6432/rdpg?fallback_application_name=rdpg-agent&connect_timeout=5&sslmode=disable" 
```

To run the agent during development,

```sh
go run rdpg-agent.go
```
### Testing
Once GoConvey is running visit [The Web UI](http://127.0.0.1:8080)

