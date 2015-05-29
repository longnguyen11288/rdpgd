# Reliable Distributed PostgreSQL AIP with CF Service Broker API

## Development
```
export \
  RDPGAPI_SB_USER=cf \
  RDPGAPI_SB_PASS=cf \
  RDPGAPI_PGURI="postgresql://postgres:pgbdr@10.244.2.2:6432/pgbdr?fallback_application_name=rdpg-agent&connect_timeout=5&sslmode=disable" 
go run rdpg-agent.go
```

## TODOs
* Add an endpoint which allows for dynamically adding a CF.
* Enabling replication on creation of database.
* Monitoring background worker

