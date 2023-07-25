package server

import (
	"log"
	"net/http"

	"github.com/VladislavTyurin/image_previewer/cache"
	"github.com/VladislavTyurin/image_previewer/config"
	"github.com/VladislavTyurin/image_previewer/middleware"
	"github.com/VladislavTyurin/image_previewer/previewer"
)

type Server interface {
	Run()
}

type serverImpl struct {
	conf *config.Config
	pr   previewer.Previewer
	m    middleware.Middleware
}

func NewServer(conf *config.Config, cache cache.Cache) Server {
	return &serverImpl{
		conf: conf,
		pr:   previewer.NewPreviewer(),
		m:    *middleware.NewMiddleware(conf, cache),
	}
}

func (s *serverImpl) Run() {
	fillHandler := s.m.ValidateURL(s.m.GetFromSource(http.HandlerFunc(s.pr.Fill)))
	if err := http.ListenAndServe(s.conf.Address(), fillHandler); err != nil { //nolint:gosec
		log.Fatal(err)
	}
}
