package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/shared"
)

func listImages(target shared.Remote, raw bool) {
	imgs, err := client.ListImages(target)
	if err != nil {
		fatal(err)
	}

	if len(imgs) > 0 {
		if raw {
			for _, i := range imgs {
				fmt.Println(i.Name, i.Type, i.Source, i.Desc)
			}
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Type", "Source", "Description"})

			for _, i := range imgs {
				table.Append([]string{i.Name, i.Type, i.Source, i.Desc})
			}

			table.Render()
		}
	}
}

func createImage(target shared.Remote, name, t, source, desc string, mainPart int) {
	var req shared.Image
	req.Name = name
	req.Type = t
	req.Source = source
	req.Desc = desc
	req.MainPartition = mainPart

	if !strings.Contains(source, "//") {
		req.Source = "file://" + source
	}

	_, err := client.CreateImage(target, req)
	if err != nil {
		fatal(err)
	}
}

func showImage(target shared.Remote, name string) {
	img, err := client.GetImage(target, name)
	if err != nil {
		fatal(err)
	}

	fmt.Println("Name:", img.Name)
	fmt.Println("Type:", img.Type)
	fmt.Println("Source:", img.Source)
	fmt.Println("Description:", img.Desc)
	fmt.Println("Main parition", img.MainPartition)
}

func deleteImage(target shared.Remote, name string) {
	err := client.DeleteImage(target, name)
	if err != nil {
		fatal(err)
	}
}
