package main

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type internalConfig struct {
	TimeoutSeconds float64 `json:"timeout"`
	ShowqPath      string  `json:"showq_path"`
	Listen         string  `json:"listen"`
}

type Config struct {
	Timeout   time.Duration
	ShowqPath string
	Listen    string
}

func NewConfig(filename string) *Config {
	fd, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}
	defer fd.Close()

	config := &internalConfig{
		TimeoutSeconds: 10,
		ShowqPath:      "/var/spool/postfix/public/showq",
		Listen:         ":9091",
	}
	if err := json.NewDecoder(fd).Decode(config); err != nil {
		log.Fatalln(err)
	}

	return &Config{
		Timeout:   time.Duration(config.TimeoutSeconds * float64(time.Second)),
		ShowqPath: config.ShowqPath,
		Listen:    config.Listen,
	}
}
