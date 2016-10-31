package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/shared"
)

// MachineList lists all the machines on
// the remote
func MachineList() {
	ms, err := client.MachineList(GetRemote())
	if err != nil {
		Fatal(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"ID",
		"Name",
		"Image",
		"Cores",
		"Memory",
		"Disk",
	})

	for _, m := range ms {
		table.Append([]string{
			m.ID,
			m.Name,
			m.Image,
			strconv.Itoa(m.Cores),
			strconv.FormatUint(m.Memory, 10),
			strconv.FormatUint(m.Disk, 10),
		})
	}

	table.Render()
}

// MachineCreate creates a new
// machine on the remote
func MachineCreate() {
	var req shared.MachineDef
	req.Name = *CMachineCreateName
	req.Image = *CMachineCreateImage
	req.Cores = *CMachineCreateCores
	req.Memory = *CMachineCreateMemory
	req.Disk = *CMachineCreateDisk

	m, err := client.MachineCreate(GetRemote(), req)
	if err != nil {
		Fatal(err)
	}

	fmt.Println(m.ID)
}

// MachineUpdate updates the specified
// machine on the remote
func MachineUpdate() {
	req, err := client.MachineGet(GetRemote(), *CMachineUpdateID)
	if err != nil {
		Fatal(err)
	}

	if len(*CMachineUpdateName) > 0 {
		req.Name = *CMachineUpdateName
	}
	if *CMachineUpdateCores > 0 {
		req.Cores = *CMachineUpdateCores
	}
	if *CMachineUpdateMemory > 0 {
		req.Memory = *CMachineUpdateMemory
	}
	if *CMachineUpdateDisk > 0 {
		req.Disk = *CMachineUpdateDisk
	}

	_, err = client.MachineUpdate(GetRemote(), *CMachineUpdateID, req)
	if err != nil {
		Fatal(err)
	}
}

// MachineDelete deletes the specified
// machine from the remote
func MachineDelete() {
	err := client.MachineDelete(GetRemote(), *CMachineDeleteID)
	if err != nil {
		Fatal(err)
	}
}

// MachineGetKvmOpts shows the KVM-specific
// options of the machine
func MachineGetKvmOpts() {
	opts, err := client.MachineGetKvmOpts(GetRemote(), *CMachineKvmGetID)
	if err != nil {
		Fatal(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Hypervisor PID",
		"VNC Enabled",
		"VNC Address",
		"VNC Port",
	})

	enabled := "false"
	if opts.VNC.Enabled {
		enabled = "true"
	}

	table.Append([]string{
		strconv.Itoa(opts.PID),
		enabled,
		opts.VNC.Address,
		strconv.Itoa(opts.VNC.Port),
	})

	table.Render()
}

// MachineSetKvmOpts sets the KVM-specific
// options of the machine
func MachineSetKvmOpts() {
	req, err := client.MachineGetKvmOpts(GetRemote(), *CMachineKvmSetID)
	if err != nil {
		Fatal(err)
	}

	// TODO: Enable/Disable VNC

	if len(*CMachineKvmSetVncAddr) > 0 {
		req.VNC.Address = *CMachineKvmSetVncAddr
	}
	if *CMachineKvmSetVncPort != 0 {
		req.VNC.Port = *CMachineKvmSetVncPort
	}

	_, err = client.MachineSetKvmOpts(GetRemote(), *CMachineKvmSetID, req)
	if err != nil {
		Fatal(err)
	}
}

// MachineStart starts a
// machine on the remote
func MachineStart() {
	err := client.MachineStart(GetRemote(), *CMachineStartID)
	if err != nil {
		Fatal(err)
	}
}

// MachineStop stops a
// machine on the remote
func MachineStop() {
	err := client.MachineStop(GetRemote(), *CMachineStopID)
	if err != nil {
		Fatal(err)
	}
}

// MachineStatus prints the machine
// status information
func MachineStatus() {
	status, err := client.MachineStatus(GetRemote(), *CMachineStatusID)
	if err != nil {
		Fatal(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Running",
		"CPU Usage (%)",
		"RAM Usage (MiB)",
		"Disk Usage (bytes)",
	})

	running := "false"
	if status.Running {
		running = "true"
	}

	table.Append([]string{
		running,
		strconv.FormatFloat(float64(status.CpuUsage), 'f', 2, 32),
		strconv.FormatUint(status.RamUsage, 10),
		strconv.FormatUint(status.DiskUsage, 10),
	})

	table.Render()
}
