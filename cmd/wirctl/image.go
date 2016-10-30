package main

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/shared"
)

// ImageList lists all the images on
// the remote
func ImageList() {
	imgs, err := client.ImageList(GetRemote())
	if err != nil {
		Fatal(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Type", "Source"})

	for _, img := range imgs {
		table.Append([]string{img.ID, img.Name, img.Type, img.Source})
	}

	table.Render()
}

// ImageCreate creates a new
// image on the remote
func ImageCreate() {
	var req shared.ImageDef
	req.Name = *CImageCreateName
	req.Type = *CImageCreateType
	req.Source = *CImageCreateSource

	img, err := client.ImageCreate(GetRemote(), req)
	if err != nil {
		Fatal(err)
	}

	fmt.Println(img.ID)
}

// ImageUpdate updates the specified
// image on the remote
func ImageUpdate() {
	req, err := client.ImageGet(GetRemote(), *CImageUpdateID)
	if err != nil {
		Fatal(err)
	}

	req.Name = *CImageUpdateName

	_, err = client.ImageUpdate(GetRemote(), *CImageUpdateID, req)
	if err != nil {
		Fatal(err)
	}
}

// ImageDelete deletes the specified
// image from the remote
func ImageDelete() {
	err := client.ImageDelete(GetRemote(), *CImageDeleteID)
	if err != nil {
		Fatal(err)
	}
}
