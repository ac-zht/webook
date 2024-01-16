package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var ErrUserDuplicate = errors.New("用户邮箱或者手机号冲突")

type UserDAO interface {
	Insert(ctx context.Context, u User) error
	FindById(ctx context.Context, id int64) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	UpdateNonZeroFields(ctx context.Context, u User) error
}

type GORMUserDAO struct {
	db *gorm.DB
}

type User struct {
	Id       int64          `gorm:"primaryKey,autoIncrement"`
	Email    sql.NullString `gorm:"unique"`
	Password string
	Phone    sql.NullString `gorm:"unique"`
	Nickname sql.NullString
	Birthday sql.NullInt64
	AboutMe  sql.NullString `gorm:"column:about_me;type:varchar(1024)"`

	Ctime int64
	Utime int64
}

func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &GORMUserDAO{
		db: db,
	}
}

func (ud *GORMUserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	err := ud.db.WithContext(ctx).Create(&u).Error
	if me, ok := err.(*mysql.MySQLError); ok {
		const uniqueIndexErrNo uint16 = 1062
		if me.Number == uniqueIndexErrNo {
			return ErrUserDuplicate
		}
	}
	return err
}

func (ud *GORMUserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var user User
	err := ud.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (ud *GORMUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := ud.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (ud *GORMUserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var user User
	err := ud.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (ud *GORMUserDAO) UpdateNonZeroFields(ctx context.Context, u User) error {
	return ud.db.Updates(&u).Error
}
