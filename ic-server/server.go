//GORM_DIALECT=postgres DB_URL="postgres://postgres:6Vno3r3gH9sZHSxo@localhost/ic_signal_proc_test?sslmode=disable"
package main

import (
	"github.com/empirefox/gin-oauth2"
	"github.com/empirefox/gotool/paas"
	. "github.com/empirefox/ic-server-ws-signal/account"
	"github.com/empirefox/ic-server-ws-signal/connections"
	"github.com/empirefox/ic-server-ws-signal/server"
)

// Must set PORT and DB_URL to test mode
func main() {
	//	runtime.GOMAXPROCS(runtime.NumCPU())

	//	as := NewAccountService()
	//	as.DropTables()
	//	as.CreateTables()

	h := connections.NewHub()
	go h.Run()

	conf, oauthBs := NewGoauthConf()
	conf.PathSuccess = "/rooms.html"
	conf.NewUserFunc = func() goauth.OauthUser {
		return &Oauth{}
	}

	s := server.Server{
		Hub:         h,
		OauthConfig: conf,
		OauthJson:   oauthBs,
		IsDevMode:   paas.IsDevMode,
	}
	s.Run()
}
