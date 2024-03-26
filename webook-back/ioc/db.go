package ioc

import (
	"github.com/spf13/viper"
	prometheus2 "github.com/zht-account/webook/pkg/gormx/callbacks/prometheus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/prometheus"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(viper.GetString("db.dsn")))
	if err != nil {
		panic(err)
	}
	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          "webook",
		RefreshInterval: 15,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"Threads_running"},
			},
		},
	}))
	if err != nil {
		panic(err)
	}

	prom := prometheus2.Callbacks{
		Namespace:  "go_item",
		Subsystem:  "webook",
		Name:       "gorm",
		InstanceID: "my_instance_1",
		Help:       "gorm DB 查询",
	}
	err = prom.Register(db)
	if err != nil {
		panic(err)
	}

	//err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}
