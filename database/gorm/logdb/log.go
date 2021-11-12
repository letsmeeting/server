package logdb

import (
    "github.com/jinuopti/lilpop-server/configure"
    "github.com/jinuopti/lilpop-server/database/gorm"
    . "github.com/jinuopti/lilpop-server/log"
    "github.com/jinuopti/lilpop-server/restapi/v1"
    "gorm.io/gorm"
    "strings"
)

const LogTableName = "log"

const (
    Error = 0
    Warning = 1
    Info = 2
    Debug = 3
)

const (
    TAG_USER = "USER"
    TAG_WEBRTC = "WEBRTC"
    TAG_JANUS = "JANUS"
    TAG_HTTP = "HTTP"
    TAG_WS = "WEBSOCKET"
    TAG_DB = "DB"
)

type LogModel struct {
    gorm.Model
    LogType  string
    Severity int
    Message  string  `gorm:"type:varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci"`
}

func (LogModel) TableName() string {
    return LogTableName
}

func DropTableLogging() {
    if gormdb.LogDB == nil {
        Loge("Database is nil")
        return
    }
    gormdb.LogDB.Db.Exec("DROP TABLE " + LogTableName)
}

func CreateTableLogging() {
    if gormdb.LogDB == nil {
        Loge("gormdb.LogDB is nil")
        return
    }
    var err error
    err = gormdb.LogDB.Db.AutoMigrate(&LogModel{})
    if err != nil {
        Loge("Failed auto migrate table")
        return
    }
}

func WriteLog(log *LogModel) {
   gormdb.LogDB.Db.Create(log)
}

func FindBetweenTime(appName []string, offset int, limit int, severity int, start string, end string, keyword string) (
    logs []LogModel, totalCount int64, count int, errCode int) {
    if gormdb.LogDB == nil {
        Logd("gormdb.LogDB is nil")
        return nil, 0, 0, restapi.CodeErrDatabase
    }

    errCode = 0
    conf := configure.GetConfig()

    if len(start) <= 0 {
        start = "1970-01-01 00:00:00"
    }
    if len(end) <= 0 {
        end = "9999-12-31 23:59:59"
    }

    var appNameString string
    appNameString += "("
    for i := 0; i < len(appName); i++ {
        appName[i] = strings.ToLower(appName[i])
        if i == 0 {
            appNameString += "("
        } else {
            appNameString += " or ("
        }
        appNameString += " app_name like '%" + appName[i] + "%' )"
    }
    appNameString += ") AND "

    var severityString string
    if severity & 0x01 > 0 {
        severityString += "(severity = 0"
    }
    if severity & 0x02 > 0 {
        if len(severityString) > 0 {
            severityString += " or "
        } else {
            severityString += "("
        }
        severityString += "severity = 1 "
    }
    if severity & 0x04 > 0 {
        if len(severityString) > 0 {
            severityString += " or "
        } else {
            severityString += "("
        }
        severityString += "severity = 2 "
    }
    if severity & 0x08 > 0 {
        if len(severityString) > 0 {
            severityString += " or "
        } else {
            severityString += "("
        }
        severityString += "severity = 3 "
    }
    if len(severityString) > 0 {
        severityString += ") AND "
    }

    var keywordString string
    if len(keyword) > 0 {
        keywordString = "message like '%" + keyword + "%' AND "
    }

    var query string
    // query = "(SELECT id FROM " + conf.Db.DbNameLog + ".log WHERE " + appNameString + " " +
    //     severityString + " " + keywordString + "created_at BETWEEN '" + start + "' AND '" + end + "')"
    // gormdb.LogDB.Db.Model(&Total{}).Raw(query).Count(&totalCount)

    query = "(SELECT * FROM " + conf.Db.DbNameLog + ".log WHERE " + appNameString + " " +
        severityString + " " + keywordString + "created_at BETWEEN '" + start + "' AND '" + end + "')"
    query += " ORDER BY created_at desc"
    //query += " LIMIT " + strconv.Itoa(limit) + " OFFSET " + strconv.Itoa(offset)

    Logd("Offset:%d, Limit:%d, TotalCount:%d, Query: [%s]", offset, limit, totalCount, query)

    rows, err := gormdb.LogDB.Db.Raw(query).Rows()
    defer rows.Close()

    if err != nil {
        Logd("Error: %s", err)
    } else {
        for rows.Next() {
            var log LogModel
            _ = gormdb.LogDB.Db.ScanRows(rows, &log)
            logs = append(logs, log)
            count++
        }
    }

    if count == 0 {
        errCode = restapi.CodeEmpty
    }

    return logs, totalCount, count, errCode
}

func LogE(tag string, message string) {
    log := &LogModel{
        LogType: strings.ToUpper(tag),
        Severity: Error,
        Message: message,
    }
    WriteLog(log)
}

func LogW(tag string, message string) {
    log := &LogModel{
        LogType: strings.ToUpper(tag),
        Severity: Warning,
        Message: message,
    }
    WriteLog(log)
}

func LogI(tag string, message string) {
    log := &LogModel{
        LogType: strings.ToUpper(tag),
        Severity: Info,
        Message: message,
    }
    WriteLog(log)
}

func LogD(tag string, message string) {
    log := &LogModel{
        LogType: strings.ToUpper(tag),
        Severity: Debug,
        Message: message,
    }
    WriteLog(log)
}

func InitLoggingTable() {
    if gormdb.LogDB == nil {
        Logd("gormdb.LogDB is nil")
        return
    }
    CreateTableLogging()
    Logd("DB: Init Logging Table")
}
