package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/machine"
)

type machineNet struct {
	BrIf string
}

func listMachines(target client.Remote, raw bool) {
	ms, err := client.ListMachines(target)
	if err != nil {
		fatal(err)
	}

	if len(ms) > 0 {
		if raw {
			for _, m := range ms {
				fmt.Println(strconv.Itoa(int(m.Index)), m.Name, image.TypeToString(m.Type), m.Image, machine.StateToString(m.State))
			}
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Index", "Name", "Type", "Image", "State"})

			for _, m := range ms {
				table.Append([]string{strconv.Itoa(int(m.Index)), m.Name, image.TypeToString(m.Type), m.Image, machine.StateToString(m.State)})
			}

			table.Render()
		}
	}
}

func createMachine(target client.Remote, name, img string, cpus, mem int, net machineNet) {
	var req client.MachineRequest
	req.Name = name
	req.Image = img
	req.Cores = cpus
	req.Memory = mem
	req.Network = machine.NetworkMode{net.BrIf}

	m, err := client.CreateMachine(target, req)
	if err != nil {
		fatal(err)
	}

	fmt.Println(m.Name)
}

func showMachine(target client.Remote, name string) {
	m, err := client.GetMachine(target, name)
	if err != nil {
		fatal(err)
	}

	fmt.Println("Index:", m.Index)
	fmt.Println("Name:", m.Name)
	fmt.Println("Type:", image.TypeToString(m.Type))
	fmt.Println("Image:", m.Image)
	fmt.Println("State:", machine.StateToString(m.State))
	fmt.Println("Cores:", m.Cores)
	fmt.Println("Memory:", m.Memory)
}

func startMachine(target client.Remote, name string) {
	err := client.StartMachine(target, name)
	if err != nil {
		fatal(err)
	}
}

func stopMachine(target client.Remote, name string) {
	err := client.StopMachine(target, name)
	if err != nil {
		fatal(err)
	}
}

func deleteMachine(target client.Remote, name string) {
	err := client.DeleteMachine(target, name)
	if err != nil {
		fatal(err)
	}
}
