package startup

import (
	"context"
	"database/sql"
	"fmt"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
)

var db *gorm.DB

var mongoDB *mongo.Database

func InitTestDB() *gorm.DB {
	if db == nil {
		dsn := "root:18627502290@tcp(localhost:3306)/webook"
		sqlDB, err := sql.Open("mysql", dsn)
		if err != nil {
			panic(err)
		}
		for {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			err = sqlDB.PingContext(ctx)
			cancel()
			if err == nil {
				break
			}
			log.Println("等待连接 MYSQL", err)
		}
		db, err = gorm.Open(mysql.Open(dsn))
		if err != nil {
			panic(err)
		}
		//err = dao.InitTables(db)
		//if err != nil {
		//	panic(err)
		//}
	}
	return db
}

func InitMongoDB() *mongo.Database {
	if mongoDB == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		monitor := &event.CommandMonitor{
			Started: func(ctx context.Context,
				startedEvent *event.CommandStartedEvent) {
				fmt.Println(startedEvent.Command)
			},
		}
		opts := options.Client().
			ApplyURI("").
			SetMonitor(monitor)
		client, err := mongo.Connect(ctx, opts)
		if err != nil {
			panic(err)
		}
		mongoDB = client.Database("webook")
	}
	return mongoDB
}
