package ioc

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:18627502290@tcp(localhost:3306)/test"))
	if err != nil {
		panic(err)
	}
	//err = dao.InitTables(db)
	//if err != nil {
	//    panic(err)
	//}
	return db
}
