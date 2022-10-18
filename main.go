package main

import (
	"encoding/json"
	"os"

	"github.com/o98k-ok/alfred-manager-flow/core"
	"github.com/o98k-ok/lazy/v2/alfred"
)

func main() {
	cli := alfred.NewApp("alfred workflows manager plugin")
	cli.Bind("get", core.GetWorkflows)
	cli.Bind("path", func(s []string) {
		if len(s) <= 0 {
			alfred.Deliver(alfred.EmptyItems().Encode())
			return
		}

		var flow core.Flow
		if err := json.Unmarshal([]byte(s[0]), &flow); err != nil {
			alfred.Deliver(alfred.ErrItems("get path error", err).Encode())
			return
		}

		alfred.Deliver(flow.Path)
	})

	cli.Bind("envs", func(s []string) {
		if len(s) <= 0 {
			alfred.Deliver(alfred.EmptyItems().Encode())
			return
		}

		var flow core.Flow
		if err := json.Unmarshal([]byte(s[0]), &flow); err != nil {
			alfred.Deliver(alfred.ErrItems("get envs error", err).Encode())
			return
		}

		var result alfred.Items
		for k, v := range flow.Envs {
			result.Append(alfred.NewItem(k, v, v))
		}
		alfred.Deliver(result.Encode())
	})

	cli.Bind("url", func(s []string) {
		if len(s) <= 0 {
			alfred.Deliver(alfred.EmptyItems().Encode())
			return
		}

		var flow core.Flow
		if err := json.Unmarshal([]byte(s[0]), &flow); err != nil {
			alfred.Deliver(alfred.ErrItems("get envs error", err).Encode())
			return
		}
		alfred.Deliver(flow.WebSite)
	})

	err := cli.Run(os.Args)
	if err != nil {
		alfred.ErrItems("run error", err)
		return
	}
}
