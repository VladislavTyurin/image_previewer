package main

import (
	"flag"
	"log"
	"os"

	"github.com/VladislavTyurin/image_previewer/cache"
	"github.com/VladislavTyurin/image_previewer/config"
	"github.com/VladislavTyurin/image_previewer/server"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "", "./imagepreviewer -c=<path_to_config>")

	flag.Parse()
	if configPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	conf, err := config.LoadConfig(configPath)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	defer os.RemoveAll(conf.CacheDir)

	cacheObj := cache.NewCache(conf.CacheLimit)
	// При удалении значения из кэша автоматически удаляем с диска
	cache.NewCustomDeleter(func(value interface{}) {
		os.RemoveAll(value.(string))
	})

	s := server.NewServer(conf, cacheObj)
	s.Run()
}
