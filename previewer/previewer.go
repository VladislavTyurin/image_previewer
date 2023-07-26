package previewer

import (
	"errors"
	"image/jpeg"
	"net/http"
	"os"

	"github.com/VladislavTyurin/image_previewer/middleware"
	"github.com/disintegration/imaging"
)

var (
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

type previewerImpl struct{}

func NewPreviewer() Previewer {
	return &previewerImpl{}
}

func (p *previewerImpl) Fill(w http.ResponseWriter, r *http.Request) {
	width := middleware.GetFromContext(r.Context(), "width").(int)
	height := middleware.GetFromContext(r.Context(), "height").(int)
	image := middleware.GetFromContext(r.Context(), "image").(*os.File)

	if width < 128 || height < 128 {
		errResponse(w, http.StatusBadRequest, ErrSizeTooSmall.Error())
		return
	}

	img, err := jpeg.Decode(image)
	if err != nil {
		errResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	preview := imaging.Resize(img, width, height, imaging.Lanczos)
	if err = jpeg.Encode(w, preview, &jpeg.Options{
		Quality: 100,
	}); err != nil {
		errResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "image/jpeg")
}
