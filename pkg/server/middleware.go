package server

import (
	"compress/gzip"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"net/http"
	"strings"
)

const (
	hdrContentEncoding = "Content-Encoding"
	hdrAcceptEncoding  = "Accept-Encoding"

	gzipEncoding = "gzip"
)

func DecompressGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(hdrContentEncoding) == gzipEncoding {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				HandleError(w, r, err, "decompressor: failed to create")
				return
			}
			defer func() {
				if err := gz.Close(); err != nil {
					HandleError(w, r, err, "decompressor: failed to close")
				}
			}()
			r.Body = gz
		}
		next.ServeHTTP(w, r)
	})
}

func CompressGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get(hdrAcceptEncoding), gzipEncoding) {
			gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				HandleError(w, r, err, "compressor: failed to create")
				return
			}
			defer func() {
				if err := gz.Close(); err != nil {
					HandleError(w, r, err, "compressor: failed to close")
				}
			}()

			w.Header().Set(hdrContentEncoding, gzipEncoding)
			next.ServeHTTP(ResponseCustomWriter{ResponseWriter: w, Writer: gz}, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cid := r.Header.Get(logging.CorrelationIDHeader)
		if cid == "" {
			cid = logging.NewCID()
		}
		ctx, _ := logging.SetCID(r.Context(), cid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, logger := logging.GetOrCreateLogger(ctx, logging.WithCID(ctx))

		logger = logger.With().
			Stringer("header", Header(r.Header)).
			Str("remote_addr", r.RemoteAddr).
			Logger()

		logger.Info().Msgf("%s %s", r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}
