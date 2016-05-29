package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/image"
)

func listImages() {
	imgs, err := client.ListImages()
	if err != nil {
		fatal(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Type", "Source"})

	for _, i := range imgs {
		table.Append([]string{i.Name, image.TypeToString(i.Type), i.Source})
	}

	table.Render()
}

func createImage(name, t, source string) {
	var req client.ImageRequest
	req.Name = name
	req.Type = image.StringToType(t)
	req.Source = source

	if !strings.Contains(source, "//") {
		req.Source = "file://" + source
	}

	_, err := client.CreateImage(req)
	if err != nil {
		fatal(err)
	}
}

func showImage(name string) {
	img, err := client.GetImage(name)
	if err != nil {
		fatal(err)
	}

	fmt.Println("Name:", img.Name)
	fmt.Println("Type:", image.TypeToString(img.Type))
	fmt.Println("Source:", img.Source)
}

func deleteImage(name string) {
	err := client.DeleteImage(name)
	if err != nil {
		fatal(err)
	}
}
