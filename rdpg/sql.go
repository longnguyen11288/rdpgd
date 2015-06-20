package rdpg

//uuid_generate_v1mc(), // uuid-ossp

var SQL map[string]string = map[string]string{
	"postgres_schemas": `
CREATE SCHEMA IF NOT EXISTS rdpg;
	`,
	"rdpg_schemas": `
CREATE SCHEMA IF NOT EXISTS rdpg;
CREATE SCHEMA IF NOT EXISTS pgbdr;
CREATE SCHEMA IF NOT EXISTS cfsbapi;
CREATE SCHEMA IF NOT EXISTS config;
CREATE SCHEMA IF NOT EXISTS backups;
CREATE SCHEMA IF NOT EXISTS metrics;
CREATE SCHEMA IF NOT EXISTS scheduler;
CREATE SCHEMA IF NOT EXISTS audit;
CREATE SCHEMA IF NOT EXISTS work;
`,
	"rdpg_extensions": `
CREATE EXTENSION IF NOT EXISTS btree_gist;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
`,
	"create_table_cfsbapi_services": `
CREATE TABLE IF NOT EXISTS cfsbapi.services (
  id               BIGSERIAL PRIMARY KEY NOT NULL,
  service_id       TEXT      DEFAULT gen_random_uuid(),
  name             TEXT,
  description      TEXT,
  bindable         boolean   DEFAULT TRUE,
  dashboard_client json,
  created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  effective_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  ineffective_at   TIMESTAMP
);
`,
	"create_table_cfsbapi_plans": `
CREATE TABLE IF NOT EXISTS cfsbapi.plans (
  id             BIGSERIAL    PRIMARY KEY NOT NULL,
  service_id     BIGINT       REFERENCES cfsbapi.services(id),
  plan_id        TEXT DEFAULT gen_random_uuid(),
  name           TEXT,
  description    TEXT,
  free           boolean   DEFAULT true,
  created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  effective_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  ineffective_at TIMESTAMP
);
`,
	"bdr_nodes":     `SELECT * FROM bdr.bdr_nodes;`,
	"bdr_nodes_dsn": `SELECT node_local_dsn FROM bdr.bdr_nodes;`,
	"create_table_cfsbapi_instances": `
CREATE TABLE IF NOT EXISTS cfsbapi.instances (
  id                BIGSERIAL PRIMARY KEY NOT NULL,
  instance_id       TEXT      NOT NULL,
  service_id        TEXT      NOT NULL,
  plan_id           TEXT      NOT NULL,
  organization_id   TEXT      NOT NULL,
  space_id          TEXT      NOT NULL,
  dbname            TEXT      NOT NULL,
  uname             TEXT      NOT NULL,
  pass              TEXT      NOT NULL,
  created_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  effective_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  ineffective_at    TIMESTAMP,
  decommissioned_at TIMESTAMP
);`,
	"create_table_cfsbapi_bindings": `
CREATE TABLE IF NOT EXISTS cfsbapi.bindings (
  id             BIGSERIAL PRIMARY KEY NOT NULL,
  instance_id    TEXT      NOT NULL,
  binding_id     TEXT      NOT NULL,
  created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  effective_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  ineffective_at TIMESTAMP
);`,
	"create_table_cfsbapi_credentials": `
CREATE TABLE IF NOT EXISTS cfsbapi.credentials (
  id             BIGSERIAL PRIMARY KEY NOT NULL,
  instance_id    TEXT      NOT NULL,
  binding_id     TEXT      NOT NULL,
  host           TEXT,
  port           TEXT,
  uname          TEXT,
  pass           TEXT,
  dbname         TEXT,
  created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  effective_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  ineffective_at TIMESTAMP
);`,
	"create_table_rdpg_consul_watch_notifications": `
CREATE TABLE IF NOT EXISTS rdpg.consul_watch_notifications ( 
  id BIGSERIAL NOT NULL, 
  host TEXT,
  msg TEXT,
  created_at TIMESTAMP DEFAULT NOW(), 
  CONSTRAINT consul_watch_notification_pkey PRIMARY KEY (id, host)
);`,
	"create_table_rdpg_events": `
CREATE TABLE IF NOT EXISTS rdpg.events ( 
  id BIGSERIAL NOT NULL PRIMARY KEY, 
  host TEXT,
  key TEXT,
  msg TEXT,
	created_at TIMESTAMP DEFAULT NOW()
);`,
	"create_table_work_tasks": `
CREATE TABLE IF NOT EXISTS work.tasks ( 
  id BIGSERIAL NOT NULL PRIMARY KEY, 
  task_id TEXT NOT NULL,
  action TEXT NOT NULL,
  data TEXT NOT NULL,
  locked_by text,
  created_at TIMESTAMP DEFAULT NOW(),
  processing_at TIMESTAMP,
  processed_at TIMESTAMP
);`,
	"create_table_rdpg_schedules": `
CREATE TABLE IF NOT EXISTS rdpg.schedules ( 
  id BIGSERIAL NOT NULL PRIMARY KEY, 
  schedule_id TEXT NOT NULL,
  action TEXT NOT NULL,
  data TEXT NOT NULL,
  enabled BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT NOW(),
  scheduling_at TIMESTAMP
  scheduled_at TIMESTAMP,
);`,
	"create_table_rdpg_config": `
CREATE TABLE IF NOT EXISTS rdpg.config ( 
		key text PRIMARY KEY,
		value text,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP
);`,
	"insert_default_rdpg_config": `
INSERT INTO rdpg.config (key,value)
VALUES 
("BackupsPath", "/var/vcap/store/pgbdr/backups"),
`,
	"insert_default_cfsbapi_services": `
INSERT INTO cfsbapi.services (name,description,bindable,dashboard_client)
VALUES ('rdpg', 'Reliable PostgrSQL Service', true, '{}') ;
`,
	"insert_default_cfsbapi_plans": `
INSERT INTO cfsbapi.plans (service_id,name,description,free) 
VALUES ((SELECT id AS svc_id FROM cfsbapi.services WHERE name='rdpg' LIMIT 1), 
'shared', 'A Reliable PostgreSQL database on a shared server.', true);
`,
	"create_function_rdpg_disable_database": `
CREATE OR REPLACE FUNCTION rdpg.bdr_disable_database(name text) RETURNS VOID
AS $func$
-- NOTE: This may only be run on the 'postgres' datbase
DECLARE
  r RECORD;
BEGIN
  IF name IN ('postgres','rdpg')
  THEN RETURN;
  END IF;

  UPDATE pg_database 
  SET datallowconn = 'false' 
  WHERE datname = name;

  EXECUTE 'ALTER DATABASE ' || name || ' OWNER TO postgres;';

  PERFORM pg_terminate_backend(pg_stat_activity.pid) 
  FROM pg_stat_activity 
  WHERE pg_stat_activity.datname = name
  AND pid <> pg_backend_pid();

  FOR r IN 
    SELECT slot_name 
    FROM pg_replication_slots 
    WHERE database = name 
  LOOP 
    PERFORM pg_drop_replication_slot(r.slot_name);
  END LOOP;
END;
$func$ LANGUAGE plpgsql;
`,
}
