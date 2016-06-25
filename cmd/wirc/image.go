package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/image"
)

func listImages(target client.Remote, raw bool) {
	imgs, err := client.ListImages(target)
	if err != nil {
		fatal(err)
	}

	if len(imgs) > 0 {
		if raw {
			for _, i := range imgs {
				fmt.Println(i.Name, image.TypeToString(i.Type), i.Source, i.Arch, i.Distro, i.Release)
			}
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Type", "Source", "Arch", "Distro", "Release"})

			for _, i := range imgs {
				table.Append([]string{i.Name, image.TypeToString(i.Type), i.Source, i.Arch, i.Distro, i.Release})
			}

			table.Render()
		}
	}
}

func createImage(target client.Remote, name, t, source, arch, distro, release string) {
	var req client.ImageRequest
	req.Name = name
	req.Type = image.StringToType(t)
	req.Source = source
	req.Arch = arch
	req.Distro = distro
	req.Release = release

	if !strings.Contains(source, "//") && req.Type != image.TypeLXC {
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
