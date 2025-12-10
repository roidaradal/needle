package main

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/str"
)

// Add dependency report data
func addDepsReport(mod *Module, rep dict.StringMap) {
	// External dependencies
	externalDepsCount := len(mod.Deps.ExternalUsers)
	entries := dict.Entries(mod.Deps.ExternalUsers)
	slices.SortFunc(entries, func(a, b dict.Entry[string, []string]) int {
		// Sort by descending order of dependent counts
		score1 := cmp.Compare(len(b.Value), len(a.Value))
		if score1 != 0 {
			return score1
		}
		// Tie-breaker: alphabetical
		return cmp.Compare(a.Key, b.Key)
	})
	out := make([]string, 0)
	for _, entry := range entries {
		extPkg, users := entry.Tuple()
		out = append(out,
			fmt.Sprintf("<li>(<b>%d</b>) %s<ul>", len(users), extPkg),
			listItems(users),
			"</ul></li>",
		)
	}
	rep["ExternalDepsCount"] = str.Int(externalDepsCount)
	rep["ExternalDepsList"] = strings.Join(out, "")

	// Independent packages
	independentCount := len(mod.Deps.Independent)
	rep["IndependentCount"] = str.Int(independentCount)
	rep["IndependentList"] = listItems(mod.Deps.Independent)

	// Dependency packages
	out = make([]string, 0)
	levels := dict.Keys(mod.Deps.Levels)
	slices.Sort(levels)
	for _, level := range levels {
		for _, subPkg := range mod.Deps.Levels[level] {
			outCount := len(mod.Deps.Of[subPkg])
			inCount := len(mod.Deps.InternalUsers[subPkg])
			outDetails, inDetails := " ", " "
			if outCount > 0 {
				outDetails = strings.Join(mod.Deps.Of[subPkg], "<br/>")
			}
			if inCount > 0 {
				inDetails = strings.Join(mod.Deps.InternalUsers[subPkg], "<br/>")
			}
			out = append(out,
				"<tr>",
				fmt.Sprintf("<td class='center'>%d</td>", level),
				fmt.Sprintf("<td>%s</td>", subPkg),
				fmt.Sprintf("<td class='center'>%d</td>", outCount),
				fmt.Sprintf("<td class='center'>%d</td>", inCount),
				fmt.Sprintf("<td>%s</td>", outDetails),
				fmt.Sprintf("<td>%s</td>", inDetails),
				"</tr>",
			)
		}
	}
	dependentCount := mod.Stats.PackageCount - independentCount
	rep["DependentCount"] = str.Int(dependentCount)
	rep["DependencyLevels"] = strings.Join(out, "")
}
