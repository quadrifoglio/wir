package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/shared"
)

func createInterface(target shared.Remote, machine, mode, mac, ip string) {
	iface, err := client.CreateInterface(target, machine, mode, mac, ip)
	if err != nil {
		fatal(err)
	}

	fmt.Println(iface.MAC)
}

func listInterfaces(target shared.Remote, machine string, raw bool) {
	ifaces, err := client.ListInterfaces(target, machine)
	if err != nil {
		fatal(err)
	}

	if len(ifaces) > 0 {
		if raw {
			for i, iface := range ifaces {
				fmt.Println(i, iface.Mode, iface.MAC, iface.IP)
			}
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Index", "Mode", "MAC", "IP"})

			for i, iface := range ifaces {
				table.Append([]string{strconv.Itoa(i), iface.Mode, iface.MAC, iface.IP})
			}

			table.Render()
		}
	}
}

func deleteInterface(target shared.Remote, machine string, index int) {
	err := client.DeleteInterface(target, machine, index)
	if err != nil {
		fatal(err)
	}
}
