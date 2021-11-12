package configure

import (
	"fmt"
	ini "gopkg.in/ini.v1"
	"strings"
)

const (
	SECT_CORE		= "CORE"
	SECT_TIMER		= "TIMER"
	SECT_NET		= "NET"
	SECT_LOG 		= "LOG"
	SECT_DATABASE	= "DATABASE"
	SECT_TEST		= "TEST"

	SECT_Lilpop = "Lilpop"	// Lilpop: insert your section name string
)

var (
	cfgFile *ini.File
	config  *Values
)

type Values struct {
	// Core
	Core *ValueCore
	Time *ValueTime
	Net  *ValueNet
	Log  *ValueLog
	Db   *ValueDatabase
	Test *ValueTest

	// Application
	Lilpop *ValueLilpop		// Lilpop: insert your structure variable
}

type ValueCore struct {
	sectionName string
}

type ValueTime struct {
	sectionName string

	IdleTimeout	uint64
}

type ValueNet struct {
	sectionName string

	EnableHttp		bool
	EnableWebsocket	bool
	EnableGrpc		bool
	EnableTcp 		bool
	EnableSerial	bool
	EnableMqueue	bool
}

type ValueLog struct {
	sectionName string

	EnableDebug	bool
	EnableInfo	bool
	EnableError	bool
	LogFile		string
	MaxSize 	int
	MaxBackups  int
	MaxAge      int
	LocalTime   bool
	Compress    bool
}

type ValueDatabase struct {
	sectionName string

	Type		string
	UserId		string
	Password	string
	IpAddress	string
	Port 		int
	DbName		string
	DbNameLog   string

	MongoAddr   string
	MongoPort	int
	MongoAuth	bool
	MongoId		string
	MongoPwd	string

	RedisAddr   string
	RedisPort   int
}

type ValueTest struct {
	sectionName 	string

	EnableSwagger	bool
}

func NewValues() *Values {
	if config == nil {
		config = &Values{}

		config.Core = &ValueCore{}
		config.Time = &ValueTime{}
		config.Net	= &ValueNet{}
		config.Log  = &ValueLog{}
		config.Db	= &ValueDatabase{}
		config.Test = &ValueTest{}

		config.Lilpop = &ValueLilpop{}	// Lilpop: insert your new structure
	}
	return config
}

func GetConfig() *Values {
	return config
}

func (c *Values) chkCfgFile(filePath string) (*Values, error) {
	if cfgFile == nil {
		var err error
		cfgFile, err = ini.Load(filePath)
		if err != nil {
			return nil, err
		}
		//fmt.Printf("Load INI file successed [%s]\n", filePath)
	}
	return c, nil
}

func (c *Values) GetValueALL(filePath string) (*Values, error) {
	var err error
	_, err = c.chkCfgFile(filePath)
	if err != nil {
		return nil, err
	}

	_, err = c.GetValueCore(filePath)
	if err != nil {
		return nil, err
	}
	_, err = c.GetValueTimer(filePath)
	if err != nil {
		return nil, err
	}
	_, err = c.GetValueNet(filePath)
	if err != nil {
		return nil, err
	}
	_, err = c.GetValueLog(filePath)
	if err != nil {
		return nil, err
	}
	_, err = c.GetValueDatabase(filePath)
	if err != nil {
		return nil, err
	}
	_, err = c.GetValueTest(filePath)
	if err != nil {
		return nil, err
	}

	_, err = c.GetValueLilpop(filePath)	// Lilpop: insert your option parsing fuction
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Values) GetValueCore(filePath string) (*Values, error) {
	_, err := c.chkCfgFile(filePath)
	if err != nil {
		return nil, err
	}

	c.Core.sectionName = SECT_CORE

	//fmt.Printf("Core Config Values: %+v\n", c.Log)

	return c, nil
}

func (c *Values) GetValueTimer(filePath string) (*Values, error) {
	_, err := c.chkCfgFile(filePath)
	if err != nil {
		return nil, err
	}

	c.Time.sectionName  = SECT_TIMER
	c.Time.IdleTimeout	= cfgFile.Section(SECT_TIMER).Key("IdleTimeout").MustUint64(10)

	//fmt.Printf("Time Config Values: %+v\n", c.Log)

	return c, nil
}

func (c *Values) GetValueNet(filePath string) (*Values, error) {
	_, err := c.chkCfgFile(filePath)
	if err != nil {
		return nil, err
	}

	c.Net.sectionName = SECT_NET
	c.Net.EnableHttp = cfgFile.Section(SECT_NET).Key("EnableHttp").MustBool(false)
	c.Net.EnableWebsocket = cfgFile.Section(SECT_NET).Key("EnableWebsocket").MustBool(false)
	c.Net.EnableGrpc = cfgFile.Section(SECT_NET).Key("EnableGrpc").MustBool(false)
	c.Net.EnableTcp = cfgFile.Section(SECT_NET).Key("EnableTcp").MustBool(false)
	c.Net.EnableSerial = cfgFile.Section(SECT_NET).Key("EnableSerial").MustBool(false)
	c.Net.EnableMqueue = cfgFile.Section(SECT_NET).Key("EnableMqueue").MustBool(false)

	//fmt.Printf("Net Config Values: %+v\n", c.Log)

	return c, nil
}

