package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/shared"
)

func createBackup(target shared.Remote, machine string) {
	bk, err := client.CreateBackup(target, machine)
	if err != nil {
		fatal(err)
	}

	fmt.Println(bk.Name)
}

func listBackups(target shared.Remote, machine string, raw bool) {
	bks, err := client.ListBackups(target, machine)
	if err != nil {
		fatal(err)
	}

	if len(bks) > 0 {
		if raw {
			for _, b := range bks {
				fmt.Println(machine, b.Name, b.Timestamp)
			}
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Machine", "Name", "Timestamp"})

			for _, b := range bks {
				table.Append([]string{machine, b.Name, strconv.FormatInt(b.Timestamp, 10)})
			}

			table.Render()
		}
	}
}

func restoreBackup(target shared.Remote, machine, name string) {
	err := client.RestoreBackup(target, machine, name)
	if err != nil {
		fatal(err)
	}
}

func deleteBackup(target shared.Remote, machine, name string) {
	err := client.DeleteBackup(target, machine, name)
	if err != nil {
		fatal(err)
	}
}
