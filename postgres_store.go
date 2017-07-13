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

// Update saves the given task state in Postgres only if the version of the
// task state is greater than what is stored or if the task state hasn't been
// stored yet.
func (s *PostgresStore) Update(state ECSTaskState) error {
	if len(state.Containers) == 0 {
		state.Containers = []ECSContainer{ECSContainer{}}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var needsUpdate bool
	err = tx.QueryRow("SELECT 1 FROM tasks WHERE task_arn=$1 AND version < $2", state.TaskARN, state.Version).Scan(&needsUpdate)
	if err != nil || !needsUpdate {
		return err
	}

	_, err = tx.Exec("DELETE FROM tasks WHERE task_arn=$1", state.TaskARN)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO tasks
			(task_arn, task_def_arn, cluster_arn, container_instance_arn, created_at,
			started_at, stopped_at, stopped_reason, desired_status, last_status,
			container_arn, container_exit_code, container_last_status,
			container_name, version)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
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
	if err != nil {
		return err
	}

	return tx.Commit()
}
