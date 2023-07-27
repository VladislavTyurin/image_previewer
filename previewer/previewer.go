package previewer

import (
	"errors"
	"image/jpeg"
	"net/http"
	"regexp"
	"sync"

	"github.com/VladislavTyurin/image_previewer/cache"
	"github.com/VladislavTyurin/image_previewer/config"
	"github.com/disintegration/imaging"
)

var (
	ErrInvalidURL    = errors.New("invalid url")
	ErrGetFromSource = errors.New("error getting image from source")
	ErrNotJpegImage  = errors.New("image isn't JPEG image")
	ErrImageNotFound = errors.New("image not found")
	ErrSizeTooSmall  = errors.New("width or height must be greater or equal 128")
)

func errResponse(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	w.Write([]byte(message))
}

type Previewer interface {
	Fill(http.ResponseWriter, *http.Request)
}

type previewerImpl struct {
	client      *http.Client
	fillPattren *regexp.Regexp
	cache       cache.Cache
	conf        *config.Config
	mtx         sync.RWMutex
}

func NewPreviewer(conf *config.Config, cache cache.Cache) Previewer {
	return &previewerImpl{
		client:      &http.Client{},
		fillPattren: regexp.MustCompile(`/fill/(\d+)/(\d+)/(.+)`),
		cache:       cache,
		conf:        conf,
	}
}

func (p *previewerImpl) Fill(w http.ResponseWriter, r *http.Request) {
	params, err := p.validateURL(r)
	if err != nil {
		if errors.Is(err, ErrInvalidURL) {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if params.width < 128 || params.height < 128 {
		errResponse(w, http.StatusBadRequest, ErrSizeTooSmall.Error())
		return
	}

	img, err := p.getFromSource(w, r, params.source)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	preview := imaging.Resize(img, params.width, params.height, imaging.Lanczos)
	if err = jpeg.Encode(w, preview, &jpeg.Options{
		Quality: 100,
	}); err != nil {
		errResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "image/jpeg")
}
