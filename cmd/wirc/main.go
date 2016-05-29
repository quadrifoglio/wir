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

	machineCreate      = cmdMachine.Command("create", "Create a new machine based on an existing image")
	machineCreateCores = machineCreate.Flag("cpus", "Number of CPU cores to use").Short('c').Default("1").Int()
	machineCreateMem   = machineCreate.Flag("memory", "Amount of memory to use").Short('m').Default("512").Int()
	machineCreateNet   = machineCreate.Flag("net", "Type of networking to use").Short('n').Default("bridge").String()
	machineCreateNetIf = machineCreate.Flag("net-if", "Network interface to use (in case of bridge networking)").Short('i').Default("eth0").String()
	machineCreateImage = machineCreate.Arg("image", "Name of image to use").Required().String()

	machineShow   = cmdMachine.Command("show", "Show machine details")
	machineShowId = machineShow.Arg("id", "Machine ID").Required().String()

	machineStart   = cmdMachine.Command("start", "Start machine")
	machineStartId = machineStart.Arg("id", "Machine ID").Required().String()

	machineStop   = cmdMachine.Command("stop", "Stop machine")
	machineStopId = machineStop.Arg("id", "Machine ID").Required().String()

	machineDelete   = cmdMachine.Command("delete", "Delete a machine")
	machineDeleteId = machineDelete.Arg("id", "Machine ID").Required().String()
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
		net := machineNet{*machineCreateNet, *machineCreateNetIf}
		createMachine(*machineCreateImage, *machineCreateCores, *machineCreateMem, net)
		break
	case "machine show":
		showMachine(*machineShowId)
		break
	case "machine start":
		startMachine(*machineStartId)
		break
	case "machine stop":
		stopMachine(*machineStopId)
		break
	case "machine delete":
		deleteMachine(*machineDeleteId)
		break
	}
}
