package rdpg

var SQLStatements map[string]string = map[string]string {
	"create_extensions": `
CREATE EXTENSION IF NOT EXISTS btree_gist;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
`,
"create_table_services": `
CREATE TABLE IF NOT EXISTS rdpg.services(
  id               BIGSERIAL PRIMARY KEY NOT NULL,
  uuid             UUID DEFAULT gen_random_uuid(), -- uuid_generate_v1mc(),
  name             TEXT,
  description      TEXT,
  bindable         boolean DEFAULT TRUE,
  dashboard_client json,
  created_at       timestamp DEFAULT CURRENT_TIMESTAMP,
  effective_at     timestamp DEFAULT CURRENT_TIMESTAMP,
  ineffective_at   timestamp
);
`,
"create_table_plans": `
CREATE TABLE IF NOT EXISTS rdpg.plans(
  id             BIGSERIAL PRIMARY KEY NOT NULL,
  uuid           UUID DEFAULT   gen_random_uuid(), -- uuid_generate_v1mc(),
  service_id     UUID REFERENCES rdpg.services(id),
  name           TEXT,
  description    TEXT,
  free           boolean DEFAULT true,
  created_at     timestamp DEFAULT CURRENT_TIMESTAMP,
  effective_at   timestamp DEFAULT CURRENT_TIMESTAMP,
  ineffective_at timestamp
);
`,
"insert_default_services": `
INSERT INTO rdpg.services (name,description,bindable,dashboard_client)
VALUES ('rdpg', 'A Relilable Distributed PostgrSQL Service', true, '{}') ;
`,
"insert_default_plans": `
INSERT INTO rdpg.plans (service_id,name,description,free) 
VALUES ((SELECT id AS svc_id FROM rdpg.services WHERE name='rdpg' LIMIT 1), 
'small', 'A small shared reliable PostgreSQL database.', true);
`,
}

