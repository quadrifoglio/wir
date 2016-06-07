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
	table.SetHeader([]string{"Name", "Type", "Image", "State"})

	for _, m := range ms {
		table.Append([]string{m.Name, image.TypeToString(m.Type), m.Image, machine.StateToString(m.State)})
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

	fmt.Println(m.Name)
}

func showMachine(name string) {
	m, err := client.GetMachine(name)
	if err != nil {
		fatal(err)
	}

	fmt.Println("Name:", m.Name)
	fmt.Println("Type:", image.TypeToString(m.Type))
	fmt.Println("Image:", m.Image)
	fmt.Println("State:", machine.StateToString(m.State))
	fmt.Println("Cores:", m.Cores)
	fmt.Println("Memory:", m.Memory)
}

func startMachine(name string) {
	err := client.StartMachine(name)
	if err != nil {
		fatal(err)
	}
}

func stopMachine(name string) {
	err := client.StopMachine(name)
	if err != nil {
		fatal(err)
	}
}

func deleteMachine(name string) {
	err := client.DeleteMachine(name)
	if err != nil {
		fatal(err)
	}
}
