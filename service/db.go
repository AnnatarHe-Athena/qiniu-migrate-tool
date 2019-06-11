package service

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"strings"
	"github.com/douban-girls/qiniu-migrate/config"
	_ "github.com/lib/pq"
)

const baseWhereCond = " FROM cells WHERE img not like '%qn://%' "

var db *sql.DB

func errorChecker(err error) {
	if err != nil {
		fmt.Println(err.Error())
		db.Close()
		panic(err)
	}
}

func DbConnect() *sql.DB {
	dbPath := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", config.Host, config.Username, config.Pwd, config.Dbname)
	dbInstance, err := sql.Open("postgres", dbPath)
	// dbInstance.SetMaxOpenConns(16)
	db = dbInstance
	errorChecker(err)
	return db
}

func GetImageLen(normal bool) (count int) {
	prefix := "SELECT count(id)"
	condition := getWhereCondition(normal)
	result, err := db.Query(prefix + condition)
	defer result.Close()
	errorChecker(err)
	for result.Next() {
		if err := result.Scan(&count); err != nil {
			errorChecker(err)
		}
	}
	return
}
func getWhereCondition(normal bool) string {
	condition := baseWhereCond + "AND premission="
	if normal {
		condition += "2"
	} else {
		condition += "3"
	}
	if normal {
		return condition
	}
	return strings.Replace(condition, "not ", "", -1)
}

func GetImages(result chan *config.Cell, count int, normal bool) {
	times := int(math.Ceil(float64(count) / 1000))
	condition := getWhereCondition(normal)

	stmt, err := db.Prepare("SELECT id, img" + condition + "ORDER BY id ASC LIMIT 1000 OFFSET $1")
	errorChecker(err)
	for i := 0; i < times; i++ {
		rows, err := stmt.Query(1000 * i)
		errorChecker(err)
		defer rows.Close()

		if err := rows.Err(); err != nil {
			errorChecker(err)
		}
		for rows.Next() {
			var id int
			var src string
			if err := rows.Scan(&id, &src); err != nil {
				errorChecker(err)
			}
			cell := &config.Cell{
				ID:  id,
				Src: src,
				Md5: sha256hex(src),
			}
			result <- cell
		}
	}
	defer stmt.Close()
	close(result)
}

func UpdateImage(cell *config.Cell) (rep bool) {
	rows, err := db.Query("UPDATE cells SET img=$1, md5=$2 WHERE id=$3", cell.Src, cell.Md5, cell.ID)
	if err != nil {
		log.Println("update error: ", err, cell.Src)
		log.Println("md5: ", cell.Md5)
		rep = false
		return
	}
	rep = true
	// 手动关闭
	rows.Close()
	return
}

func DeleteRecord(cell *config.Cell) bool {
	rows, err := db.Query("DELETE FROM cells WHERE id=$1", cell.ID)
	errorChecker(err)
	defer rows.Close()
	return true
}

func DeleteRecordSoft(cell *config.Cell) bool {
	_, err := db.Exec("UPDATE cells SET premission=5 WHERE id=$1", cell.ID)
	if err != nil {
		log.Println("soft delete cell has error: ", err)
		return false
	}
	return true
}