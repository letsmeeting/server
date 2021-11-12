package userdb

import (
    gormdb "github.com/jinuopti/lilpop-server/database/gorm"
    . "github.com/jinuopti/lilpop-server/log"
    "gorm.io/gorm"
)

const UserTableName = "users"

const (
    Singer      = 1
    Actor       = 2
    Comedian    = 3
    Dancer      = 4
    Magician    = 5

    Etc         = 0
)

type User struct {
    gorm.Model
    UserId      string `gorm:"not null; uniqueIndex; size:80"`  // id or email
    Email       string
    Password    string `gorm:"not null"`
    Name        string `gorm:"not null; type:varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci"`
    Gender      string
    Country     string  // South Korea, America ...
    PhoneNumber string  // 010xxxxyyyy
    Birthday    string  // "1980-06-11"
    Picture     string  // Profile picture file
    Introduce   string `gorm:"type:varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci"`
    UserType    string `gorm:"not null;"`   // Entertainer or User or Admin
    AuthType    string  // "naver,kakao,google,email"
    Category    string  // Singer,Actor,Comedian,Dancer,Magician,Etc
}

func (User) TableName() string {
    return UserTableName
}

func DropTableUser() {
    if gormdb.MainDB == nil {
        Loge("gormdb.MainDB is nil")
        return
    }
    gormdb.MainDB.Db.Exec("DROP TABLE " + UserTableName)
}

func CreateTableUser() {
    var err error

    if gormdb.MainDB == nil {
        Logd("gormdb.MainDB is nil")
        return
    }

    err = gormdb.MainDB.Db.AutoMigrate(&User{})
    if err != nil {
        Loge("Failed auto migrate table : %s", err)
        return
    }
}

func InsertUser(user *User) error {
    gormdb.MainDB.Db.Create(user)
    return nil
}

func UpdateUser(user *User) error {
    gormdb.MainDB.Db.Where("user_id = ?", user.UserId).Updates(user)

    return nil
}

func DeleteUser(id string) error {
    gormdb.MainDB.Db.Where("user_id = ?", id).Unscoped().Delete(&User{})
    return nil
}

func GetUserList(offset int, limit int) (retUser []User) {
    var user []User

    if limit <= 0 {
        gormdb.MainDB.Db.Model(&User{}).Offset(offset).Find(&user)
    } else {
        gormdb.MainDB.Db.Model(&User{}).Limit(limit).Offset(offset).Find(&user)
    }

    for _, info := range user {
        retUser = append(retUser, info)
    }

    return retUser
}

func GetUserTotalCount() int64 {
    var count int64
    gormdb.MainDB.Db.Find(&User{}).Count(&count)
    return count
}

func FindByIdUser(id string) *User {
    if gormdb.MainDB == nil {
        Logd("gormdb.MainDB is nil")
        return nil
    }
    var user User
    first := gormdb.MainDB.Db.First(&user, "user_id = ?", id)

    if first.Error != nil {
        Logd("No match User Info")
        return nil
    }

    Logd("User UserId=%s, Password=%s, Name=%s, Email=%s, UserType=%s, Picture=%s, CreatedAt=%s, UpdatedAt=%s",
        user.UserId, user.Password, user.Name, user.Email, user.UserType, user.Picture,
        user.Model.CreatedAt.Format("2006-01-02 15:04:05"), user.Model.UpdatedAt.Format("2006-01-02 15:04:05"))

    return &user
}

func InitUserTable() {
    if gormdb.MainDB == nil {
        Logd("gormdb.MainDB is nil")
        return
    }
    CreateTableUser()
    Logd("DB: Init User Table")
}
