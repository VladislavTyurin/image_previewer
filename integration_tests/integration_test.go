// go:build integration

package integrationtests

import (
	"context"
	"fmt"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/VladislavTyurin/image_previewer/cache"
	"github.com/VladislavTyurin/image_previewer/config"
	"github.com/VladislavTyurin/image_previewer/middleware"
	"github.com/VladislavTyurin/image_previewer/previewer"
	"github.com/stretchr/testify/require"
)

func fillHandler() (http.Handler, cache.Cache) {
	c := cache.NewCache(2)
	m := middleware.NewMiddleware(&config.Config{
		CacheLimit: 2,
		CacheDir:   "tmp",
		Host:       "127.0.0.1",
		Port:       1234,
	},
		c)
	pr := previewer.NewPreviewer()
	return m.ValidateURL(m.GetFromSource(http.HandlerFunc(pr.Fill))), c
}

func TestGetImageSuccess(t *testing.T) {
	require.NoError(t, os.MkdirAll("tmp", 0o755))
	defer os.RemoveAll("tmp")

	w := httptest.NewRecorder()
	uri := "http://127.0.0.1:1234/fill/200/300/127.0.0.1:8181/gopher1.jpg"
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, uri, nil)
	require.NoError(t, err)

	handler, c := fillHandler()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	_, ok := c.Get(cache.Key("127.0.0.1:8181/gopher1.jpg"))
	require.True(t, ok)

	img, err := jpeg.Decode(w.Body)
	require.NoError(t, err)
	require.Equal(t, 200, img.Bounds().Dx())
	require.Equal(t, 300, img.Bounds().Dy())
}

func TestInvalidPath(t *testing.T) {
	require.NoError(t, os.MkdirAll("tmp", 0o755))
	defer os.RemoveAll("tmp")

	w := httptest.NewRecorder()
	uri := "http://127.0.0.1:1234/filll/200/300/127.0.0.1/gopher1.jpg"
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, uri, nil)
	require.NoError(t, err)

	handler, c := fillHandler()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
	_, ok := c.Get(cache.Key("127.0.0.1/gopher1.jpg"))
	require.False(t, ok)
}

func TestInvalidServiceName(t *testing.T) {
	require.NoError(t, os.MkdirAll("tmp", 0o755))
	defer os.RemoveAll("tmp")

	w := httptest.NewRecorder()

	uri := "http://127.0.0.1:1234/fill/200/300/127.0.0.:8181/gopher1.jpg"
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, uri, nil)
	require.NoError(t, err)

	handler, c := fillHandler()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadGateway, w.Code)
	_, ok := c.Get(cache.Key("127.0.0.:8181/gopher1.jpg"))
	require.False(t, ok)
}

func TestImageNotFound(t *testing.T) {
	require.NoError(t, os.MkdirAll("tmp", 0o755))
	defer os.RemoveAll("tmp")

	w := httptest.NewRecorder()

	uri := "http://127.0.0.1:1234/fill/200/300/127.0.0.1:8181/gopher_1.jpg"
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, uri, nil)
	require.NoError(t, err)

	handler, c := fillHandler()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
	_, ok := c.Get(cache.Key("127.0.0.1:8181/gopher_1.jpg"))
	require.False(t, ok)
}

func TestNotJpegImage(t *testing.T) {
	require.NoError(t, os.MkdirAll("tmp", 0o755))
	defer os.RemoveAll("tmp")

	w := httptest.NewRecorder()

	uri := "http://127.0.0.1:1234/fill/200/300/127.0.0.1:8181/gopher.png"
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, uri, nil)
	require.NoError(t, err)

	handler, c := fillHandler()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
	_, ok := c.Get(cache.Key("127.0.0.1:8181/gopher.png"))
	require.False(t, ok)
}

func TestManyRequests(t *testing.T) {
	require.NoError(t, os.MkdirAll("tmp", 0o755))
	defer os.RemoveAll("tmp")

	handler, c := fillHandler()
	cache.NewCustomDeleter(func(value interface{}) {
		os.RemoveAll(value.(string))
	})
	defer cache.ResetDeleter()

	for i := 1; i < 5; i++ {
		w := httptest.NewRecorder()
		uri := fmt.Sprintf("http://127.0.0.1:1234/fill/200/300/127.0.0.1:8181/gopher%d.jpg", i)
		req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, uri, nil)
		require.NoError(t, err)
		handler.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		_, ok := c.Get(cache.Key(fmt.Sprintf("127.0.0.1:8181/gopher%d.jpg", i)))
		require.True(t, ok)
		if i > 2 {
			for j := i - 2; j > 0; j-- {
				_, ok = c.Get(cache.Key(fmt.Sprintf("127.0.0.1:8181/gopher%d.jpg", j)))
				require.False(t, ok)
			}
		}
		entries, err := os.ReadDir("tmp")
		require.NoError(t, err)
		require.LessOrEqual(t, len(entries), 2)
	}
}

func TestManyConcurrentRequests(t *testing.T) {
	require.NoError(t, os.MkdirAll("tmp", 0o755))
	defer os.RemoveAll("tmp")

	handler, _ := fillHandler()
	cache.NewCustomDeleter(func(value interface{}) {
		os.RemoveAll(value.(string))
	})
	defer cache.ResetDeleter()
	wg := sync.WaitGroup{}
	wg.Add(4)
	for i := 1; i < 5; i++ {
		go func(i int) {
			defer wg.Done()
			w := httptest.NewRecorder()
			uri := fmt.Sprintf("http://127.0.0.1:1234/fill/200/300/127.0.0.1:8181/gopher%d.jpg", i)
			req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, uri, nil)
			require.NoError(t, err)
			handler.ServeHTTP(w, req)
			require.Equal(t, http.StatusOK, w.Code)
			entries, err := os.ReadDir("tmp")
			require.NoError(t, err)
			require.LessOrEqual(t, len(entries), 2)
		}(i)
	}
	wg.Wait()
}
