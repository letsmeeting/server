package database

import (
    . "github.com/jinuopti/lilpop-server/log"
    "github.com/jinuopti/lilpop-server/configure"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "fmt"
    "log"
    "gorm.io/gorm/logger"
    "time"
)

type DbInfo struct {
    Db       *gorm.DB
    Id       string
    Password string
    Ip       string
    Port     int
    DbName   string
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

    if dbInfo.Port < 0 {
        dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
            dbInfo.Id, dbInfo.Password, dbInfo.Ip, dbInfo.DbName)
    } else {
        dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
            dbInfo.Id, dbInfo.Password, dbInfo.Ip, dbInfo.Port, dbInfo.DbName)
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
        Loge("Failed to connect database %s : %s", dbInfo.DbName, err)
        return nil, err
    }
    return dbInfo.Db, nil
}

func (dbInfo *DbInfo) CreateDatabase() {
    if dbInfo.Db == nil || len(dbInfo.DbName) < 1 {
        return
    }
    dbInfo.Db.Exec("CREATE DATABASE IF NOT EXISTS " + dbInfo.DbName)
    dbInfo.Db.Exec("commit;")
}

func ConnNewDbFromConfig() (dbInfo *DbInfo, err error) {
    conf := configure.GetConfig()

    dbInfo = &DbInfo{
        Id:       conf.Db.UserId,
        Password: conf.Db.Password,
        Ip:       conf.Db.IpAddress,
        Port:     conf.Db.Port,
        DbName:   conf.Db.DbName,
    }

    _, err = dbInfo.ConnectDatabase()
    if err != nil {
        _, err = dbInfo.ConnectMysql()
        if err != nil {
            Loge("Failed to connect Database %s : %s", dbInfo.DbName, err)
            return nil, err
        }
        dbInfo.CreateDatabase()
    }
    return dbInfo, nil
}
