package main

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/yangl900/kui-sample/data"
)

var (
	mainView *tview.TextView
)

func main() {
	newPrimitive := func(text string) *tview.TextView {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(text)
	}
	mainView = newPrimitive("Main content")
	footer := newPrimitive("Some important text")

	app := tview.NewApplication()
	list := tview.NewList()
	list.ShowSecondaryText(false)

	clusters, err := data.GetClusterState()
	if err != nil {
		footer.SetText(err.Error())
	}

	for _, c := range clusters {
		cc := c
		list.AddItem(c.Context, fmt.Sprintf("%d/%d", len(*c.Nodes), len(*c.Pods)), ' ', func() { render(mainView, cc) })
	}
	list.AddItem("Quit", "Press to exit", 'q', func() {
		app.Stop()
	})

	grid := tview.NewGrid().
		SetRows(1, 0, 1).
		SetColumns(30, 0).
		SetBorders(true).
		AddItem(newPrimitive("Fleet Viewer"), 0, 0, 1, 3, 0, 0, false).
		AddItem(footer, 2, 0, 1, 3, 0, 0, false)

	// Layout for screens narrower than 100 cells (menu and side bar are hidden).
	grid.AddItem(list, 1, 0, 1, 1, 0, 0, false).
		AddItem(mainView, 1, 1, 1, 2, 0, 0, false)

	if err := app.SetRoot(grid, true).SetFocus(list).Run(); err != nil {
		panic(err)
	}
}

func render(v *tview.TextView, c data.ClusterState) {
	pools := map[string][]string{}
	for _, n := range *c.Nodes {
		if pool, ok := n.Labels["pool"]; ok {
			if nodes, ok := pools[pool]; ok {
				nodes = append(nodes, n.Name)
			} else {
				pools[pool] = []string{n.Name}
			}
		}
	}
	v.SetText(fmt.Sprintf("Total Nodes: %d\nTotal Pods: %d\nNode Pools: %d", len(*c.Nodes), len(*c.Pods), len(pools)))
}
