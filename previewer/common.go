package previewer

import (
	"image"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/VladislavTyurin/image_previewer/cache"
)

type fillParams struct {
	width  int
	height int
	source string
}

func (p *previewerImpl) validateURL(r *http.Request) (*fillParams, error) {
	if !p.fillPattren.MatchString(r.URL.String()) {
		return nil, ErrInvalidURL
	}

	groups := p.fillPattren.FindAllStringSubmatch(r.URL.Path, -1)
	width, err := strconv.Atoi(groups[0][1])
	if err != nil {
		return nil, err
	}
	height, err := strconv.Atoi(groups[0][2])
	if err != nil {
		return nil, err
	}
	source := groups[0][3]
	return &fillParams{
		width:  width,
		height: height,
		source: source,
	}, nil
}

func (p *previewerImpl) getFromSource(w http.ResponseWriter, r *http.Request, source string) (image.Image, error) {
	if value, ok := p.cache.Get(cache.Key(source)); ok {
		return p.handleImageFromDisk(value.(string))
	}

	return p.handleFromRemote(w, r, source)
}

func (p *previewerImpl) handleImageFromDisk(imageFileName string) (image.Image, error) {
	f, err := os.Open(imageFileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	return img, err
}

func (p *previewerImpl) handleFromRemote(w http.ResponseWriter, r *http.Request, source string) (image.Image, error) {
	proxyPequest, err := http.NewRequestWithContext(r.Context(), http.MethodGet, "http://"+source, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}
	proxyPequest.Header = r.Header

	resp, err := p.client.Do(proxyPequest)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(resp.StatusCode)
		return nil, ErrGetFromSource
	}

	if resp.Header.Get("Content-Type") != "image/jpeg" {
		w.WriteHeader(http.StatusBadRequest)
		return nil, ErrNotJpegImage
	}

	p.mtx.Lock()
	defer p.mtx.Unlock()
	if value, ok := p.cache.Get(cache.Key(source)); ok {
		return p.handleImageFromDisk(value.(string))
	}
	f, err := os.CreateTemp(p.conf.CacheDir, "image_")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	f.Seek(0, io.SeekStart)
	p.cache.Set(cache.Key(source), f.Name())
	img, _, err := image.Decode(f)

	return img, err
}
