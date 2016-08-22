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

func listNetworks(target shared.Remote, raw bool) {
	netws, err := client.ListNetworks(target)
	if err != nil {
		fatal(err)
	}

	if len(netws) > 0 {
		if raw {
			for _, n := range netws {
				var net string

				if len(n.Addr) > 0 {
					net = fmt.Sprintf("%s/%d", n.Addr, n.Mask)
				}

				fmt.Println(n.Name, net, n.Router, n.Gateway)
			}
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Address", "Router", "Gateway"})

			for _, n := range netws {
				var net string

				if len(n.Addr) > 0 {
					net = fmt.Sprintf("%s/%d", n.Addr, n.Mask)
				}

				table.Append([]string{n.Name, net, n.Router, n.Gateway})
			}

			table.Render()
		}
	}
}

func createNetwork(target shared.Remote, name, addr, router, gateway string) {
	var req shared.Network
	req.Name = name
	req.Router = router
	req.Gateway = gateway

	if len(addr) > 0 {
		pts := strings.Split(addr, "/")
		if len(pts) != 2 {
			fatal(fmt.Errorf("network address must be <ip>/<mask>"))
		}

		req.Addr = pts[0]

		mask, err := strconv.Atoi(pts[1])
		if err != nil {
			fatal(fmt.Errorf("invalid network mask"))
		}

		req.Mask = mask
	}

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
