package swagger

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func DBConnect(connStr string) *sql.DB {

	db, err := sql.Open("postgres", connStr)
	CheckErrorFatal(err, "PostgreSQL connection error")

	err = db.Ping()
	CheckErrorFatal(err, "PostgreSQL DB connection status error")

	return db
}

func FillDb(db *sql.DB, pgTableName, jsonFilename string) {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '%s')", pgTableName)
	err := db.QueryRow(query).Scan(&exists)
	CheckErrorFatal(err, "Error querying table existence "+pgTableName)

	if !exists {
		log.Printf("Table %s does not exist, creating and populating table\n", pgTableName)

		// Чтение данных из JSON файла
		data, err := os.ReadFile(jsonFilename)
		CheckErrorFatal(err, "Error reading file "+jsonFilename)

		var records []FolderInfo
		err = json.Unmarshal(data, &records)
		CheckErrorFatal(err, "Error unmarshalling file "+jsonFilename)

		// Создание таблицы
		createTableQuery := fmt.Sprintf(`CREATE TABLE %s (
 id INTEGER PRIMARY KEY,
 parent_id INTEGER,
 name VARCHAR(255) NOT NULL,
 description TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_parent_id ON %s (parent_id);`, pgTableName, pgTableName)
		_, err = db.Exec(createTableQuery)
		CheckErrorFatal(err, "Error creating table "+pgTableName)
		log.Printf("Table %s created\n", pgTableName)
		log.Println("Data Population...")

		// Вставка данных в таблицу
		for _, record := range records {
			if record.ParentId == nil {
				insertQuery := fmt.Sprintf("INSERT INTO %s (id, name, description) VALUES ($1, $2, $3)", pgTableName)
				_, err = db.Exec(insertQuery, record.Id, record.Name, record.Description)
				CheckErrorFatal(err, "Error inserting row into "+pgTableName)
			}
		}

		// Вторая вставка с уже имеющимися parent_id
		for _, record := range records {
			if record.ParentId != nil {
				insertQuery := fmt.Sprintf("INSERT INTO %s (id, parent_id, name, description) VALUES ($1, $2, $3, $4)", pgTableName)
				_, err = db.Exec(insertQuery, record.Id, *record.ParentId, record.Name, record.Description)
				CheckErrorFatal(err, "Error inserting row into "+pgTableName)
			}
		}
		log.Printf("Table %s populated with data from %s\n", pgTableName, jsonFilename)
		log.Println("Index Creation...")

		alterTableQuery := fmt.Sprintf(`ALTER TABLE %s
ADD CONSTRAINT fk_parent
FOREIGN KEY (parent_id) 
REFERENCES %s(id)
ON UPDATE CASCADE ON DELETE SET NULL;`, pgTableName, pgTableName)
		_, err = db.Exec(alterTableQuery)
		CheckErrorFatal(err, "Error altering table "+pgTableName)

		log.Println("The indexes have been created")
	} else {
		log.Printf("Table %s already exists\n", pgTableName)
	}
}
