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
	machineCreateImage = machineCreate.Arg("image", "Name of image to use").Required().String()

	machineShow     = cmdMachine.Command("show", "Show machine details")
	machineShowName = machineShow.Arg("id", "Machine ID").Required().String()

	machineDelete     = cmdMachine.Command("delete", "Delete a machine")
	machineDeleteName = machineDelete.Arg("id", "Machine ID").Required().String()
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
		createMachine(*machineCreateImage)
		break
	case "machine show":
		showMachine(*machineShowName)
		break
	case "machine delete":
		deleteMachine(*machineDeleteName)
		break
	}
}
