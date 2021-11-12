package configure

// exmaple structure
type ValueLilpop struct {
	sectionName string
	// Insert configure values here
	Enabled 				bool

	EnableHttpServer 		bool
	EnableWebsocketServer	bool
	EnableGrpcServer		bool
	EnableTcpServer	 		bool
	EnableSerial	 		bool
	EnableMqueue			bool
	EnableMongoDb           bool
	EnableRedis             bool

	EnableSSL				bool
	HttpListenAddr			string
	HttpServerPort 	 		string
	WebsocketServerPort		string
	SslCertFile				string
	SslKeyFile				string
	GrpcServerPort			string

	TcpServerPortLilpop 	string
	TcpClientPortLilpop 	string

	JanusAddr				string
	JanusHttpPort			string
	JanusWebsocketPort		string

	SwaggerPort				string

	// User login
	JwtAtExpiredTime 		int // JWT Access Token 만료시간 (sec)
	JwtRtExpiredTime 		int	// JWT Refresh Token 만료시간 (sec)

	// Google API
	GoogleClientId			string
	GoogleClientSecret		string
	GoogleCredentialFile 	string
	FirebaseCredentialFile  string
	FirebaseProjectId 		string

	// Scheduler
	SchedulerMinute         bool
	SchedulerHour           bool
	SchedulerDay            bool
}

func (c *Values) GetValueLilpop(filePath string) (*Values, error) {
	_, err := c.chkCfgFile(filePath)
	if err != nil {
		return nil, err
	}

	c.Lilpop.sectionName = SECT_Lilpop
	c.Lilpop.Enabled = cfgFile.Section(SECT_Lilpop).Key("Enabled").MustBool(false)

	// Insert configure values here
	c.Lilpop.EnableHttpServer = cfgFile.Section(SECT_Lilpop).Key("EnableHttpServer").MustBool(false)
	c.Lilpop.EnableWebsocketServer = cfgFile.Section(SECT_Lilpop).Key("EnableWebsocketServer").MustBool(false)
	c.Lilpop.EnableGrpcServer = cfgFile.Section(SECT_Lilpop).Key("EnableGrpcServer").MustBool(false)
	c.Lilpop.EnableTcpServer = cfgFile.Section(SECT_Lilpop).Key("EnableTcpServer").MustBool(false)
	c.Lilpop.EnableSerial = cfgFile.Section(SECT_Lilpop).Key("EnableSerial").MustBool(false)
	c.Lilpop.EnableMqueue = cfgFile.Section(SECT_Lilpop).Key("EnableMqueue").MustBool(false)
	c.Lilpop.EnableMongoDb = cfgFile.Section(SECT_Lilpop).Key("EnableMongoDb").MustBool(false)
	c.Lilpop.EnableRedis = cfgFile.Section(SECT_Lilpop).Key("EnableRedis").MustBool(false)

	c.Lilpop.EnableSSL = cfgFile.Section(SECT_Lilpop).Key("EnableSSL").MustBool(false)
	c.Lilpop.HttpListenAddr = cfgFile.Section(SECT_Lilpop).Key("HttpListenAddr").MustString("lilpop.kr")
	c.Lilpop.HttpServerPort = cfgFile.Section(SECT_Lilpop).Key("HttpServerPort").MustString("1323")
	c.Lilpop.WebsocketServerPort = cfgFile.Section(SECT_Lilpop).Key("WebsocketServerPort").MustString("1324")
	c.Lilpop.SslCertFile = cfgFile.Section(SECT_Lilpop).Key("SslCertFile").MustString("/etc/ssl/lilpop.crt")
	c.Lilpop.SslKeyFile = cfgFile.Section(SECT_Lilpop).Key("SslKeyFile").MustString("/etc/ssl/lilpop.key")
	c.Lilpop.GrpcServerPort = cfgFile.Section(SECT_Lilpop).Key("GrpcServerPort").MustString("1325")

	c.Lilpop.JanusAddr = cfgFile.Section(SECT_Lilpop).Key("JanusAddr").MustString("localhost")
	c.Lilpop.TcpServerPortLilpop = cfgFile.Section(SECT_Lilpop).Key("TcpServerPort_Lilpop").MustString("9900")
	c.Lilpop.TcpClientPortLilpop = cfgFile.Section(SECT_Lilpop).Key("TcpClientPort_Lilpop").MustString("9901")

	c.Lilpop.JanusHttpPort = cfgFile.Section(SECT_Lilpop).Key("JanusHttpPort").MustString("8088")
	c.Lilpop.JanusWebsocketPort = cfgFile.Section(SECT_Lilpop).Key("JanusWebsocketPort").MustString("8188")

	c.Lilpop.JwtAtExpiredTime = cfgFile.Section(SECT_Lilpop).Key("JwtAtExpiredTime").MustInt(60 * 60 * 24)
	c.Lilpop.JwtRtExpiredTime = cfgFile.Section(SECT_Lilpop).Key("JwtRtExpiredTime").MustInt(60 * 60 * 24 * 60)

	c.Lilpop.SwaggerPort = cfgFile.Section(SECT_Lilpop).Key("SwaggerPort").MustString("")

	c.Lilpop.GoogleClientId = cfgFile.Section(SECT_Lilpop).Key("GoogleClientId").MustString("")
	c.Lilpop.GoogleClientSecret = cfgFile.Section(SECT_Lilpop).Key("GoogleClientSecret").MustString("")
	c.Lilpop.GoogleCredentialFile = cfgFile.Section(SECT_Lilpop).Key("GoogleCredentialFile").
		MustString("/opt/lilpop/configure/google.json")
	c.Lilpop.FirebaseCredentialFile = cfgFile.Section(SECT_Lilpop).Key("FirebaseCredentialFile").
		MustString("/opt/lilpop/configure/firebase.json")
	c.Lilpop.FirebaseProjectId = cfgFile.Section(SECT_Lilpop).Key("FirebaseProjectId").MustString("lilpop-5e230")

	// Scheduler
	c.Lilpop.SchedulerMinute = cfgFile.Section(SECT_Lilpop).Key("SchedulerMinute").MustBool(false)
	c.Lilpop.SchedulerHour = cfgFile.Section(SECT_Lilpop).Key("SchedulerHour").MustBool(false)
	c.Lilpop.SchedulerDay = cfgFile.Section(SECT_Lilpop).Key("SchedulerDay").MustBool(false)

	return c, nil
}
