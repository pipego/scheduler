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
}

type Plugin struct {
	Disabled []Disabled `yaml:"disabled"`
	Enabled  []Enabled  `yaml:"enabled"`
}

type Disabled struct {
	Name string
	Path string
}

type Enabled struct {
	Name   string
	Path   string
	Weight int64
}

var (
	Build   string
	Version string
)

func New() *Config {
	return &Config{}
}
