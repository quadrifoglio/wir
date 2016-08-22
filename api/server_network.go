package api

import (
	gonet "net"

	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/net"
	"github.com/quadrifoglio/wir/shared"
)

func handleNetworkCreate(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	var netw shared.Network

	err := json.NewDecoder(r.Body).Decode(&netw)
	if err != nil {
		ErrorResponse(errors.BadRequest).Send(w, r)
		return
	}

	if len(netw.Name) == 0 {
		ErrorResponse(errors.BadRequest).Send(w, r)
		return
	}

	if len(netw.Gateway) > 0 && !net.InterfaceExists(netw.Gateway) {
		ErrorResponse(errors.BadRequest).Send(w, r)
		return
	}

	if len(netw.Addr) > 0 && gonet.ParseIP(netw.Addr) == nil {
		ErrorResponse(errors.BadRequest).Send(w, r)
		return
	}

	br := net.BridgeName(netw.Name)

	err = net.CreateBridge(br)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	if len(netw.Gateway) > 0 {
		err = net.AddBridgeIf(br, netw.Gateway)
		if err != nil {
			ErrorResponse(err).Send(w, r)
			return
		}
	}

	err = DBStoreNetwork(netw)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(netw).Send(w, r)
}

func handleNetworkList(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	netws, err := DBListNetworks()
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(netws).Send(w, r)
}

func handleNetworkGet(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	vars := mux.Vars(r)
	name := vars["name"]

	netw, err := DBGetNetwork(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(netw).Send(w, r)
}

func handleNetworkDelete(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	vars := mux.Vars(r)
	name := vars["name"]

	netw, err := DBGetNetwork(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = net.DeleteBridge(net.BridgeName(netw.Name))
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = DBDeleteNetwork(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(nil).Send(w, r)
}
