package previewer

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/VladislavTyurin/OTUS_image_previewer/middleware"
	"github.com/stretchr/testify/require"
)

func TestFillSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "any", nil)
	require.NoError(t, err)

	require.NoError(t, os.MkdirAll("tmp", 0o755))
	defer os.RemoveAll("tmp")

	f, err := os.CreateTemp("tmp", "image_")
	require.NoError(t, err)
	sourceImage := image.NewNRGBA(image.Rect(0, 0, 1024, 1024))
	require.NoError(t, jpeg.Encode(f, sourceImage, &jpeg.Options{Quality: 100}))
	defer f.Close()
	f.Seek(0, io.SeekStart)

	ctx := context.WithValue(r.Context(), middleware.ContextKey("width"), 200)
	ctx = context.WithValue(ctx, middleware.ContextKey("height"), 200)
	ctx = context.WithValue(ctx, middleware.ContextKey("image"), f)

	pr := NewPreviewer()

	fillHandler := http.HandlerFunc(pr.Fill)

	fillHandler.ServeHTTP(w, r.WithContext(ctx))
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header(), "Content-Type")
	require.Equal(t, w.Header().Get("Content-Type"), "image/jpeg")

	result, err := jpeg.Decode(w.Body)
	require.NoError(t, err)

	require.Equal(t, 200, result.Bounds().Dx())
	require.Equal(t, 200, result.Bounds().Dy())
}

func TestFillFail(t *testing.T) {
	samples := []struct {
		width  int
		heigt  int
		format string
	}{
		{
			width:  0,
			heigt:  128,
			format: "jpeg",
		},
		{
			width:  128,
			heigt:  0,
			format: "jpeg",
		},
		{
			width:  128,
			heigt:  128,
			format: "png",
		},
	}

	require.NoError(t, os.MkdirAll("tmp", 0o755))
	defer os.RemoveAll("tmp")

	pr := NewPreviewer()
	fillHandler := http.HandlerFunc(pr.Fill)

	for _, s := range samples {
		t.Run(fmt.Sprintf("%d_%d_%s", s.width, s.heigt, s.format), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "any", nil)
			require.NoError(t, err)

			f, err := os.CreateTemp("tmp", "image_")
			require.NoError(t, err)
			sourceImage := image.NewNRGBA(image.Rect(0, 0, 1024, 1024))
			defer f.Close()
			if s.format == "jpeg" {
				require.NoError(t, jpeg.Encode(f, sourceImage, &jpeg.Options{Quality: 100}))
			} else {
				require.NoError(t, png.Encode(f, sourceImage))
			}

			ctx := context.WithValue(r.Context(), middleware.ContextKey("width"), s.width)
			ctx = context.WithValue(ctx, middleware.ContextKey("height"), s.heigt)
			ctx = context.WithValue(ctx, middleware.ContextKey("image"), f)

			fillHandler.ServeHTTP(w, r.WithContext(ctx))
			require.NotEqual(t, http.StatusOK, w.Code)
		})
	}
}
