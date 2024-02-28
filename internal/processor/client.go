package processor

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/autobrr/omegabrr/internal/buildinfo"
)

func setUserAgent(req *http.Request) {
	agent := fmt.Sprintf("omegabrr/%s (%s %s)", buildinfo.Version, runtime.GOOS, runtime.GOARCH)

	req.Header.Set("User-Agent", agent)
}
