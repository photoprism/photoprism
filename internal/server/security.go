package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/pkg/clean"
)

const (
	stsHeader           = "Strict-Transport-Security"
	stsSubdomainString  = "; includeSubdomains"
	frameOptionsHeader  = "X-Frame-Options"
	frameOptionsValue   = "DENY"
	contentTypeHeader   = "X-Content-Type-Options"
	contentTypeValue    = "nosniff"
	xssProtectionHeader = "X-XSS-Protection"
	xssProtectionValue  = "1; mode=block"
	cspHeader           = "Content-Security-Policy"
)

func defaultBadHostHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Bad Host", http.StatusInternalServerError)
}

// SecurityOptions is a struct for specifying configuration options for the HTTP security middleware.
type SecurityOptions struct {
	// AllowedHosts is a list of fully qualified domain names that are allowed. Default is empty list, which allows any and all host names.
	AllowedHosts []string
	// If SSLRedirect is set to true, then only allow https requests. Default is false.
	SSLRedirect bool
	// If SSLTemporaryRedirect is true, a 302 will be used while redirecting. Default is false (301).
	SSLTemporaryRedirect bool
	// SSLHost is the host name that is used to redirect http requests to https. Default is "", which indicates to use the same host.
	SSLHost string
	// SSLProxyHeaders is set of header keys with associated values that would indicate a valid https request. Useful when using Nginx: `map[string]string{"X-Forwarded-Proto": "https"}`. Default is blank map.
	SSLProxyHeaders map[string]string
	// STSSeconds is the max-age of the Strict-Transport-Security header. Default is 0, which would NOT include the header.
	STSSeconds int64
	// If STSIncludeSubdomains is set to true, the `includeSubdomains` will be appended to the Strict-Transport-Security header. Default is false.
	STSIncludeSubdomains bool
	// If FrameDeny is set to true, adds the X-Frame-Options header with the value of `DENY`. Default is false.
	FrameDeny bool
	// CustomFrameOptionsValue allows the X-Frame-Options header value to be set with a custom value. This overrides the FrameDeny option.
	CustomFrameOptionsValue string
	// If ContentTypeNosniff is true, adds the X-Content-Type-Options header with the value `nosniff`. Default is false.
	ContentTypeNosniff bool
	// If BrowserXssFilter is true, adds the X-XSS-Protection header with the value `1; mode=block`. Default is false.
	BrowserXssFilter bool
	// ContentSecurityPolicy allows the Content-Security-Policy header value to be set with a custom value. Default is "".
	ContentSecurityPolicy string
	// When developing, the AllowedHosts, SSL, and STS options can cause some unwanted effects. Usually testing happens on http, not https, and on localhost, not your production domain... so set this to true for dev environment.
	// If you would like your development environment to mimic production with complete Host blocking, SSL redirects, and STS headers, leave this as false. Default if false.
	IsDevelopment bool

	// Handlers for when an error occurs (ie bad host).
	BadHostHandler http.Handler
}

// security is an HTTP middleware that enables basic security features. A single SecurityOptions struct can be
// provided to configure which features should be enabled, and the ability to override a few of the default values.
// Code is based on https://github.com/gin-gonic/contrib/tree/master/secure released under the MIT license:
// https://github.com/gin-gonic/contrib/blob/master/LICENSE
type security struct {
	// Customize with an SecurityOptions struct.
	opt SecurityOptions
}

// process handles an HTTP request.
func (s *security) process(w http.ResponseWriter, r *http.Request) error {
	// Allowed hosts check.
	if len(s.opt.AllowedHosts) > 0 && !s.opt.IsDevelopment {
		isGoodHost := false
		for _, allowedHost := range s.opt.AllowedHosts {
			if strings.EqualFold(allowedHost, r.Host) {
				isGoodHost = true
				break
			}
		}

		if !isGoodHost {
			s.opt.BadHostHandler.ServeHTTP(w, r)
			return fmt.Errorf("server: bad host %s", clean.Log(r.Host))
		}
	}

	// SSL check.
	if s.opt.SSLRedirect && s.opt.IsDevelopment == false {
		isSSL := false
		if strings.EqualFold(r.URL.Scheme, "https") || r.TLS != nil {
			isSSL = true
		} else {
			for k, v := range s.opt.SSLProxyHeaders {
				if r.Header.Get(k) == v {
					isSSL = true
					break
				}
			}
		}

		if isSSL == false {
			url := r.URL
			url.Scheme = "https"
			url.Host = r.Host

			if len(s.opt.SSLHost) > 0 {
				url.Host = s.opt.SSLHost
			}

			status := http.StatusMovedPermanently
			if s.opt.SSLTemporaryRedirect {
				status = http.StatusTemporaryRedirect
			}

			http.Redirect(w, r, url.String(), status)
			return fmt.Errorf("server: https redirect")
		}
	}

	// Strict Transport Security header.
	if s.opt.STSSeconds != 0 && !s.opt.IsDevelopment {
		stsSub := ""
		if s.opt.STSIncludeSubdomains {
			stsSub = stsSubdomainString
		}

		w.Header().Add(stsHeader, fmt.Sprintf("max-age=%d%s", s.opt.STSSeconds, stsSub))
	}

	// Frame Options header.
	if len(s.opt.CustomFrameOptionsValue) > 0 {
		w.Header().Add(frameOptionsHeader, s.opt.CustomFrameOptionsValue)
	} else if s.opt.FrameDeny {
		w.Header().Add(frameOptionsHeader, frameOptionsValue)
	}

	// Content Type Options header.
	if s.opt.ContentTypeNosniff {
		w.Header().Add(contentTypeHeader, contentTypeValue)
	}

	// XSS Protection header.
	if s.opt.BrowserXssFilter {
		w.Header().Add(xssProtectionHeader, xssProtectionValue)
	}

	// Content Security Policy header.
	if len(s.opt.ContentSecurityPolicy) > 0 {
		w.Header().Add(cspHeader, s.opt.ContentSecurityPolicy)
	}

	return nil

}

// Security registers the HTTP security middleware.
func Security(options SecurityOptions) gin.HandlerFunc {
	if options.BadHostHandler == nil {
		options.BadHostHandler = http.HandlerFunc(defaultBadHostHandler)
	}

	s := &security{
		opt: options,
	}

	return func(c *gin.Context) {
		err := s.process(c.Writer, c.Request)
		if err != nil {
			if c.Writer.Written() {
				c.AbortWithStatus(c.Writer.Status())
			} else {
				_ = c.AbortWithError(http.StatusInternalServerError, err)
			}
		}
	}

}
