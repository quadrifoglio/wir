package main

import (
	"fmt"
	"os"

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
				fmt.Println(n.Name, n.Gateway)
			}
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Gateway"})

			for _, i := range netws {
				table.Append([]string{i.Name, i.Gateway})
			}

			table.Render()
		}
	}
}

func createNetwork(target shared.Remote, name, gateway string) {
	var req shared.Network
	req.Name = name
	req.Gateway = gateway

	_, err := client.CreateNetwork(target, req)
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
