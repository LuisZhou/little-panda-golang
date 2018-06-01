// Package conf defines configuration of server.
package conf

type ModuleConfig struct {
	GoLen              int
	TimerDispatcherLen int
	AsynCallLen        int
	ChanRPCLen         int
	TimeoutAsynRet     int
}

type DbConfig struct {
	Default  string
	Mysql    string
	Postgres string
	Sqlite   string
	Mssql    string
}

var (
	LenStackBuf = 4096

	// log
	LogLevel string
	LogPath  string
	LogFlag  int

	// console
	ConsolePort   int
	ConsolePrompt string = "Lpge# "
	ProfilePath   string

	// cluster
	ListenAddr      string
	ConnAddrs       []string
	PendingWriteNum int

	// gate config
	GateConfig ModuleConfig

	// agent config
	AgentConfig ModuleConfig

	// function module config
	FunctionConfig ModuleConfig

	// db config
	DBConfig DbConfig
)
