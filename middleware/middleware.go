package middleware

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"

	"github.com/VladislavTyurin/image_previewer/cache"
	"github.com/VladislavTyurin/image_previewer/config"
)

var (
	ErrInvalidURL   = errors.New("invalid url")
	ErrNotJpegImage = errors.New("iile isn't JPEG image")
)

type ContextKey string

func GetFromContext(ctx context.Context, key string) interface{} {
	return ctx.Value(ContextKey(key))
}

type Middleware struct {
	client      *http.Client
	fillPattren *regexp.Regexp
	cache       cache.Cache
	conf        *config.Config
	mtx         sync.RWMutex
}

func NewMiddleware(conf *config.Config, cache cache.Cache) *Middleware {
	return &Middleware{
		client:      &http.Client{},
		fillPattren: regexp.MustCompile(`/fill/(\d+)/(\d+)/(.+)`),
		cache:       cache,
		conf:        conf,
	}
}

func (m *Middleware) ValidateURL(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case m.fillPattren.MatchString(r.URL.Path):
			groups := m.fillPattren.FindAllStringSubmatch(r.URL.Path, -1)
			width, err := strconv.Atoi(groups[0][1])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
			height, err := strconv.Atoi(groups[0][2])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
			ctx := context.WithValue(r.Context(), ContextKey("width"), width)
			ctx = context.WithValue(ctx, ContextKey("height"), height)
			ctx = context.WithValue(ctx, ContextKey("source"), groups[0][3])
			next.ServeHTTP(w, r.WithContext(ctx))
		default:
			http.NotFound(w, r)
		}
	})
}

func (m *Middleware) GetFromSource(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if source, ok := GetFromContext(r.Context(), "source").(string); !ok {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("source path cannot be converted to string\n"))
		} else {
			if value, ok := m.cache.Get(cache.Key(source)); ok {
				m.handleImageFromDisk(next, w, r, value.(string))
			} else {
				m.handleFromRemote(next, w, r, source)
			}
		}
	})
}

func (m *Middleware) handleImageFromDisk(
	next http.Handler,
	w http.ResponseWriter,
	r *http.Request,
	image string,
) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	f, err := os.Open(image)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer f.Close()

	ctx := context.WithValue(r.Context(), ContextKey("image"), f)
	next.ServeHTTP(w, r.WithContext(ctx))
}

func (m *Middleware) handleFromRemote(
	next http.Handler,
	w http.ResponseWriter,
	r *http.Request,
	source string,
) {
	proxyPequest, err := http.NewRequestWithContext(r.Context(), http.MethodGet, "http://"+source, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	proxyPequest.Header = r.Header

	resp, err := m.client.Do(proxyPequest)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(resp.StatusCode)
		return
	}

	if resp.Header.Get("Content-Type") != "image/jpeg" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(ErrNotJpegImage.Error()))
		return
	}

	m.mtx.Lock()
	if value, ok := m.cache.Get(cache.Key(source)); ok {
		m.mtx.Unlock()
		m.handleImageFromDisk(next, w, r, value.(string))
		return
	}
	f, err := os.CreateTemp(m.conf.CacheDir, "image_")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	f.Seek(0, io.SeekStart)

	m.cache.Set(cache.Key(source), f.Name())
	m.mtx.Unlock()

	ctx := context.WithValue(r.Context(), ContextKey("image"), f)
	next.ServeHTTP(w, r.WithContext(ctx))
}
