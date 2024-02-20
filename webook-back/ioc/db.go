package ioc

import (
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(viper.GetString("db.dsn")))
	if err != nil {
		panic(err)
	}
	//err = dao.InitTables(db)
	//if err != nil {
	//    panic(err)
	//}
	return db
}
