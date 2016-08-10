package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/shared"
)

func createVolume(target shared.Remote, machine, name string, size uint64) {
	vol, err := client.CreateVolume(target, machine, name, size)
	if err != nil {
		fatal(err)
	}

	fmt.Println(vol.Name)
}

func listVolumes(target shared.Remote, machine string, raw bool) {
	vols, err := client.ListVolumes(target, machine)
	if err != nil {
		fatal(err)
	}

	if len(vols) > 0 {
		if raw {
			for _, v := range vols {
				fmt.Println(machine, v.Name, v.Size)
			}
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Machine", "Name", "Size (bytes)"})

			for _, v := range vols {
				table.Append([]string{machine, v.Name, strconv.FormatUint(v.Size, 10)})
			}

			table.Render()
		}
	}
}

func deleteVolume(target shared.Remote, machine, name string) {
	err := client.DeleteVolume(target, machine, name)
	if err != nil {
		fatal(err)
	}
}
