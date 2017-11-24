package qn

import (
	"database/sql"
	"fmt"

	"github.com/douban-girls/qiniu-migrate/config"
)

func DbConnect() *sql.DB {
	cfg := config.GetConfig()
	dbPath := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", cfg.Host, cfg.Username, cfg.Pwd, cfg.Dbname)
	db, err := sql.Open("postgres", dbPath)
	config.ErrorHandle(err)
	return db
}

func GetImages(db *sql.DB) (result []*config.Cell) {
	rows, err := db.Query("SELECT id, src FROM cells WHERE cate=176 AND cate=179")
	config.ErrorHandle(err)
	defer rows.Close()

	for rows.Next() {
		var id int
		var src string
		if err := rows.Scan(&id, &src); err != nil {
			config.ErrorHandle(err)
		}

		result = append(result, &config.Cell{
			ID:  id,
			Src: src,
		})
	}
	return
}

func UpdateImage(db *sql.DB, cell *config.Cell) bool {
	_, err := db.Query("UPDATE cells SET img=%1 WHERE id=%2", cell.Src, cell.ID)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}
