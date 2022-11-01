package model

type Bucket struct {
	Name   string `json:name`
	Region string `json:region`
}

type Config struct {
	Version string `mapstructure:"VERSION""`
}
