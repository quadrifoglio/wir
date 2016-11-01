package main

import (
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/shared"
)

func MachineCheckpointCreate() {
	var req shared.CheckpointDef
	req.Name = *CCheckpointCreateName

	chk, err := client.CheckpointCreate(GetRemote(), *CCheckpointCreateMachine, req)
	if err != nil {
		Fatal(err)
	}

	fmt.Println(chk.Timestamp)
}

func MachineCheckpointList() {
	chks, err := client.CheckpointList(GetRemote(), *CCheckpointListMachine)
	if err != nil {
		Fatal(err)
	}

	if len(chks) > 0 {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{
			"Machine ID",
			"Name",
			"Date",
		})

		for _, chk := range chks {
			table.Append([]string{
				*CCheckpointListMachine,
				chk.Name,
				time.Unix(chk.Timestamp, 0).Format(time.RFC1123),
			})
		}

		table.Render()
	}
}

func MachineCheckpointDelete() {
	err := client.CheckpointDelete(GetRemote(), *CCheckpointDeleteMachine, *CCheckpointDeleteName)
	if err != nil {
		Fatal(err)
	}
}

func MachineCheckpointRestore() {
	err := client.CheckpointRestore(GetRemote(), *CCheckpointRestoreMachine, *CCheckpointRestoreName)
	if err != nil {
		Fatal(err)
	}
}
