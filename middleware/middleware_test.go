package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/VladislavTyurin/image_previewer/cache"
	"github.com/VladislavTyurin/image_previewer/config"
	"github.com/stretchr/testify/require"
)

type validateURLResponse struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Source string `json:"source"`
}

func validateURLNext(w http.ResponseWriter, r *http.Request) {
	width := GetFromContext(r.Context(), "width").(int)
	height := GetFromContext(r.Context(), "height").(int)
	source := GetFromContext(r.Context(), "source").(string)

	res := validateURLResponse{
		Width:  width,
		Height: height,
		Source: source,
	}

	b, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func TestValidateUrlSuccess(t *testing.T) {
	m := NewMiddleware(&config.Config{}, cache.NewCache(1))
	w := httptest.NewRecorder()
	r, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "/fill/200/300/source", nil)
	require.NoError(t, err)

	validateHandler := m.ValidateURL(http.HandlerFunc(validateURLNext))

	validateHandler.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)

	res := validateURLResponse{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	require.Equal(t, 200, res.Width)
	require.Equal(t, 300, res.Height)
	require.Equal(t, "source", res.Source)
}

func TestValidateUrlFail(t *testing.T) {
	urls := []struct {
		url          string
		expectedCode int
	}{
		{
			url:          "/fil/200/300/source",
			expectedCode: http.StatusNotFound,
		},
		{
			url:          "/fill/200/300",
			expectedCode: http.StatusNotFound,
		},
		{
			url:          "/fill/width/300/source",
			expectedCode: http.StatusNotFound,
		},
		{
			url:          "/fill/200/height/source",
			expectedCode: http.StatusNotFound,
		},
		{
			url:          "/fill/222222222222222222222222222/300/source",
			expectedCode: http.StatusBadRequest,
		},
	}

	m := NewMiddleware(&config.Config{}, cache.NewCache(1))
	for _, u := range urls {
		t.Run(u.url, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, u.url, nil)
			require.NoError(t, err)

			validateHandler := m.ValidateURL(http.HandlerFunc(validateURLNext))

			validateHandler.ServeHTTP(w, r)
			require.Equal(t, u.expectedCode, w.Code)
		})
	}
}

func TestGetFromSourceSuccess(t *testing.T) {
	require.NoError(t, os.MkdirAll("tmp", 0o755))
	defer os.RemoveAll("tmp")
	m := NewMiddleware(&config.Config{CacheDir: "tmp"}, cache.NewCache(1))

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		image := GetFromContext(r.Context(), "image").(*os.File)
		w.Write([]byte(image.Name()))
	})

	var imageName string
	// Запустим 2 раза, первый раз картинку получим из источника, второй раз из кэша (то же имя)
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()

		r, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "any", nil)
		require.NoError(t, err)
		r = r.WithContext(context.WithValue(r.Context(), ContextKey("source"),
			"raw.githubusercontent.com/OtusGolang/final_project/master/examples/image-previewer/_gopher_original_1024x504.jpg"))
		sourceHandler := m.GetFromSource(next)

		sourceHandler.ServeHTTP(w, r)
		require.Equal(t, http.StatusOK, w.Code)

		if i == 0 {
			entries, err := os.ReadDir("tmp")
			require.NoError(t, err)
			require.Equal(t, 1, len(entries))
			imageName = path.Join("tmp", entries[0].Name())
		}
		if i == 1 {
			_, err = os.Stat(imageName)
			require.NoError(t, err)
		}

		require.Equal(t, imageName, w.Body.String())
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
	m := NewMiddleware(&config.Config{CacheDir: "tmp"}, cache.NewCache(1))

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		image := GetFromContext(r.Context(), "image").(*os.File)
		w.Write([]byte(image.Name()))
	})

	for _, s := range samples {
		t.Run(s.source, func(t *testing.T) {
			w := httptest.NewRecorder()

			r, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "any", nil)
			require.NoError(t, err)
			r = r.WithContext(context.WithValue(r.Context(), ContextKey("source"), s.source))
			sourceHandler := m.GetFromSource(next)

			sourceHandler.ServeHTTP(w, r)
			require.Equal(t, s.expectedCode, w.Code)
		})
	}
}
