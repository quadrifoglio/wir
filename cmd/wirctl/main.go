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

	// Volume command
	CVolumeCommand = kingpin.Command("volume", "Volume manipulation actions")

	CVolumeList = CVolumeCommand.Command("list", "List all the volumes")

	// Volume creation
	CVolumeCreate     = CVolumeCommand.Command("create", "Create a new volume")
	CVolumeCreateName = CVolumeCreate.Flag("name", "Volume name").Required().String()
	CVolumeCreateType = CVolumeCreate.Flag("type", "Volume type (kvm, vz)").Required().String()
	CVolumeCreateSize = CVolumeCreate.Flag("size", "Volume size in bytes").Required().Uint64()

	// Volume update
	CVolumeUpdate     = CVolumeCommand.Command("update", "Update a volume")
	CVolumeUpdateID   = CVolumeUpdate.Arg("id", "Volume ID").Required().String()
	CVolumeUpdateName = CVolumeUpdate.Flag("name", "Volume name").Required().String()

	// Volume delete
	CVolumeDelete   = CVolumeCommand.Command("delete", "Delete a volume")
	CVolumeDeleteID = CVolumeDelete.Arg("id", "Volume ID").Required().String()

	// Machine command
	CMachineCommand = kingpin.Command("machine", "Machine manipulation actions")

	CMachineList = CMachineCommand.Command("list", "List all the machines")

	// Machine creation
	CMachineCreate       = CMachineCommand.Command("create", "Create a new machine")
	CMachineCreateName   = CMachineCreate.Flag("name", "Machine name").Required().String()
	CMachineCreateImage  = CMachineCreate.Flag("image", "Image ID").String()
	CMachineCreateCores  = CMachineCreate.Flag("cores", "Number of CPUs").Required().Int()
	CMachineCreateMemory = CMachineCreate.Flag("ram", "Quantity of RAM in MiB").Required().Uint64()
	CMachineCreateDisk   = CMachineCreate.Flag("disk", "Maximum disk space in bytes").Uint64()

	// Machine update
	CMachineUpdate       = CMachineCommand.Command("update", "Update a machine")
	CMachineUpdateID     = CMachineUpdate.Arg("id", "Machine ID").Required().String()
	CMachineUpdateName   = CMachineUpdate.Flag("name", "Machine name").String()
	CMachineUpdateCores  = CMachineUpdate.Flag("cores", "Number of CPUs").Int()
	CMachineUpdateMemory = CMachineUpdate.Flag("ram", "Quantity of RAM in MiB").Uint64()
	CMachineUpdateDisk   = CMachineUpdate.Flag("disk", "Maximum disk space in bytes").Uint64()

	// Machine delete
	CMachineDelete   = CMachineCommand.Command("delete", "Delete a machine")
	CMachineDeleteID = CMachineDelete.Arg("id", "Machine ID").Required().String()

	// Machine KVM options
	CMachineKvm = CMachineCommand.Command("kvm", "KVM machine options manipulation")

	CMachineKvmGet   = CMachineKvm.Command("get", "Show KVM options")
	CMachineKvmGetID = CMachineKvmGet.Arg("id", "Machine ID").Required().String()

	CMachineKvmSet           = CMachineKvm.Command("set", "Set KVM options")
	CMachineKvmSetID         = CMachineKvmSet.Arg("id", "Machine ID").Required().String()
	CMachineKvmSetVncEnabled = CMachineKvmSet.Flag("vnc", "VNC server active").Bool()
	CMachineKvmSetVncAddr    = CMachineKvmSet.Flag("vnc-address", "VNC server bind address").String()
	CMachineKvmSetVncPort    = CMachineKvmSet.Flag("vnc-display", "VNC display port").Int()

	// Machine start
	CMachineStart   = CMachineCommand.Command("start", "Start a machine")
	CMachineStartID = CMachineStart.Arg("id", "Machine ID").Required().String()

	// Machine stop
	CMachineStop   = CMachineCommand.Command("stop", "Stop a machine")
	CMachineStopID = CMachineStop.Arg("id", "Machine ID").Required().String()

	// Machine start
	CMachineStatus   = CMachineCommand.Command("status", "Status of a machine")
	CMachineStatusID = CMachineStatus.Arg("id", "Machine ID").Required().String()

	// Machine checkpoints
	CCheckpoint = CMachineCommand.Command("checkpoint", "Checkpoint manipulation actions")

	CCheckpointList        = CCheckpoint.Command("list", "List checkpoints")
	CCheckpointListMachine = CCheckpointList.Arg("machine", "Machine ID").Required().String()

	CCheckpointCreate        = CCheckpoint.Command("create", "Create a checkpoint")
	CCheckpointCreateMachine = CCheckpointCreate.Arg("machine", "Machine ID").Required().String()
	CCheckpointCreateName    = CCheckpointCreate.Arg("name", "Checkpoint name").Required().String()

	CCheckpointDelete        = CCheckpoint.Command("delete", "Delete a checkpoint")
	CCheckpointDeleteMachine = CCheckpointDelete.Arg("machine", "Machine ID").Required().String()
	CCheckpointDeleteName    = CCheckpointDelete.Arg("name", "Checkpoint name").Required().String()

	CCheckpointRestore        = CCheckpoint.Command("restore", "Restore a checkpoint")
	CCheckpointRestoreMachine = CCheckpointRestore.Arg("machine", "Machine ID").Required().String()
	CCheckpointRestoreName    = CCheckpointRestore.Arg("name", "Checkpoint name").Required().String()
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

	case "volume create":
		VolumeCreate()
		break
	case "volume list":
		VolumeList()
		break
	case "volume update":
		VolumeUpdate()
		break
	case "volume delete":
		VolumeDelete()
		break

	case "machine create":
		MachineCreate()
		break
	case "machine list":
		MachineList()
		break
	case "machine update":
		MachineUpdate()
		break
	case "machine delete":
		MachineDelete()
		break

	case "machine kvm get":
		MachineGetKvmOpts()
		break
	case "machine kvm set":
		MachineSetKvmOpts()
		break

	case "machine start":
		MachineStart()
		break
	case "machine stop":
		MachineStop()
		break

	case "machine status":
		MachineStatus()
		break

	case "machine checkpoint create":
		MachineCheckpointCreate()
		break
	case "machine checkpoint list":
		MachineCheckpointList()
		break
	case "machine checkpoint delete":
		MachineCheckpointDelete()
		break
	case "machine checkpoint restore":
		MachineCheckpointRestore()
		break

	default:
		Fatal(fmt.Errorf("Invalid command"))
		break
	}
}
