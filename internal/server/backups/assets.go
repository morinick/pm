package backups

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func InsertAssetsToDB(db *sql.DB) error {
	assetsDir := "/assets"
	query := "insert into services (id, name, logo) values "
	args := []any{}

	assetsFiles, err := os.ReadDir(assetsDir)
	if err != nil {
		return fmt.Errorf("failed reading assets directory: %v", err)
	}

	if len(assetsFiles) == 0 {
		return nil
	}

	for _, logoFile := range assetsFiles {
		query += "(?, ?, ?), "
		args = append(args, uuid.New(), logoFile.Name(), filepath.Join(assetsDir, logoFile.Name()))
	}

	// Cut last ", "
	query = query[:len(query)-2]

	if _, err := db.Exec(query, args...); err != nil {
		return fmt.Errorf("failed query execution: %v", err)
	}

	return nil
}
