package server

import (
	"log"
	"net/http"

	"github.com/VladislavTyurin/image_previewer/cache"
	"github.com/VladislavTyurin/image_previewer/config"
	"github.com/VladislavTyurin/image_previewer/previewer"
)

type Server interface {
	Run()
}

type serverImpl struct {
	conf *config.Config
	pr   previewer.Previewer
}

func NewServer(conf *config.Config, cache cache.Cache) Server {
	return &serverImpl{
		conf: conf,
		pr:   previewer.NewPreviewer(conf, cache),
	}
}

func (s *serverImpl) Run() {
	if err := http.ListenAndServe( //nolint:gosec
		s.conf.Address(),
		http.HandlerFunc(s.pr.Fill)); err != nil {
		log.Fatal(err)
	}
}
