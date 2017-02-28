package main

import (
	log "github.com/Sirupsen/logrus"
	//"github.com/jung-kurt/gofpdf"
	"github.com/muesli/smolder"
	"gopkg.in/gcfg.v1"
	"net/http"
)

type Config struct {
	Network struct {
		ListenTo string
	}
	Templates struct {
		Directory string
	}
	Log struct {
		File  string
		Level string
	}
}

func main() {
	cfg := Config{}
	err := gcfg.ReadFileInto(&cfg, "/etc/quit/quit.conf")
	if err != nil {
		log.WithField("Error", err).Fatalln("Could not read config file")
	}

	lvl, err := log.ParseLevel(cfg.Log.Level)
	if err != nil {
		log.WithField("Error", err).Error("Could not parse config level, defaulting to info")
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(lvl)
	}

	log.Info("Config read successfull")

	wsContainer := smolder.NewSmolderContainer(
		smolder.APIConfig{
			BaseURL:    cfg.Network.ListenTo,
			PathPrefix: "",
		}, nil, nil)

	server := http.Server{Addr: cfg.Network.ListenTo, Handler: wsContainer}
	log.Fatal(server.ListenAndServe())
}
