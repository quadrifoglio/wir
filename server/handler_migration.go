package server

// HandlerMigration - All the migration-related handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/system"
	"github.com/quadrifoglio/wir/utils"
)

func validateFetch(req shared.MachineFetchDef) (error, int) {
	if len(req.Remote.Host) == 0 {
		return fmt.Errorf("Please specify a 'Remote.Address'"), 400
	}
	if req.Remote.Port == 0 {
		return fmt.Errorf("Please specify a 'Remote.Port'"), 400
	}
	if len(req.ID) != 8 {
		return fmt.Errorf("Invalid remote machine 'ID'"), 400
	}

	return nil, 200
}

// POST /machines/fetch
func HandleMachineFetch(w http.ResponseWriter, r *http.Request) {
	var req shared.MachineFetchDef

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ErrorResponse(w, r, err, 400)
		return
	}

	if err, status := validateFetch(req); err != nil {
		ErrorResponse(w, r, err, status)
		return
	}

	err = getRemoteMachine(req.Remote, req.ID)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}
}

func getRemoteMachine(r shared.RemoteDef, id string) error {
	m, err := client.MachineGet(r, id)
	if err != nil {
		return err
	}

	img, err := client.ImageGet(r, m.Image)
	if err != nil {
		return err
	}

	err = fetchImage(r, &img)
	if err != nil {
		return err
	}

	err = fetchMachine(r, &m)
	if err != nil {
		return err
	}

	return nil
}

func fetchImage(r shared.RemoteDef, img *shared.ImageDef) error {
	if DBImageExists(img.ID) {
		return nil
	}

	// Download the distant image
	newSource := ImageFile(img.ID)

	err := os.MkdirAll(filepath.Dir(newSource), 0755)
	if err != nil {
		return err
	}

	err = system.DownloadHttp(fmt.Sprintf("http://%s:%d/images/%s/data", r.Host, r.Port, img.ID), newSource)
	if err != nil {
		return err
	}

	// Update parameters
	img.Source = newSource

	err = DBImageCreate(*img)
	if err != nil {
		return err
	}

	return nil
}

func fetchMachine(r shared.RemoteDef, m *shared.MachineDef) error {
	// Generate a new ID
	var newId string
	for {
		newId = utils.RandID()
		if !DBMachineExists(newId) {
			break
		}
	}

	// Download the distant disk
	newDisk := MachineDisk(newId)

	err := os.MkdirAll(filepath.Dir(newDisk), 0755)
	if err != nil {
		return err
	}

	err = system.DownloadHttp(fmt.Sprintf("http://%s:%d/machines/%s/disk/data", r.Host, r.Port, m.ID), newDisk)
	if err != nil {
		return err
	}

	// Update parameters
	m.ID = newId

	// TODO: Migrate volumes
	// TODO: Migrate network interfaces

	return nil
}
