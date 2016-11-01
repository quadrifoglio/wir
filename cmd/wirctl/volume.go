package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/shared"
)

// VolumeList lists all the volumes on
// the remote
func VolumeList() {
	vols, err := client.VolumeList(GetRemote())
	if err != nil {
		Fatal(err)
	}

	if len(vols) > 0 {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{
			"ID",
			"Name",
			"Type",
			"Size",
		})

		for _, vol := range vols {
			table.Append([]string{
				vol.ID,
				vol.Name,
				vol.Type,
				strconv.FormatUint(vol.Size, 10),
			})
		}

		table.Render()
	}
}

// VolumeCreate creates a new
// volume on the remote
func VolumeCreate() {
	var req shared.VolumeDef
	req.Name = *CVolumeCreateName
	req.Type = *CVolumeCreateType
	req.Size = *CVolumeCreateSize

	vol, err := client.VolumeCreate(GetRemote(), req)
	if err != nil {
		Fatal(err)
	}

	fmt.Println(vol.ID)
}

// VolumeUpdate updates the specified
// volume on the remote
func VolumeUpdate() {
	req, err := client.VolumeGet(GetRemote(), *CVolumeUpdateID)
	if err != nil {
		Fatal(err)
	}

	if len(*CVolumeUpdateName) > 0 {
		req.Name = *CVolumeUpdateName
	}

	_, err = client.VolumeUpdate(GetRemote(), *CVolumeUpdateID, req)
	if err != nil {
		Fatal(err)
	}
}

// VolumeDelete deletes the specified
// volume from the remote
func VolumeDelete() {
	err := client.VolumeDelete(GetRemote(), *CVolumeDeleteID)
	if err != nil {
		Fatal(err)
	}
}
