package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/shared"
)

func listNetworks(target shared.Remote, raw bool) {
	netws, err := client.ListNetworks(target)
	if err != nil {
		fatal(err)
	}

	if len(netws) > 0 {
		if raw {
			for _, n := range netws {
				if n.UseDHCP {
					fmt.Println(n.Name, n.Gateway, n.Addr, n.Mask, n.Router, n.StartIP, n.NumIP)
				} else {
					fmt.Println(n.Name, n.Gateway, n.Addr, n.Mask)
				}
			}
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Gateway", "Address", "Netmask", "Router", "Start IP", "Number of leases"})

			for _, n := range netws {
				table.Append([]string{n.Name, n.Gateway, n.Addr, n.Mask, n.Router, n.StartIP, strconv.Itoa(n.NumIP)})
			}

			table.Render()
		}
	}
}

func createNetwork(target shared.Remote, netw shared.Network) {
	_, err := client.CreateNetwork(target, netw)
	if err != nil {
		fatal(err)
	}
}

func deleteNetwork(target shared.Remote, name string) {
	err := client.DeleteNetwork(target, name)
	if err != nil {
		fatal(err)
	}
}
