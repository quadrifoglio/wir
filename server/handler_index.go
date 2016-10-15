package server

import (
	"net/http"
	"os"

	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/system"
)

// GET /
func HandleIndex(w http.ResponseWriter, r *http.Request) {
	hostname, err := os.Hostname()
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	cpu, err := system.CpuUsage()
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	memUsed, memTotal, err := system.MemoryUsage()
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	var resp shared.IndexDef
	resp.Hostname = hostname
	resp.CpuUsage = cpu
	resp.MemoryUsage = memUsed
	resp.MemoryTotal = memTotal

	SuccessResponse(w, r, resp)
}
