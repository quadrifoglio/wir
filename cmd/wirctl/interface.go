package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/shared"
)

func MachineInterfaceCreate() {
	req, err := client.MachineGet(GetRemote(), *CMachineNicCreateMachine)
	if err != nil {
		Fatal(err)
	}

	var nic shared.InterfaceDef
	nic.Network = *CMachineNicCreateNetwork
	nic.MAC = *CMachineNicCreateMAC
	nic.IP = *CMachineNicCreateIP

	req.Interfaces = append(req.Interfaces, nic)

	_, err = client.MachineUpdate(GetRemote(), req.ID, req)
	if err != nil {
		Fatal(err)
	}
}

func MachineInterfaceList() {
	req, err := client.MachineGet(GetRemote(), *CMachineNicListMachine)
	if err != nil {
		Fatal(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Index",
		"Network ID",
		"IP Address",
		"MAC Address",
	})

	if len(req.Interfaces) > 0 {
		for i, nic := range req.Interfaces {

			table.Append([]string{
				strconv.Itoa(i),
				nic.Network,
				nic.MAC,
				nic.IP,
			})
		}

		table.Render()
	}
}

func MachineInterfaceUpdate() {
	req, err := client.MachineGet(GetRemote(), *CMachineNicUpdateMachine)
	if err != nil {
		Fatal(err)
	}

	index := *CMachineNicUpdateIndex
	if index >= len(req.Interfaces) {
		Fatal(fmt.Errorf("No such interface"))
	}

	if len(*CMachineNicUpdateNetwork) > 0 {
		req.Interfaces[index].Network = *CMachineNicUpdateNetwork
	}
	if len(*CMachineNicUpdateMAC) > 0 {
		req.Interfaces[index].MAC = *CMachineNicUpdateMAC
	}
	if len(*CMachineNicUpdateIP) > 0 {
		req.Interfaces[index].IP = *CMachineNicUpdateIP
	}

	_, err = client.MachineUpdate(GetRemote(), req.ID, req)
	if err != nil {
		Fatal(err)
	}
}

func MachineInterfaceDelete() {
	req, err := client.MachineGet(GetRemote(), *CMachineNicDeleteMachine)
	if err != nil {
		Fatal(err)
	}

	index := *CMachineNicDeleteIndex
	if index >= len(req.Interfaces) {
		Fatal(fmt.Errorf("No such interface"))
	}

	req.Interfaces = append(req.Interfaces[:index], req.Interfaces[index+1:]...)

	_, err = client.MachineUpdate(GetRemote(), req.ID, req)
	if err != nil {
		Fatal(err)
	}
}
