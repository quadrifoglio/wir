package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/quadrifoglio/wir/shared"
)

var (
	// Global flags
	CRemote = kingpin.Flag("remote", "Remote API server (host:port)").Default("127.0.0.1:8000").String()

	// Image command
	CImageCommand = kingpin.Command("image", "Images manipulation actions")

	CImageList = CImageCommand.Command("list", "List all the images")

	// Image creation
	CImageCreate       = CImageCommand.Command("create", "Create a new image")
	CImageCreateName   = CImageCreate.Flag("name", "Image name").Required().String()
	CImageCreateType   = CImageCreate.Flag("type", "Image type (kvm, vz)").Required().String()
	CImageCreateSource = CImageCreate.Flag("source", "Image source (scheme://[userinfo@][host]/path)").Required().String()

	// Image update
	CImageUpdate     = CImageCommand.Command("update", "Update an image")
	CImageUpdateID   = CImageUpdate.Arg("id", "Image ID").Required().String()
	CImageUpdateName = CImageUpdate.Flag("name", "New image name").Required().String()

	// Image delete
	CImageDelete   = CImageCommand.Command("delete", "Delete an image")
	CImageDeleteID = CImageDelete.Arg("id", "Image ID").Required().String()

	// Network command
	CNetworkCommand = kingpin.Command("network", "Networks manipulation actions")

	CNetworkList = CNetworkCommand.Command("list", "List all the networks")

	// Network creation
	CNetworkCreate             = CNetworkCommand.Command("create", "Create a new network")
	CNetworkCreateName         = CNetworkCreate.Flag("name", "Network name").Required().String()
	CNetworkCreateCIDR         = CNetworkCreate.Flag("cidr", "Network address in CIDR notation").Required().String()
	CNetworkCreateGatewayIface = CNetworkCreate.Flag("gateway-iface", "Physical interface to connect to the network").String()
	CNetworkCreateDhcpEnabled  = CNetworkCreate.Flag("dhcp", "True if the internal DHCP server should be used").Bool()
	CNetworkCreateDhcpStartIP  = CNetworkCreate.Flag("dhcp-start", "First IP address to lease").String()
	CNetworkCreateDhcpNumIP    = CNetworkCreate.Flag("dhcp-count", "Number of IP addresses to lease").Int()
	CNetworkCreateDhcpRouter   = CNetworkCreate.Flag("dhcp-router", "IP address of the router supplied via DHCP").String()

	// Network update
	CNetworkUpdate             = CNetworkCommand.Command("update", "Update a network")
	CNetworkUpdateID           = CNetworkUpdate.Arg("id", "Network ID").Required().String()
	CNetworkUpdateName         = CNetworkUpdate.Flag("name", "Network name").String()
	CNetworkUpdateCIDR         = CNetworkUpdate.Flag("cidr", "Network address in CIDR notation").String()
	CNetworkUpdateGatewayIface = CNetworkUpdate.Flag("gateway-iface", "Physical interface to connect to the network").String()
	// TODO: Update DHCP status
	CNetworkUpdateDhcpStartIP = CNetworkUpdate.Flag("dhcp-start", "First IP address to lease").String()
	CNetworkUpdateDhcpNumIP   = CNetworkUpdate.Flag("dhcp-count", "Number of IP addresses to lease").Int()
	CNetworkUpdateDhcpRouter  = CNetworkUpdate.Flag("dhcp-router", "IP address of the router supplied via DHCP").String()

	// Network delete
	CNetworkDelete   = CNetworkCommand.Command("delete", "Delete a network")
	CNetworkDeleteID = CNetworkDelete.Arg("id", "Network ID").Required().String()
)

// Fatal displays the error and
// then exits with an error code
func Fatal(err error) {
	fmt.Println(err)
	os.Exit(1)
}

// GetRemote returns a RemoteDef structure
// corresponding to the requested remote server
func GetRemote() shared.RemoteDef {
	l := strings.Split(*CRemote, ":")
	if len(l) < 2 {
		Fatal(fmt.Errorf("Invalid remote, must be host:port"))
	}

	port, err := strconv.Atoi(l[1])
	if err != nil {
		Fatal(fmt.Errorf("Invalid remote port"))
	}

	return shared.RemoteDef{l[0], port}
}

func main() {
	switch kingpin.Parse() {
	case "image create":
		ImageCreate()
		break
	case "image list":
		ImageList()
		break
	case "image update":
		ImageUpdate()
		break
	case "image delete":
		ImageDelete()
		break
	case "network create":
		NetworkCreate()
		break
	case "network list":
		NetworkList()
		break
	case "network update":
		NetworkUpdate()
		break
	case "network delete":
		NetworkDelete()
		break
	default:
		Fatal(fmt.Errorf("Invalid command"))
		break
	}
}
