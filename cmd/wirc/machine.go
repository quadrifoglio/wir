package main

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/machine"
)

type machineNet struct {
	BrIf string
}

func listMachines() {
	ms, err := client.ListMachines()
	if err != nil {
		fatal(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Type", "Image", "State"})

	for _, m := range ms {
		table.Append([]string{m.ID, m.Name, image.TypeToString(m.Type), m.Image, machine.StateToString(m.State)})
	}

	table.Render()
}

func createMachine(name, img string, cpus, mem int, net machineNet) {
	var req client.MachineRequest
	req.Name = name
	req.Image = img
	req.Cores = cpus
	req.Memory = mem
	req.Network = machine.NetworkMode{net.BrIf}

	m, err := client.CreateMachine(req)
	if err != nil {
		fatal(err)
	}

	fmt.Println(m.ID)
}

func showMachine(id string) {
	m, err := client.GetMachine(id)
	if err != nil {
		fatal(err)
	}

	fmt.Println("ID:", m.ID)
	fmt.Println("Name:", m.Name)
	fmt.Println("Type:", image.TypeToString(m.Type))
	fmt.Println("Image:", m.Image)
	fmt.Println("State:", machine.StateToString(m.State))
	fmt.Println("Cores:", m.Cores)
	fmt.Println("Memory:", m.Memory)
}

func startMachine(id string) {
	err := client.StartMachine(id)
	if err != nil {
		fatal(err)
	}
}

func stopMachine(id string) {
	err := client.StopMachine(id)
	if err != nil {
		fatal(err)
	}
}

func deleteMachine(id string) {
	err := client.DeleteMachine(id)
	if err != nil {
		fatal(err)
	}
}
