package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/shared"
)

func listMachines(target shared.Remote, raw bool) {
	ms, err := client.ListMachines(target)
	if err != nil {
		fatal(err)
	}

	if len(ms) > 0 {
		if raw {
			for _, m := range ms {
				fmt.Println(strconv.Itoa(int(m.Index)), m.Name, m.Type, m.Image, shared.StateToString(m.State))
			}
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Index", "Name", "Type", "Image", "State", "MAC", "IP"})

			for _, m := range ms {
				table.Append([]string{strconv.Itoa(int(m.Index)), m.Name, m.Type, m.Image, shared.StateToString(m.State), m.Network.MAC, m.Network.IP})
			}

			table.Render()
		}
	}
}

func createMachine(target shared.Remote, name, img string, cpus, mem, disk int, net shared.MachineNetwork) {
	var req shared.MachineInfo
	req.Name = name
	req.Image = img
	req.Cores = cpus
	req.Memory = mem
	req.Disk = disk
	req.Network = net

	m, err := client.CreateMachine(target, req)
	if err != nil {
		fatal(err)
	}

	fmt.Println(m.Name)
}

func showMachine(target shared.Remote, name string) {
	m, err := client.GetMachine(target, name)
	if err != nil {
		fatal(err)
	}

	fmt.Println("Index:", m.Index)
	fmt.Println("Name:", m.Name)
	fmt.Println("Type:", m.Type)
	fmt.Println("Image:", m.Image)
	fmt.Println("State:", shared.StateToString(m.State))
	fmt.Println("Cores:", m.Cores)
	fmt.Println("Memory:", m.Memory)
	fmt.Println("Memory:", m.Disk)
	fmt.Println("Net:", m.Network.Mode)
	fmt.Println("MAC:", m.Network.MAC)
	fmt.Println("IP:", m.Network.IP)
}

func updateMachine(target shared.Remote, name string, cpus, mem, disk int, net shared.MachineNetwork) {
	var req shared.MachineInfo
	req.Cores = cpus
	req.Memory = mem
	req.Disk = disk
	req.Network = net

	err := client.UpdateMachine(target, name, req)
	if err != nil {
		fatal(err)
	}
}

func linuxSysprepMachine(target shared.Remote, name, hostname, rootPasswd string) {
	var req client.LinuxSysprep
	req.Hostname = hostname
	req.RootPasswd = rootPasswd

	err := client.LinuxSysprepMachine(target, name, req)
	if err != nil {
		fatal(err)
	}
}

func startMachine(target shared.Remote, name string) {
	err := client.StartMachine(target, name)
	if err != nil {
		fatal(err)
	}
}

func stopMachine(target shared.Remote, name string) {
	err := client.StopMachine(target, name)
	if err != nil {
		fatal(err)
	}
}

func migrateMachine(target shared.Remote, name, remotestr string, live bool) {
	s := strings.Split(remotestr, ":")
	if len(s) <= 1 {
		fatal(fmt.Errorf("target node must be ip:port (ex: 149.91.13.2:8964)"))
	}

	v, err := strconv.Atoi(s[1])
	if err != nil {
		fatal(fmt.Errorf("port must be an integer"))
	}

	err = client.MigrateMachine(target, name, shared.Remote{s[0], target.SSHUser, v}, live)
	if err != nil {
		fatal(err)
	}
}

func deleteMachine(target shared.Remote, name string) {
	err := client.DeleteMachine(target, name)
	if err != nil {
		fatal(err)
	}
}
