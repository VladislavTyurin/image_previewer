package previewer

import (
	"context"
	"fmt"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"regexp"
	"testing"

	"github.com/VladislavTyurin/image_previewer/cache"
	"github.com/VladislavTyurin/image_previewer/config"
	"github.com/stretchr/testify/require"
)

func TestValidateUrlSuccess(t *testing.T) {
	p := previewerImpl{
		fillPattren: regexp.MustCompile(`/fill/(\d+)/(\d+)/(.+)`),
	}
	r, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "/fill/200/300/source", nil)
	require.NoError(t, err)

	params, err := p.validateURL(r)
	require.NoError(t, err)

	require.Equal(t, 200, params.width)
	require.Equal(t, 300, params.height)
	require.Equal(t, "source", params.source)
}

func TestValidateUrlFail(t *testing.T) {
	urls := []string{
		"/fil/200/300/source",
		"/fill/200/300",
		"/fill/width/300/source",
		"/fill/200/height/source",
		"/fill/222222222222222222222222222/300/source",
	}

	p := previewerImpl{
		fillPattren: regexp.MustCompile(`/fill/(\d+)/(\d+)/(.+)`),
	}
	for _, u := range urls {
		t.Run(u, func(t *testing.T) {
			r, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, u, nil)
			require.NoError(t, err)

			_, err = p.validateURL(r)
			require.Error(t, err)
		})
	}
}

func TestGetFromSourceSuccess(t *testing.T) {
	require.NoError(t, os.MkdirAll("tmp", 0o755))
	defer os.RemoveAll("tmp")
	p := previewerImpl{
		conf:        &config.Config{CacheDir: "tmp"},
		cache:       cache.NewCache(1),
		client:      &http.Client{},
		fillPattren: regexp.MustCompile(`/fill/(\d+)/(\d+)/(.+)`),
	}

	var imageName string
	source := "raw.githubusercontent.com/OtusGolang/final_project/master/examples/image-previewer/_gopher_original_1024x504.jpg" //nolint:lll
	// Запустим 2 раза, первый раз картинку получим из источника, второй раз из кэша (то же имя)
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()

		r, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "any", nil)
		require.NoError(t, err)

		_, err = p.getFromSource(w, r, source)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, w.Code)

		if i == 0 {
			entries, err := os.ReadDir("tmp")
			require.NoError(t, err)
			require.Equal(t, 1, len(entries))
			imageName = path.Join("tmp", entries[0].Name())
		}
		if i == 1 {
			value, ok := p.cache.Get(cache.Key(source))
			require.True(t, ok)
			require.Equal(t, imageName, value.(string))
			_, err = os.Stat(imageName)
			require.NoError(t, err)
		}
	}
}

func TestGetFromSourceFail(t *testing.T) {
	samples := []struct {
		source       string
		expectedCode int
	}{
		{
			source:       "raw.githubusercontent.com/OtusGolang/final_project/blob/master/image-previewer/gopher_1024x505.jpg",
			expectedCode: http.StatusNotFound,
		},
		{
			source:       "raw.githubusercontent.com",
			expectedCode: http.StatusBadRequest,
		},
		{
			source:       "github.com/OtusGolang/final_project/blob/master/examples/banners-rotation/conceptual_model.png",
			expectedCode: http.StatusBadRequest,
		},
		{
			source:       "1",
			expectedCode: http.StatusBadGateway,
		},
	}

	require.NoError(t, os.MkdirAll("tmp", 0o755))
	defer os.RemoveAll("tmp")
	p := previewerImpl{
		conf:        &config.Config{CacheDir: "tmp"},
		cache:       cache.NewCache(1),
		client:      &http.Client{},
		fillPattren: regexp.MustCompile(`/fill/(\d+)/(\d+)/(.+)`),
	}

	for _, s := range samples {
		t.Run(s.source, func(t *testing.T) {
			w := httptest.NewRecorder()

			r, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "any", nil)
			require.NoError(t, err)

			_, err = p.getFromSource(w, r, s.source)
			require.Error(t, err)
			require.Equal(t, s.expectedCode, w.Code)
		})
	}
}

func TestFillSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	url := "/fill/200/300/"
	url += "raw.githubusercontent.com/OtusGolang/final_project/master/examples/image-previewer/_gopher_original_1024x504.jpg" //nolint:lll
	r, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, url, nil)
	require.NoError(t, err)

	require.NoError(t, os.MkdirAll("tmp", 0o755))
	defer os.RemoveAll("tmp")

	pr := NewPreviewer(&config.Config{CacheDir: "tmp"}, cache.NewCache(1))
	fillHandler := http.HandlerFunc(pr.Fill)

	fillHandler.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header(), "Content-Type")
	require.Equal(t, w.Header().Get("Content-Type"), "image/jpeg")

	result, err := jpeg.Decode(w.Body)
	require.NoError(t, err)

	require.Equal(t, 200, result.Bounds().Dx())
	require.Equal(t, 300, result.Bounds().Dy())
}

func TestFillFail(t *testing.T) {
	samples := []struct {
		width  int
		heigt  int
		source string
	}{
		{
			width:  0,
			heigt:  128,
			source: "raw.githubusercontent.com/OtusGolang/final_project/master/examples/image-previewer/_gopher_original_1024x504.jpg", //nolint:lll
		},
		{
			width:  128,
			heigt:  0,
			source: "raw.githubusercontent.com/OtusGolang/final_project/master/examples/image-previewer/_gopher_original_1024x504.jpg", //nolint:lll
		},
		{
			width:  128,
			heigt:  128,
			source: "https://raw.githubusercontent.com/OtusGolang/final_project/master/examples/banners-rotation/conceptual_model.png", //nolint:lll
		},
	}

	require.NoError(t, os.MkdirAll("tmp", 0o755))
	defer os.RemoveAll("tmp")

	pr := NewPreviewer(&config.Config{CacheDir: "tmp"}, cache.NewCache(1))
	fillHandler := http.HandlerFunc(pr.Fill)

	for _, s := range samples {
		t.Run(fmt.Sprintf("%d_%d", s.width, s.heigt), func(t *testing.T) {
			w := httptest.NewRecorder()
			url := fmt.Sprintf("/fill/%d/%d/%s", s.width, s.heigt, s.source)
			r, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, url, nil)
			require.NoError(t, err)

			fillHandler.ServeHTTP(w, r)
			require.NotEqual(t, http.StatusOK, w.Code)
		})
	}
}
