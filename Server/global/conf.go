package global

type (
	GameConfStruct struct {
		Game struct {
			Port      string `mapstructure:"port"`
			Connector string `mapstructure:"connector"`
			Serveridx uint32 `mapstructure:"serveridx"`
			HostName  string `mapstructure:"hostName"`
		} `mapstructure:"game"`
		Logger struct {
			Loglevel  string `mapstructure:"log-level"`
			LogPath   string `mapstructure:"log-root-path"`
			LogMaxAge string `mapstructure:"log-rotate-max-hour"`
			LogTime   string `mapstructure:"log-rotate-hour"`
			LogSize   string `mapstructure:"log-rotate-byte"`
			LogStdOut string `mapstructure:"log-stdout"`
			LogFile   string `mapstructure:"log-file"`
		} `mapstructure:"logger"`
		Postgres struct {
			Host     string `mapstructure:"host"`
			Port     string `mapstructure:"port"`
			User     string `mapstructure:"user"`
			Password string `mapstructure:"password"`
			DBName   string `mapstructure:"dbname"`
		} `mapstructure:"postgres"`
	}
)

var (
	GameConf GameConfStruct
)
