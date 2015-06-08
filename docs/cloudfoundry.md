# RDPG Agent w/ Cloud Foundry

`rdpg-agent` is configured to listen on port 8888 by default (configurable) with
an http API listener that allows for [Cloud Foundry Service Broker]() functionality.

In development `rdpg-agent` can be registered with CF via,
```sh
cf create-service-broker rdpg cfadmin cfadmin http://10.244.2.2:8888
```

In production you will need to make sure that a domain name passes through to 
this backend port on any of the nodes, first node by default.