func (c *Values) GetValueLog(filePath string) (*Values, error) {
	_, err := c.chkCfgFile(filePath)
	if err != nil {
		return nil, err
	}

	c.Log.sectionName = SECT_LOG
	c.Log.EnableDebug = cfgFile.Section(SECT_LOG).Key("EnableDebug").MustBool(true)
	c.Log.EnableInfo  = cfgFile.Section(SECT_LOG).Key("EnableInfo").MustBool(true)
	c.Log.EnableError = cfgFile.Section(SECT_LOG).Key("EnableError").MustBool(true)
	c.Log.LogFile     = cfgFile.Section(SECT_LOG).Key("LogFile").MustString("sample.log")
	c.Log.MaxSize 	  = cfgFile.Section(SECT_LOG).Key("MaxSize").MustInt(100)
	c.Log.MaxBackups  = cfgFile.Section(SECT_LOG).Key("MaxBackups").MustInt(30)
	c.Log.MaxAge      = cfgFile.Section(SECT_LOG).Key("MaxAge").MustInt(30)
	c.Log.LocalTime   = cfgFile.Section(SECT_LOG).Key("LocalTime").MustBool(true)
	c.Log.Compress    = cfgFile.Section(SECT_LOG).Key("Compress").MustBool(false)

	//fmt.Printf("Log Config Values: %+v\n", c.Log)

	return c, nil
}

func (c *Values) GetValueDatabase(filePath string) (*Values, error) {
	_, err := c.chkCfgFile(filePath)
	if err != nil {
		return nil, err
	}

	c.Db.sectionName = SECT_DATABASE
	c.Db.Type 		 = cfgFile.Section(SECT_DATABASE).Key("Type").MustString("mysql")
	c.Db.UserId  	 = cfgFile.Section(SECT_DATABASE).Key("UserId").MustString("posgo")
	c.Db.Password 	 = cfgFile.Section(SECT_DATABASE).Key("Password").MustString("posgo")
	c.Db.IpAddress 	 = cfgFile.Section(SECT_DATABASE).Key("IpAddress").MustString("127.0.0.1")
	c.Db.Port 		 = cfgFile.Section(SECT_DATABASE).Key("Port").MustInt(-1)
	c.Db.DbName		 = cfgFile.Section(SECT_DATABASE).Key("DbName").MustString("posgo")
	c.Db.DbNameLog	 = cfgFile.Section(SECT_DATABASE).Key("DbNameLog").MustString("logs")

	c.Db.MongoAddr	 = cfgFile.Section(SECT_DATABASE).Key("MongoAddr").MustString("127.0.0.1")
	c.Db.MongoPort	 = cfgFile.Section(SECT_DATABASE).Key("MongoPort").MustInt(27017)
	c.Db.MongoAuth	 = cfgFile.Section(SECT_DATABASE).Key("MongoAuth").MustBool(false)
	c.Db.MongoId	 = cfgFile.Section(SECT_DATABASE).Key("MongoId").MustString("")
	c.Db.MongoPwd	 = cfgFile.Section(SECT_DATABASE).Key("MongoPwd").MustString("")

	c.Db.RedisAddr   = cfgFile.Section(SECT_DATABASE).Key("RedisAddr").MustString("127.0.0.1")
	c.Db.RedisPort   = cfgFile.Section(SECT_DATABASE).Key("RedisPort").MustInt(6379)

	//fmt.Printf("Dayabase Config Values: %+v\n", c.Db)

	return c, nil
}

func (c *Values) GetValueTest(filePath string) (*Values, error) {
	_, err := c.chkCfgFile(filePath)
	if err != nil {
		return nil, err
	}

	c.Db.sectionName = SECT_TEST
	c.Test.EnableSwagger	= cfgFile.Section(SECT_TEST).Key("EnableSwagger").MustBool(false)

	//fmt.Printf("Dayabase Config Values: %+v\n", c.Db)

	return c, nil
}

func (c *Values) PrintValues(sectName string) (prtStr string) {
	isAll := false

	sectName = strings.ToUpper(sectName)
	switch sectName {
	case SECT_CORE:
		prtStr += fmt.Sprintf(" * %s Config Values: [%+v]\n", sectName, c.Core)
	case SECT_TIMER:
		prtStr += fmt.Sprintf(" * %s Config Values: [%+v]\n", sectName, c.Time)
	case SECT_LOG:
		prtStr += fmt.Sprintf(" * %s Config Values: [%+v]\n", sectName, c.Log)
	case SECT_NET:
		prtStr += fmt.Sprintf(" * %s Config Values: [%+v]\n", sectName, c.Net)
	case SECT_DATABASE:
		prtStr += fmt.Sprintf(" * %s Config Values: [%+v]\n", sectName, c.Db)
	case SECT_TEST:
		prtStr += fmt.Sprintf(" * %s Config Values: [%+v]\n", sectName, c.Test)

	case SECT_Lilpop:	// application configure value
		prtStr += fmt.Sprintf(" * %s Config Values: [%+v]\n", sectName, c.Lilpop)
	default:
		isAll = true
	}

	if isAll {
		prtStr += fmt.Sprintf(" * %s Config Values: [%+v]\n", SECT_CORE, c.Core)
		prtStr += fmt.Sprintf(" * %s Config Values: [%+v]\n", SECT_TIMER, c.Time)
		prtStr += fmt.Sprintf(" * %s Config Values: [%+v]\n", SECT_LOG, c.Log)
		prtStr += fmt.Sprintf(" * %s Config Values: [%+v]\n", SECT_NET, c.Net)
		prtStr += fmt.Sprintf(" * %s Config Values: [%+v]\n", SECT_DATABASE, c.Db)
		prtStr += fmt.Sprintf(" * %s Config Values: [%+v]\n", SECT_TEST, c.Test)
		prtStr += fmt.Sprintf(" * %s Config Values: [%+v]\n", SECT_Lilpop, c.Lilpop)	// application configure value
	}
	return
}