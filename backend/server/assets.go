package server

import (
	"net/http"

	cloudhub "github.com/snetsystems/cloudhub/backend"
	"github.com/snetsystems/cloudhub/backend/dist"
)

const (
	// Dir is prefix of the assets in the bindata
	Dir = "../../frontend/build"
	// Default is the default item to load if 404
	Default = "../../frontend/build/index.html"
	// DebugDir is the prefix of the assets in development mode
	DebugDir = "../../../frontend/build"
	// DebugDefault is the default item to load if 404
	DebugDefault = "../../../frontend/build/index.html"
	// DefaultContentType is the content-type to return for the Default file
	DefaultContentType = "text/html; charset=utf-8"
)

// AssetsOpts configures the asset middleware
type AssetsOpts struct {
	// Develop when true serves assets from frontend/build directory directly; false will use internal bindata.
	Develop bool
	// Logger will log the asset served
	Logger cloudhub.Logger
}

// Assets creates a middleware that will serve a single page app.
func Assets(opts AssetsOpts) http.Handler {
	var assets cloudhub.Assets
	if opts.Develop {
		assets = &dist.DebugAssets{
			Dir:     DebugDir,
			Default: DebugDefault,
		}
	} else {
		assets = &dist.BindataAssets{
			Prefix:             Dir,
			Default:            Default,
			DefaultContentType: DefaultContentType,
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if opts.Logger != nil {
			opts.Logger.
				WithField("component", "server").
				WithField("remote_addr", r.RemoteAddr).
				WithField("method", r.Method).
				WithField("url", r.URL).
				Info("Serving assets")
		}
		assets.Handler().ServeHTTP(w, r)
	})
}
