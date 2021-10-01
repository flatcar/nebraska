package nebraska_test

import (
	"context"
	"fmt"
	"log"

	nebraska "github.com/kinvolk/nebraska/go-sdk"
)

func ExampleNebraska() {
	// Output:CompileTest!

	cnf := nebraska.Config{
		ServerURL: "http://localhost:8000",
	}
	n, err := nebraska.New(cnf)
	if err != nil {
		log.Fatal("client setup error:", err)
	}
	ctx := context.Background()

	app, err := n.Applications().Create(ctx, nebraska.AppConfig{
		Name: "testApp123",
	}, nil)
	if err != nil {
		log.Fatal("create app error:", err)
	}
	fmt.Printf("App:\n%+v\n", app.Props())

	app, err = app.Update(ctx, nebraska.AppConfig{
		Name: "testAppXYZ",
	}, nil)
	if err != nil {
		log.Fatal("update app error:", err)
	}
	fmt.Printf("Updated App:\n%+v\n", app.Props())

	count := 0
	for {
		fmt.Println(count)
		count++
		group, err := app.CreateGroup(ctx, nebraska.GroupConfig{
			Name:                      fmt.Sprintf("Group %d", count),
			PolicyMaxUpdatesPerPeriod: count,
			PolicyPeriodInterval:      "1 hours",
			PolicyUpdateTimeout:       "1 hours",
		}, nil)
		if err != nil {
			fmt.Println("create group error", count, err)
			return
		}
		if count == 20 {
			break
		}
		fmt.Println("Group:", group.Props().Name)
	}

	page := 0
	for {
		page++
		aGroups, err := app.PaginateGroups(ctx, page, 10, nil)
		if err != nil {
			fmt.Println("paginate groups err", err)
			return
		}
		for _, grp := range aGroups.Groups {
			_, err = grp.Update(ctx, nebraska.GroupConfig{
				Name:                      "test " + grp.Props().Name,
				PolicyMaxUpdatesPerPeriod: grp.Props().PolicyMaxUpdatesPerPeriod,
				PolicyPeriodInterval:      grp.Props().PolicyPeriodInterval,
				PolicyUpdateTimeout:       grp.Props().PolicyUpdateTimeout,
			}, nil)
			if err != nil {
				return
			}
		}
		if page == 2 {
			break
		}
	}
}
