package rdpg

//uuid_generate_v1mc(), // uuid-ossp

var SQL map[string]string = map[string]string{
	"rdpg_schemas": `
CREATE SCHEMA IF NOT EXISTS rdpg;
CREATE SCHEMA IF NOT EXISTS pgbdr;
CREATE SCHEMA IF NOT EXISTS cfsb;
CREATE SCHEMA IF NOT EXISTS config;
CREATE SCHEMA IF NOT EXISTS backups;
CREATE SCHEMA IF NOT EXISTS metrics;
`,
	"rdpg_extensions": `
CREATE EXTENSION IF NOT EXISTS btree_gist;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
`,
	"create_table_cfsb_services": `
CREATE TABLE IF NOT EXISTS cfsb.services (
  id               BIGSERIAL PRIMARY KEY NOT NULL,
  service_id       TEXT      DEFAULT gen_random_uuid(),
  name             TEXT,
  description      TEXT,
  bindable         boolean   DEFAULT TRUE,
  dashboard_client json,
  created_at       timestamp DEFAULT CURRENT_TIMESTAMP,
  effective_at     timestamp DEFAULT CURRENT_TIMESTAMP,
  ineffective_at   timestamp
);
`,
	"create_table_cfsb_plans": `
CREATE TABLE IF NOT EXISTS cfsb.plans (
  id             BIGSERIAL    PRIMARY KEY NOT NULL,
  service_id     BIGINT       REFERENCES cfsb.services(id),
  plan_id        TEXT DEFAULT gen_random_uuid(),
  name           TEXT,
  description    TEXT,
  free           boolean   DEFAULT true,
  created_at     timestamp DEFAULT CURRENT_TIMESTAMP,
  effective_at   timestamp DEFAULT CURRENT_TIMESTAMP,
  ineffective_at timestamp
);
`,
	"bdr_nodes":     `SELECT * FROM bdr.bdr_nodes;`,
	"bdr_nodes_dsn": `SELECT node_local_dsn FROM bdr.bdr_nodes;`,
	"create_table_cfsb_instances": `
CREATE TABLE IF NOT EXISTS cfsb.instances (
  id              BIGSERIAL PRIMARY KEY NOT NULL,
  instance_id     TEXT      NOT NULL,
  service_id      TEXT      NOT NULL,
  plan_id         TEXT      NOT NULL,
  organization_id TEXT      NOT NULL,
  space_id        TEXT      NOT NULL,
  dbname          TEXT      NOT NULL,
  uname           TEXT      NOT NULL,
  pass            TEXT      NOT NULL,
  created_at      timestamp DEFAULT CURRENT_TIMESTAMP,
  effective_at    timestamp DEFAULT CURRENT_TIMESTAMP,
  ineffective_at  timestamp
);`,
	"create_table_cfsb_bindings": `
CREATE TABLE IF NOT EXISTS cfsb.bindings (
  id             BIGSERIAL PRIMARY KEY NOT NULL,
  instance_id    TEXT      NOT NULL,
  binding_id     TEXT      NOT NULL,
  created_at     timestamp DEFAULT CURRENT_TIMESTAMP,
  effective_at   timestamp DEFAULT CURRENT_TIMESTAMP,
  ineffective_at timestamp
);`,
	"create_table_cfsb_credentials": `
CREATE TABLE IF NOT EXISTS cfsb.credentials (
  id             BIGSERIAL PRIMARY KEY NOT NULL,
  instance_id    TEXT      NOT NULL,
  binding_id     TEXT      NOT NULL,
  host           TEXT,
  port           TEXT,
  uname          TEXT,
  pass           TEXT,
  dbname         TEXT,
  created_at     timestamp DEFAULT CURRENT_TIMESTAMP,
  effective_at   timestamp DEFAULT CURRENT_TIMESTAMP,
  ineffective_at timestamp
);`,
	"insert_default_cfsb_services": `
INSERT INTO cfsb.services (name,description,bindable,dashboard_client)
VALUES ('rdpg', 'A Relilable Distributed PostgrSQL Service', true, '{}') ;
`,
	"insert_default_cfsb_plans": `
INSERT INTO cfsb.plans (service_id,name,description,free) 
VALUES ((SELECT id AS svc_id FROM cfsb.services WHERE name='rdpg' LIMIT 1), 
'small', 'A small shared reliable PostgreSQL database.', true);
`,
}
