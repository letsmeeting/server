package gormdb

import (
    "fmt"
    "log"
    "time"

    "github.com/jinuopti/lilpop-server/configure"
    . "github.com/jinuopti/lilpop-server/log"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

type DbInfo struct {
    Db        *gorm.DB
    Id        string
    Password  string
    Ip        string
    Port      int
    DbName    string
    DbNameLog string
}

var (
    MainDB *DbInfo
    LogDB  *DbInfo
)

func InitSingletonDB() error {
    conf := configure.GetConfig()

    if MainDB == nil {
        MainDB = &DbInfo{
            Id:       conf.Db.UserId,
            Password: conf.Db.Password,
            Ip:       conf.Db.IpAddress,
            Port:     conf.Db.Port,
            DbName:   conf.Db.DbName,
        }
        err := ConnectDatabase(MainDB)
        if err != nil {
            return err
        }
    }

    if LogDB == nil {
        LogDB = &DbInfo{
            Id:        conf.Db.UserId,
            Password:  conf.Db.Password,
            Ip:        conf.Db.IpAddress,
            Port:      conf.Db.Port,
            DbName:    conf.Db.DbName,
            DbNameLog: conf.Db.DbNameLog,
        }
        err := ConnectDatabase(LogDB)
        if err != nil {
            return err
        }
    }

    Logd("Success Init Singleton DB")

    return nil
}

func (dbInfo *DbInfo) ConnectMysql() (db *gorm.DB, err error) {
    var dsn string

    if dbInfo.Port < 0 {
        dsn = fmt.Sprintf("%s:%s@tcp(%s)/", dbInfo.Id, dbInfo.Password, dbInfo.Ip)
    } else {
        dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/", dbInfo.Id, dbInfo.Password, dbInfo.Ip, dbInfo.Port)
    }

    dbInfo.Db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.New(
            log.New(GetLogWriter(), "\n", log.LstdFlags), // io writer
            logger.Config{
                SlowThreshold: 200 * time.Millisecond,
                LogLevel:      logger.Warn,
                Colorful:      true,
            }),
    })
    if err != nil {
        Loge("Failed to connect database : %s", err)
        return nil, err
    }

    return dbInfo.Db, nil
}

func (dbInfo *DbInfo) ConnectDatabase() (db *gorm.DB, err error) {
    var dsn string
    var dbName string

    if len(dbInfo.DbNameLog) > 0 {
        dbName = dbInfo.DbNameLog
    } else {
        dbName = dbInfo.DbName
    }

    if dbInfo.Port < 0 {
        dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
            dbInfo.Id, dbInfo.Password, dbInfo.Ip, dbName)
    } else {
        dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
            dbInfo.Id, dbInfo.Password, dbInfo.Ip, dbInfo.Port, dbName)
    }

    dbInfo.Db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.New(
            log.New(GetLogWriter(), "\n", log.LstdFlags), // io writer
            logger.Config{
                SlowThreshold: 200 * time.Millisecond,
                LogLevel:      logger.Warn,
                Colorful:      true,
            }),
    })
    if err != nil {
        Loge("Failed to connect database %s : %s", dbName, err)
        return nil, err
    }
    return dbInfo.Db, nil
}

func (dbInfo *DbInfo) CreateDatabase() {
    if dbInfo.Db == nil {
        return
    }
    var dbName string
    if len(dbInfo.DbNameLog) > 0 {
        dbName = dbInfo.DbNameLog
    } else {
        dbName = dbInfo.DbName
    }
    dbInfo.Db.Exec("CREATE DATABASE IF NOT EXISTS " + dbName + " DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci")
    dbInfo.Db.Exec("commit;")
    Logd("Create DATABASE %s", dbName)
}

func ConnectDatabase(dbInfo *DbInfo) (err error) {
    _, err = dbInfo.ConnectDatabase()
    if err != nil {
        _, err = dbInfo.ConnectMysql()
        if err != nil {
            Loge("Failed to connect Database : %s", err)
            return err
        }
        dbInfo.CreateDatabase()
        _, err = dbInfo.ConnectDatabase()
        if err != nil {
            Loge("Error: Create DB but failed to connect DB")
            return err
        }
    }
    return nil
}
