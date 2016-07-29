package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/alecthomas/kingpin"

	"github.com/quadrifoglio/wir/shared"
)

var (
	cmdImages    = kingpin.Command("images", "List images")
	cmdImagesRaw = cmdImages.Flag("raw", "Raw listing (no table)").Bool()

	cmdImage = kingpin.Command("image", "Image management")

	remoteAddr = kingpin.Flag("remote-addr", "IP address of the remote server").Default("127.0.0.1").String()
	remotePort = kingpin.Flag("remote-port", "Port of the remote server").Default("1997").Int()
	remoteUser = kingpin.Flag("remote-user", "Username to use in SSH-related actions").Default("root").String()

	imageCreate         = cmdImage.Command("create", "Create an image from the specified source")
	imageCreateType     = imageCreate.Flag("type", "Image type (qemu, lxc, openvz)").Short('t').Default("qemu").String()
	imageCreateMainPart = imageCreate.Flag("main-partition", "Main partition number").Default("0").Int()
	imageCreateArch     = imageCreate.Flag("arch", "Image architecture (386, amd64, arm)").Short('a').Default(runtime.GOARCH).String()
	imageCreateDistro   = imageCreate.Flag("distro", "Image distribution name (debian, ubuntu, ...)").Short('d').Default("").String()
	imageCreateRelease  = imageCreate.Flag("release", "Image's distro release").Short('r').Default("").String()
	imageCreateName     = imageCreate.Arg("name", "Image name").Required().String()
	imageCreateSrc      = imageCreate.Arg("source", "Image source (scheme://[user@][host]/path)").Required().String()

	imageShow     = cmdImage.Command("show", "Show image details")
	imageShowName = imageShow.Arg("name", "Image name").Required().String()

	imageDelete     = cmdImage.Command("delete", "Delete an image")
	imageDeleteName = imageDelete.Arg("name", "Image name").Required().String()

	cmdMachines    = kingpin.Command("machines", "List machines")
	cmdMachinesRaw = cmdMachines.Flag("raw", "Raw listing (no table)").Bool()

	cmdMachine = kingpin.Command("machine", "Machine management")

	machineCreate        = cmdMachine.Command("create", "Create a new machine based on an existing image")
	machineCreateName    = machineCreate.Flag("name", "Name of the machine (will be generated if not specified)").Short('n').String()
	machineCreateCores   = machineCreate.Flag("cpus", "Number of CPU cores to use").Short('c').Default("1").Int()
	machineCreateMem     = machineCreate.Flag("memory", "Amount of memory to use").Short('m').Default("512").Int()
	machineCreateNetMode = machineCreate.Flag("net", "Network setup to use (bridge by default)").String()
	machineCreateNetMAC  = machineCreate.Flag("mac", "MAC address to use").String()
	machineCreateNetIP   = machineCreate.Flag("ip", "IP address to use").String()
	machineCreateImage   = machineCreate.Arg("image", "Name of image to use").Required().String()

	machineShow     = cmdMachine.Command("show", "Show machine details")
	machineShowName = machineShow.Arg("name", "Machine name").Required().String()

	machineUpdate        = cmdMachine.Command("update", "Update a machine")
	machineUpdateCores   = machineUpdate.Flag("cpus", "Number of CPU cores to use").Short('c').Int()
	machineUpdateMem     = machineUpdate.Flag("memory", "Amount of memory to use").Short('m').Int()
	machineUpdateNetMode = machineUpdate.Flag("net", "Network setup to use").String()
	machineUpdateNetMAC  = machineUpdate.Flag("mac", "MAC address to use").String()
	machineUpdateNetIP   = machineUpdate.Flag("ip", "IP address to use").String()
	machineUpdateName    = machineUpdate.Arg("name", "Name of the machine to update").Required().String()

	machineLinuxSysprep           = cmdMachine.Command("linux-sysprep", "Prepare machine for cloning")
	machineLinuxSysprepHostname   = machineLinuxSysprep.Flag("hostname", "Machine hostname").String()
	machineLinuxSysprepRootPasswd = machineLinuxSysprep.Flag("root-password", "Root password").String()
	machineLinuxSysprepName       = machineLinuxSysprep.Arg("name", "Machine name").Required().String()

	machineStart     = cmdMachine.Command("start", "Start machine")
	machineStartName = machineStart.Arg("name", "Machine name").Required().String()

	machineStop     = cmdMachine.Command("stop", "Stop machine")
	machineStopName = machineStop.Arg("name", "Machine name").Required().String()

	machineMigrate       = cmdMachine.Command("migrate", "Migrate machine")
	machineMigrateLive   = machineMigrate.Flag("live", "Live migration").Bool()
	machineMigrateName   = machineMigrate.Arg("name", "Machine name").Required().String()
	machineMigrateTarget = machineMigrate.Arg("target", "Target node (user@ip:port)").Required().String()

	machineDelete     = cmdMachine.Command("delete", "Delete a machine")
	machineDeleteName = machineDelete.Arg("name", "Machine name").Required().String()
)

func fatal(err error) {
	fmt.Println("fatal error:", err)
	os.Exit(1)
}

func main() {
	s := kingpin.Parse()

	var remote shared.Remote
	remote.Addr = *remoteAddr
	remote.APIPort = *remotePort
	remote.SSHUser = *remoteUser

	switch s {
	case "images":
		listImages(remote, *cmdImagesRaw)
		break
	case "image create":
		createImage(
			remote,
			*imageCreateName,
			*imageCreateType,
			*imageCreateSrc,
			*imageCreateMainPart,
			*imageCreateArch,
			*imageCreateDistro,
			*imageCreateRelease,
		)
		break
	case "image show":
		showImage(remote, *imageShowName)
		break
	case "image delete":
		deleteImage(remote, *imageDeleteName)
		break
	case "machines":
		listMachines(remote, *cmdMachinesRaw)
		break
	case "machine create":
		createMachine(
			remote,
			*machineCreateName,
			*machineCreateImage,
			*machineCreateCores,
			*machineCreateMem,
			shared.MachineNetwork{*machineCreateNetMode, *machineCreateNetMAC, *machineCreateNetIP},
		)
		break
	case "machine show":
		showMachine(remote, *machineShowName)
		break
	case "machine update":
		updateMachine(
			remote,
			*machineUpdateName,
			*machineUpdateCores,
			*machineUpdateMem,
			shared.MachineNetwork{*machineUpdateNetMode, *machineUpdateNetMAC, *machineUpdateNetIP},
		)
		break
	case "machine linux-sysprep":
		linuxSysprepMachine(
			remote,
			*machineLinuxSysprepName,
			*machineLinuxSysprepHostname,
			*machineLinuxSysprepRootPasswd,
		)
		break
	case "machine start":
		startMachine(remote, *machineStartName)
		break
	case "machine stop":
		stopMachine(remote, *machineStopName)
		break
	case "machine migrate":
		migrateMachine(remote, *machineMigrateName, *machineMigrateTarget, *machineMigrateLive)
		break
	case "machine delete":
		deleteMachine(remote, *machineDeleteName)
		break
	}
}
