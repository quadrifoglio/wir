package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/shared"
)

func createBackup(target shared.Remote, machine string) {
	bk, err := client.CreateBackup(target, machine)
	if err != nil {
		fatal(err)
	}

	fmt.Println(bk)
}

func listBackups(target shared.Remote, machine string, raw bool) {
	bks, err := client.ListBackups(target, machine)
	if err != nil {
		fatal(err)
	}

	if len(bks) > 0 {
		if raw {
			for _, b := range bks {
				t := time.Unix(int64(b), 0)
				fmt.Println(machine, b, t.Format(time.UnixDate))
			}
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Machine", "Name", "Timestamp"})

			for _, b := range bks {
				t := time.Unix(int64(b), 0)
				ts := t.Format(time.UnixDate)

				table.Append([]string{machine, strconv.FormatInt(int64(b), 10), ts})
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
