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
		} `mapstructure:"logger"`
	}
)

var (
	GameConf GameConfStruct
)
