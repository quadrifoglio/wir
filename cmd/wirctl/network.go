package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/shared"
)

// NetworkList lists all the networks on
// the remote
func NetworkList() {
	netws, err := client.NetworkList(GetRemote())
	if err != nil {
		Fatal(err)
	}

	if len(netws) > 0 {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{
			"ID",
			"Name",
			"CIDR",
			"Gateway Interface",
			"Using internal DHCP",
			"DHCP First IP",
			"DHCP IP Count",
			"DHCP Router",
		})

		for _, netw := range netws {
			enabled := "false"
			if netw.DHCP.Enabled {
				enabled = "true"
			}

			table.Append([]string{
				netw.ID,
				netw.Name,
				netw.CIDR,
				netw.GatewayIface,
				enabled,
				netw.DHCP.StartIP,
				strconv.Itoa(netw.DHCP.NumIP),
				netw.DHCP.Router,
			})
		}

		table.Render()
	}
}

// NetworkCreate creates a new
// network on the remote
func NetworkCreate() {
	var req shared.NetworkDef
	req.Name = *CNetworkCreateName
	req.CIDR = *CNetworkCreateCIDR
	req.GatewayIface = *CNetworkCreateGatewayIface
	req.DHCP.Enabled = *CNetworkCreateDhcpEnabled
	req.DHCP.StartIP = *CNetworkCreateDhcpStartIP
	req.DHCP.NumIP = *CNetworkCreateDhcpNumIP
	req.DHCP.Router = *CNetworkCreateDhcpRouter

	netw, err := client.NetworkCreate(GetRemote(), req)
	if err != nil {
		Fatal(err)
	}

	fmt.Println(netw.ID)
}

// NetworkUpdate updates the specified
// network on the remote
func NetworkUpdate() {
	req, err := client.NetworkGet(GetRemote(), *CNetworkUpdateID)
	if err != nil {
		Fatal(err)
	}

	if len(*CNetworkUpdateName) > 0 {
		req.Name = *CNetworkUpdateName
	}
	if len(*CNetworkUpdateCIDR) > 0 {
		req.CIDR = *CNetworkUpdateCIDR
	}
	if len(*CNetworkUpdateGatewayIface) > 0 {
		req.GatewayIface = *CNetworkUpdateGatewayIface
	}
	if len(*CNetworkUpdateDhcpStartIP) > 0 {
		req.DHCP.StartIP = *CNetworkUpdateDhcpStartIP
	}
	if *CNetworkUpdateDhcpNumIP > 0 {
		req.DHCP.NumIP = *CNetworkUpdateDhcpNumIP
	}
	if len(*CNetworkUpdateDhcpRouter) > 0 {
		req.DHCP.Router = *CNetworkUpdateDhcpRouter
	}

	_, err = client.NetworkUpdate(GetRemote(), *CNetworkUpdateID, req)
	if err != nil {
		Fatal(err)
	}
}

// NetworkDelete deletes the specified
// network from the remote
func NetworkDelete() {
	err := client.NetworkDelete(GetRemote(), *CNetworkDeleteID)
	if err != nil {
		Fatal(err)
	}
}
