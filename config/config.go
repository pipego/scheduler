package config

type Config struct {
	ApiVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	MetaData   MetaData `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

type MetaData struct {
	Name string `yaml:"name"`
}

type Spec struct {
	Fetch  Plugin `yaml:"fetch"`
	Filter Plugin `yaml:"filter"`
	Score  Plugin `yaml:"score"`
	Logger Logger `yaml:"logger"`
}

type Plugin struct {
	Disabled []Disabled `yaml:"disabled"`
	Enabled  []Enabled  `yaml:"enabled"`
}

type Disabled struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

type Enabled struct {
	Name     string `yaml:"name"`
	Path     string `yaml:"path"`
	Priority int64  `yaml:"priority"`
	Weight   int64  `yaml:"weight"`
}

type Logger struct {
	CallerSkip   int64  `yaml:"callerSkip"`
	FileCompress bool   `yaml:"fileCompress"`
	FileName     string `yaml:"fileName"`
	LogLevel     string `yaml:"logLevel"`
	MaxAge       int64  `yaml:"maxAge"`
	MaxBackups   int64  `yaml:"maxBackups"`
	MaxSize      int64  `yaml:"maxSize"`
}

var (
	Build   string
	Version string
)

func New() *Config {
	return &Config{}
}
