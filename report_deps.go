package main

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/list"
	"github.com/roidaradal/fn/str"
)

// Add dependency report data
func addDepsReport(mod *Module, rep dict.StringMap) {
	// External dependencies
	externalDepsCount := len(mod.Deps.ExternalUsers)
	rep["ExternalDepsCount"] = str.Int(externalDepsCount)
	if externalDepsCount > 0 {
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
		out := []string{
			"<thead><tr>",
			wrapTag(th, "Package"),
			wrapTag(th, "Dependents"),
			wrapTag(th, button("Show Dependents", withID("toggle-deps-external"), onclick("toggleList('deps','external', 'Dependents')"))),
			"</tr></thead><tbody>",
		}
		for _, entry := range entries {
			extPkg, users := entry.Tuple()
			depsList := strings.Join(list.Map(users, nodeToPackageName), "<br/>")
			out = append(out,
				"<tr>",
				wrapTag(td, extPkg),
				wrapTag(th, str.Int(len(users))),
				wrapTag(td, depsList, withClass("deps-external-list hidden")),
				"</tr>",
			)
		}
		rep["ExternalDepsTable"] = strings.Join(out, "") + "</tbody>"
	} else {
		rep["ExternalDepsTable"] = wrapTags("No external packages", tbody, tr, td)
	}

	// Independent packages
	independentCount := len(mod.Deps.Independent)
	rep["IndependentCount"] = str.Int(independentCount)
	if independentCount > 0 {
		out := list.Map(mod.Deps.Independent, func(name string) string {
			return wrapTags(nodeToPackageName(name), tr, td)
		})
		rep["IndependentTable"] = strings.Join(out, "")
	} else {
		rep["IndependentTable"] = wrapTags("No independent packages", tr, td)
	}

	// Dependency packages
	dependentCount := mod.Stats.PackageCount - independentCount
	rep["DependentCount"] = str.Int(dependentCount)
	if dependentCount > 0 {
		out := []string{
			"<thead><tr>",
			wrapTag(th, "Level", withRowspan(2)),
			wrapTag(th, "Package", withRowspan(2)),
			wrapTag(th, "Out", withRowspan(2)),
			wrapTag(th, "In", withRowspan(2)),
			wrapTag(th, button("Show", withID("toggle-deps-dependent"), onclick("toggleList('deps','dependent', '')")), withColspan(2)),
			"</tr><tr>",
			wrapTag(th, "Dependencies"),
			wrapTag(th, "Dependents"),
			"</tr></thead>",
			"<tbody>",
		}
		levels := dict.Keys(mod.Deps.Levels)
		slices.Sort(levels)
		for _, level := range levels {
			for _, subPkg := range mod.Deps.Levels[level] {
				outCount := len(mod.Deps.Of[subPkg])
				inCount := len(mod.Deps.InternalUsers[subPkg])
				var outDetails, inDetails string
				if outCount > 0 {
					outDetails = strings.Join(list.Map(mod.Deps.Of[subPkg], nodeToPackageName), "<br/>")
				}
				if inCount > 0 {
					inDetails = strings.Join(list.Map(mod.Deps.InternalUsers[subPkg], nodeToPackageName), "<br/>")
				}
				out = append(out,
					"<tr>",
					wrapTag(td, str.Int(level), withClass(center)),
					wrapTag(td, nodeToPackageName(subPkg), withClass(left)),
					wrapTag(td, str.Int(outCount), withClass(center)),
					wrapTag(td, str.Int(inCount), withClass(center)),
					wrapTag(td, outDetails, withClass("deps-dependent-list hidden left")),
					wrapTag(td, inDetails, withClass("deps-dependent-list hidden left")),
					"</tr>",
				)
			}
		}
		button := "<button id='view-deps-dependent' onclick='toggleDependentView()'>Show Graph</button><br/>"
		rep["DependencyTable"] = strings.Join(out, "") + "</tbody>"
		rep["DependencyCanvas"] = button + "<canvas id='deps-dependent-graph' width='1000' height='500' class='hidden'></canvas>"
		rep["DependencyEdges"] = strings.Join(mod.Deps.Edges, ", ")
		rep["DependencyNodes"] = strings.Join(list.Map(dict.Entries(mod.Deps.Nodes), func(e dict.Entry[string, string]) string {
			k, v := e.Tuple()
			return fmt.Sprintf("'%s' : %s", k, v)
		}), ",")
	} else {
		rep["DependencyTable"] = wrapTags("No dependent packages", tbody, tr, td)
		rep["DependencyCanvas"] = ""
		rep["DependencyEdges"] = ""
		rep["DependencyNodes"] = ""
	}
}
