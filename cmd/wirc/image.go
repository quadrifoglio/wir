package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/image"
)

func listImages(target client.Remote) {
	imgs, err := client.ListImages(target)
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

func createImage(target client.Remote, name, t, source string) {
	var req client.ImageRequest
	req.Name = name
	req.Type = image.StringToType(t)
	req.Source = source

	if !strings.Contains(source, "//") {
		req.Source = "file://" + source
	}

	_, err := client.CreateImage(target, req)
	if err != nil {
		fatal(err)
	}
}

func showImage(target client.Remote, name string) {
	img, err := client.GetImage(target, name)
	if err != nil {
		fatal(err)
	}

	fmt.Println("Name:", img.Name)
	fmt.Println("Type:", image.TypeToString(img.Type))
	fmt.Println("Source:", img.Source)
}

func deleteImage(target client.Remote, name string) {
	err := client.DeleteImage(target, name)
	if err != nil {
		fatal(err)
	}
}
