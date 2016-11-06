package server

// HandlerMigration - All the migration-related handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/quadrifoglio/go-qemu"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/system"
)

func validateFetch(req shared.MachineFetchDef) (error, int) {
	if len(req.Remote.Host) == 0 {
		return fmt.Errorf("Please specify a 'Remote.Address'"), 400
	}
	if req.Remote.Port == 0 {
		return fmt.Errorf("Please specify a 'Remote.Port'"), 400
	}
	if len(req.ID) != 20 {
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

	// Get the remote machine & image information
	m, err := client.MachineGet(req.Remote, req.ID)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	img, err := client.ImageGet(req.Remote, m.Image)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	// Fetch the remote image, if not already on this host
	err = fetchImage(req.Remote, &img)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	// Fetch the remote machine
	err = fetchMachine(req.Remote, &m)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	// If specified, delete the remote machine after fetching
	if !req.KeepRemote {
		err := client.MachineDelete(req.Remote, req.ID)
		if err != nil {
			ErrorResponse(w, r, err, 500)
			return
		}
	}
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
	var live bool // Will be true if this is a live migration

	status, err := client.MachineStatus(r, m.ID)
	if err != nil {
		return err
	}

	if status.Running {
		live = true

		// If live migration, create a checkpoint first
		_, err := client.CheckpointCreate(r, m.ID, shared.CheckpointDef{Name: "_migration"})
		if err != nil {
			return err
		}
	}

	// Download the distant disk
	newDisk := MachineDisk(m.ID)

	err = os.MkdirAll(filepath.Dir(newDisk), 0755)
	if err != nil {
		return err
	}

	err = system.DownloadHttp(fmt.Sprintf("http://%s:%d/machines/%s/disk/data", r.Host, r.Port, m.ID), newDisk)
	if err != nil {
		return err
	}

	err = MachineKvmCreate(m)
	if err != nil {
		return err
	}

	err = DBMachineCreate(*m)
	if err != nil {
		return err
	}

	// TODO: Migrate volumes
	// TODO: Migrate network interfaces

	// If it is a live migration, restore the created '_migration' checkpoint and delete it
	if live {
		err := client.MachineStop(r, m.ID)
		if err != nil {
			return err
		}

		err = MachineKvmStart(m.ID)
		if err != nil {
			return err
		}

		err = MachineKvmRestoreCheckpoint(m.ID, "_migration")
		if err != nil {
			return err
		}

		img, err := qemu.OpenImage(newDisk)
		if err != nil {
			return err
		}

		err = img.DeleteSnapshot("_migration")
		if err != nil {
			return err
		}
	}

	return nil
}
