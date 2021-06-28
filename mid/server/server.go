// Itero - Online iterative vote application
// Copyright (C) 2020 Joseph Boudou
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Package server provides classes and functions for the HTTP server side of the middleware.
//
// In particular, the package handles client sessions by producing credentials for logged user and
// by verifying these credentials for each request.
//
// It is a wrapper around net/http.
package server

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/JBoudou/Itero/mid/root"
	"github.com/JBoudou/Itero/pkg/config"
	"github.com/JBoudou/Itero/pkg/slog"

	gs "github.com/gorilla/sessions"
	"github.com/justinas/alice"
)

const (
	// Name of the cookie for sessions.
	SessionName = "s"

	// Name of the cookie for unlogged users.
	SessionUnlogged = "u"

	// Max age of the session cookies in seconds. Also used to compute deadlines.
	sessionMaxAge         = 30 * 60
	sessionUnloggedMaxAge = 30 * 24 * 3600

	// Additional delay accorded after deadline is reached.
	sessionGraceTime = 20

	sessionKeySessionId = "sid"
	sessionKeyUserName  = "usr"
	sessionKeyUserId    = "uid"
	sessionKeyDeadline  = "dl"
	sessionKeyHash      = "hash"

	defaultPort   = ":443"
	sessionHeader = "X-CSRF"

	wwwroot = "app/dist/app"
)

var (
	cfg           myConfig
	sessionStore  *gs.CookieStore
	unloggedStore *gs.CookieStore
)

// Ok indicates whether the package is usable. May be false if there is no configuration for the
// package.
var Ok bool

// SessionOptions reflects the configured options for sessions.
// Modifying it has no effect on the sessions generated by the package.
var SessionOptions gs.Options

type myConfig struct {
	Address     string
	CertFile    string
	KeyFile     string
	SessionKeys [][]byte
}

func init() {
	// Configuration
	cfg.Address = defaultPort
	if err := config.Value("server", &cfg); err != nil {
		log.Print(err)
		log.Println("WARNING: Package server not usable because there is no configuration for it.")
		Ok = false
		return
	}
	Ok = true
	cfg.Address = strings.TrimSuffix(cfg.Address, defaultPort)

	// Session
	sessionStore = gs.NewCookieStore(cfg.SessionKeys...)
	sessionStore.MaxAge(sessionMaxAge)
	sessionStore.Options.Domain = HostOnly(cfg.Address)
	sessionStore.Options.SameSite = http.SameSiteLaxMode
	sessionStore.Options.Secure = true

	SessionOptions = *sessionStore.Options

	// Unlogged
	unloggedStore = gs.NewCookieStore(cfg.SessionKeys...)
	*unloggedStore.Options = SessionOptions
	unloggedStore.MaxAge(sessionUnloggedMaxAge)
}

// HostOnly returns the host part of an address, without the port.
func HostOnly(address string) string {
	if !strings.Contains(address, ":") {
		return address
	}
	return strings.Split(address, ":")[0]
}

// User represents a logged user.
type User struct {
	Id   uint32
	Name string
	Hash uint32

	// If Logged is true then Name is meaningfull else Hash is meaningfull.
	Logged bool
}

var interceptorChain = alice.New(addLogger)

type loggerInterceptor struct {
	next   http.Handler
	logger slog.Stacked
}

func (self loggerInterceptor) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	start := time.Now()
	logger := self.logger.With(req.RemoteAddr, req.URL.Path)
	ctx := slog.CtxSaveLogger(req.Context(), logger)
	self.next.ServeHTTP(wr, req.WithContext(ctx))
	logger.Log("in", time.Now().Sub(start).String())
}

func addLogger(next http.Handler) http.Handler {
	var printer slog.Printer
	if err := root.IoC.Inject(&printer); err != nil {
		panic(err)
	}
	return loggerInterceptor{
		next: next,
		logger: &slog.SimpleLogger{
			Printer: printer,
			Stack:   []interface{}{"H"},
		},
	}
}

type oneFile struct {
	path string
}

func (self oneFile) Open(name string) (http.File, error) {
	return os.Open(self.path)
}

// Start the server.
// Parameters are taken from the configuration.
func Start() error {
	http.Handle("/r/", interceptorChain.
		Then(http.FileServer(oneFile{wwwroot + "/index.html"})))
	http.Handle("/", interceptorChain.
		Then(http.FileServer(http.Dir(wwwroot))))
	http.Handle("/s/", interceptorChain.
		Then(http.StripPrefix("/s/", http.FileServer(http.Dir("static")))))

	addr := cfg.Address
	if !strings.Contains(addr, ":") {
		addr = addr + defaultPort
	}

	return http.ListenAndServeTLS(addr, cfg.CertFile, cfg.KeyFile, nil)
}

// BaseURL returns the URL of the application.
func BaseURL() string {
	return "https://" + cfg.Address + "/"
}

// SessionKeys retrieves the session keys for test purpose.
//
// This is a low level function, made available for tests.
func SessionKeys() [][]byte {
	return cfg.SessionKeys
}
