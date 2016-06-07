package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kingpin"
)

var (
	cmdImages = kingpin.Command("images", "List images")
	cmdImage  = kingpin.Command("image", "Image management")

	imageCreate     = cmdImage.Command("create", "Create an image from the specified source")
	imageCreateType = imageCreate.Flag("type", "Image type (qemu, docker, openvz)").Short('t').Default("qemu").String()
	imageCreateName = imageCreate.Arg("name", "Image name").Required().String()
	imageCreateSrc  = imageCreate.Arg("source", "Image source (scheme://[user@][host]/path)").Required().String()

	imageShow     = cmdImage.Command("show", "Show image details")
	imageShowName = imageShow.Arg("name", "Image name").Required().String()

	imageDelete     = cmdImage.Command("delete", "Delete an image")
	imageDeleteName = imageDelete.Arg("name", "Image name").Required().String()

	cmdMachines = kingpin.Command("machines", "List machines")
	cmdMachine  = kingpin.Command("machine", "Machine management")

	machineCreate        = cmdMachine.Command("create", "Create a new machine based on an existing image")
	machineCreateName    = machineCreate.Flag("name", "Name of the machine (will be generated if not specified)").Short('n').String()
	machineCreateCores   = machineCreate.Flag("cpus", "Number of CPU cores to use").Short('c').Default("1").Int()
	machineCreateMem     = machineCreate.Flag("memory", "Amount of memory to use").Short('m').Default("512").Int()
	machineCreateNetBrIf = machineCreate.Flag("bridge-if", "Network interface to use (in case of bridge networking)").Short('b').String()
	machineCreateImage   = machineCreate.Arg("image", "Name of image to use").Required().String()

	machineShow     = cmdMachine.Command("show", "Show machine details")
	machineShowName = machineShow.Arg("id", "Machine name").Required().String()

	machineStart     = cmdMachine.Command("start", "Start machine")
	machineStartName = machineStart.Arg("id", "Machine name").Required().String()

	machineStop     = cmdMachine.Command("stop", "Stop machine")
	machineStopName = machineStop.Arg("id", "Machine name").Required().String()

	machineDelete     = cmdMachine.Command("delete", "Delete a machine")
	machineDeleteName = machineDelete.Arg("id", "Machine name").Required().String()
)

func fatal(err error) {
	fmt.Println("Fatal error:", err)
	os.Exit(1)
}

func main() {
	switch kingpin.Parse() {
	case "images":
		listImages()
		break
	case "image create":
		createImage(*imageCreateName, *imageCreateType, *imageCreateSrc)
		break
	case "image show":
		showImage(*imageShowName)
		break
	case "image delete":
		deleteImage(*imageDeleteName)
		break
	case "machines":
		listMachines()
		break
	case "machine create":
		net := machineNet{*machineCreateNetBrIf}
		createMachine(*machineCreateName, *machineCreateImage, *machineCreateCores, *machineCreateMem, net)
		break
	case "machine show":
		showMachine(*machineShowName)
		break
	case "machine start":
		startMachine(*machineStartName)
		break
	case "machine stop":
		stopMachine(*machineStopName)
		break
	case "machine delete":
		deleteMachine(*machineDeleteName)
		break
	}
}
