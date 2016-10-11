package server

import (
	"net/http"
	"os"

	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/system"
)

// HTTP Handler - GET /
func HandleIndex(w http.ResponseWriter, r *http.Request) {
	hostname, err := os.Hostname()
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	usage, err := system.CpuUsage()
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	var resp shared.IndexDef
	resp.Hostname = hostname
	resp.CpuUsage = usage

	SuccessResponse(w, r, resp)
}
