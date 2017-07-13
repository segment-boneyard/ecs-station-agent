CREATE TABLE IF NOT EXISTS tasks (
   task_arn               TEXT PRIMARY KEY,
   task_def_arn           TEXT,
   cluster_arn            TEXT,
   container_instance_arn TEXT,
   created_at             TIMESTAMP,
   started_at             TIMESTAMP,
   stopped_at             TIMESTAMP,
   stopped_reason         TEXT,
   desired_status         TEXT,
   last_status            TEXT,
   container_arn          TEXT,
   container_exit_code    INTEGER,
   container_last_status  TEXT,
   container_name         TEXT,
   version                INTEGER
);
