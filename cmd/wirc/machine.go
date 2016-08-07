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
				var macs string
				var ips string

				for _, iface := range m.Interfaces {
					macs += fmt.Sprintf("%s, ", iface.MAC)
					ips += fmt.Sprintf("%s, ", iface.IP)
				}

				macs = macs[:len(macs)-2]
				ips = ips[:len(ips)-2]

				fmt.Println(strconv.Itoa(int(m.Index)), m.Name, m.Type, m.Image, shared.StateToString(m.State), macs, ips)
			}
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Index", "Name", "Type", "Image", "State", "MAC", "IP"})

			for _, m := range ms {
				var macs string
				var ips string

				for _, iface := range m.Interfaces {
					macs += fmt.Sprintf("%s, ", iface.MAC)
					ips += fmt.Sprintf("%s, ", iface.IP)
				}

				if len(macs) > 0 {
					macs = macs[:len(macs)-2]
				}
				if len(ips) > 0 {
					ips = ips[:len(ips)-2]
				}

				table.Append([]string{strconv.Itoa(int(m.Index)), m.Name, m.Type, m.Image, shared.StateToString(m.State), macs, ips})
			}

			table.Render()
		}
	}
}

func createMachine(target shared.Remote, req shared.MachineInfo) {
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
	fmt.Println("Disk:", m.Disk)

	for i, iface := range m.Interfaces {
		fmt.Printf("Interface %d: %s %s %s\n", i, iface.Mode, iface.MAC, iface.IP)
	}
}

func updateMachine(target shared.Remote, req shared.MachineInfo) {
	err := client.UpdateMachine(target, req.Name, req)
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
