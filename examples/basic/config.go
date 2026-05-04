package main

type Config struct {
	App struct {
		Name        string `yaml:"name"`
		Environment string `yaml:"environment"`
		Port        int    `yaml:"port"`
	} `yaml:"app"`
	Database struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"database"`
}
