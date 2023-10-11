package main

import (
	"net/http"

	"github.com/sqooba/go-common/healthchecks"
)

// routes define all the routes of the http multiplexer
func (wh *mutationWH) routes(mux *http.ServeMux, env envConfig) {
	mux.Handle("/mutate", wh.admitFuncHandler(wh.applyMutations))
	mux.Handle(healthchecks.HealthCheckPath, healthchecks.AlwaysOkHealthcheckFuncHandler())
}
