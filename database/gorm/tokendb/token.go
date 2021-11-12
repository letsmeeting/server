package tokendb

import (
    gormdb "github.com/jinuopti/lilpop-server/database/gorm"
    . "github.com/jinuopti/lilpop-server/log"
    "gorm.io/gorm"
    "time"
)

const TokenTableName = "tokens"

type Token struct {
    gorm.Model
    UserId              string `gorm:"not null; uniqueIndex; size:80"`  // id or email
    AccessToken         string
    AccessTokenTime     time.Time
    RefreshToken        string
    RefreshTokenTime    time.Time
}

func (Token) TableName() string {
    return TokenTableName
}

func DropTable() {
    if gormdb.MainDB == nil {
        Loge("gormdb.MainDB is nil")
        return
    }
    gormdb.MainDB.Db.Exec("DROP TABLE " + TokenTableName)
}

func CreateTable() {
    var err error

    if gormdb.MainDB == nil {
        Logd("gormdb.MainDB is nil")
        return
    }

    err = gormdb.MainDB.Db.AutoMigrate(&Token{})
    if err != nil {
        Loge("Failed auto migrate table : %s", err)
        return
    }
}

func InsertToken(token *Token) error {
    r := gormdb.MainDB.Db.Create(token)
    if r.Error != nil {
        return r.Error
    }
    return nil
}

func UpdateToken(token *Token) error {
    gormdb.MainDB.Db.Where("user_id = ?", token.UserId).Updates(token)
    return nil
}

func FindToken(userId string) *Token {
    var token Token
    f := gormdb.MainDB.Db.First(&token, "user_id = ?", userId)
    if f.Error != nil {
        Logd("Not found user_id %s", userId)
        return nil
    }
    return &token
}

func InitTable() {
    if gormdb.MainDB == nil {
        Logd("gormdb.MainDB is nil")
        return
    }
    CreateTable()
    Logd("DB: Init Token Table")
}