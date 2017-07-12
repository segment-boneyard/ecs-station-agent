package main

import "database/sql"

// PostgresStore persists ECSTaskState objects to Postgres.
type PostgresStore struct {
	db *sql.DB
}

// DialPostgresStore returns an error if it is unable to connect to the given
// Postgres URL.
func DialPostgresStore(url string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", url)
	return &PostgresStore{db}, err
}

// Update inserts the given state in Postgres. If the version of the given
// state already exists, it does nothing.
func (s *PostgresStore) Update(state ECSTaskState) error {
	if len(state.Containers) == 0 {
		state.Containers = []ECSContainer{ECSContainer{}}
	}

	_, err := s.db.Exec(`
		INSERT INTO tasks
			(task_arn, task_def_arn, cluster_arn, container_instance_arn, created_at,
			started_at, stopped_at, stopped_reason, desired_status, last_status,
			container_arn, container_exit_code, container_last_status,
			container_name, version)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (task_arn, version) DO NOTHING
	`,
		state.TaskARN,
		state.TaskDefinitionARN,
		state.ClusterARN,
		state.ContainerInstanceARN,
		state.CreatedAt,
		state.StartedAt,
		state.StoppedAt,
		state.StoppedReason,
		state.DesiredStatus,
		state.LastStatus,
		state.Containers[0].ARN,
		state.Containers[0].ExitCode,
		state.Containers[0].LastStatus,
		state.Containers[0].Name,
		state.Version,
	)

	return err
}
