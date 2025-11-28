package needle

import (
	"fmt"
	"slices"
	"strings"

	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/ds"
	"github.com/roidaradal/fn/list"
	"github.com/roidaradal/fn/number"
)

// Build new StatsModule for Go module at path
func NewStatsModule(path string) (*StatsModule, error) {
	// Initialize stats module
	baseMod, err := baseModule(path)
	if err != nil {
		return nil, err
	}
	mod := newStatsModule(baseMod)

	// Run concurrently
	cfg := &taskConfig[TreeEntry, *Package]{
		Task: func(entry TreeEntry) (*Package, error) {
			folder, node := entry.Tuple()
			return newPackage(mod.Module, folder, node.Files)
		},
		Receive: func(pkg *Package) {
			mod.Packages = append(mod.Packages, pkg)
		},
	}
	entries := mod.ValidTreeEntries()
	err = runConcurrent(entries, cfg)
	if err != nil {
		return nil, err
	}

	return mod, nil
}

// Build Package object for Go module at path
func newPackage(mod *Module, path string, files []string) (*Package, error) {
	rootFolder := mod.Path + path
	pkg := &Package{Name: getPackageName(path)}

	// Run concurrently
	cfg := &taskConfig[string, *File]{
		Task: func(filename string) (*File, error) {
			path := fmt.Sprintf("%s/%s", rootFolder, filename)
			return newFile(path)
		},
		Receive: func(file *File) {
			pkg.Files = append(pkg.Files, file)
		},
	}
	err := runConcurrent(files, cfg)
	if err != nil {
		return nil, err
	}
	return pkg, nil
}

// Build File object for file path
func newFile(path string) (*File, error) {
	file := &File{
		Name: getFilename(path),
		Type: getFileType(path),
	}
	lines, err := ioReadRawLines(path)
	if err != nil {
		return nil, err
	}

	for _, rawLine := range lines {
		var line *Line
		cleanLine := strings.TrimSpace(rawLine)
		if cleanLine == "" {
			line = &Line{Type: LINE_SPACE, Length: 1}
		} else if strings.HasPrefix(cleanLine, "// ") {
			line = &Line{Type: LINE_COMMENT, Length: len(rawLine)}
		} else {
			line = &Line{Length: len(rawLine)}
		}
		file.Lines = append(file.Lines, line)
	}

	return file, nil
}

// StatsModule string representation
func (mod StatsModule) String() string {
	out := []string{
		fmt.Sprintf("Name: %s", mod.Name),
		mod.PerPackageFileCount(),
		mod.PerPackageFilesLineCount(),
	}
	return strings.Join(out, "\n")
}

// Per package: count files
func (mod StatsModule) PerPackageFileCount() string {
	out := make([]string, 0)
	pkgNames := mod.PackageNames()
	fileCounts := list.Map(mod.Packages, func(pkg *Package) int {
		return len(pkg.Files)
	})
	pkgFileCounts := dict.Entries(dict.Zip(pkgNames, fileCounts))
	slices.SortFunc(pkgFileCounts, sortDescCount)
	totalFileCount := list.Sum(fileCounts)
	out = append(out, fmt.Sprintf("Pkg: %d, TotalFiles: %s", len(pkgFileCounts), number.Comma(totalFileCount)))

	// Check if has test files
	totalTestCount := mod.CountTestFiles()
	hasTest := false
	if totalTestCount > 0 {
		hasTest = true
		totalCodeCount := totalFileCount - totalTestCount
		codeRatio := percentage(totalCodeCount, totalFileCount)
		testRatio := percentage(totalTestCount, totalFileCount)
		out = append(out, fmt.Sprintf("Code: %d (%s), Test: %d (%s)", totalCodeCount, codeRatio, totalTestCount, testRatio))
	}

	lookup := ds.NewLookupCode(mod.Packages)
	for _, e := range pkgFileCounts {
		pkgName, count := e.Tuple()
		pkg := lookup[pkgName]
		testCount := pkg.CountTestFiles()
		ratio := percentage(count, totalFileCount)
		if hasTest {
			codeCount := count - testCount
			out = append(out, fmt.Sprintf("\t%4s : %2d | %2d | %2d : %s", ratio, count, codeCount, testCount, pkgName))
		} else {
			out = append(out, fmt.Sprintf("\t%4s : %2d : %s", ratio, count, pkgName))
		}
	}
	return strings.Join(out, "\n")
}

// Per package: list files x line count per file
func (mod StatsModule) PerPackageFilesLineCount() string {
	out := make([]string, 0)

	// Total files, total lines, and average line per file
	totalFileCount := mod.CountFiles()
	totalLineCount := mod.CountLines()
	globalAvg := float64(totalLineCount) / float64(totalFileCount)
	out = append(out, fmt.Sprintf("TotalFiles: %d, TotalLines: %s, AvgLinePerFile: %.1f", totalFileCount, number.Comma(totalLineCount), globalAvg))

	testLineCount := mod.CountTestLines()
	if testLineCount > 0 {
		testFileCount := mod.CountTestFiles()
		codeLineCount := totalLineCount - testLineCount
		codeFileCount := totalFileCount - testFileCount
		codeRatio := percentage(codeLineCount, totalLineCount)
		testRatio := percentage(testLineCount, totalLineCount)
		codeALPF := float64(codeLineCount) / float64(codeFileCount)
		testALPF := float64(testLineCount) / float64(testFileCount)
		out = append(out, fmt.Sprintf("CodeFiles: %d, CodeLines: %s (%s), CodeALPF: %.1f", codeFileCount, number.Comma(codeLineCount), codeRatio, codeALPF))
		out = append(out, fmt.Sprintf("TestFiles: %d, TestLines: %s (%s), TestALPF: %.1f", testFileCount, number.Comma(testLineCount), testRatio, testALPF))
	}

	// Sort packages by total lines
	pkgNames := mod.PackageNames()
	lineCounts := list.Map(mod.Packages, (*Package).CountLines)
	pkgLineCounts := dict.Entries(dict.Zip(pkgNames, lineCounts))
	slices.SortFunc(pkgLineCounts, sortDescCount)

	lookup := ds.NewLookupCode(mod.Packages)
	for _, e := range pkgLineCounts {
		pkgName, pkgCount := e.Tuple()
		pkg := lookup[pkgName]
		fileNames := pkg.FileNames()
		pkgRatio := percentage(pkgCount, totalLineCount)
		countStr := number.Comma(pkgCount) + " lines"
		alpf := float64(pkgCount) / float64(len(fileNames))
		out = append(out, fmt.Sprintf("\t%4s : %-12s : %s (%.1f)", pkgRatio, countStr, pkgName, alpf))
		// Sort files by line count
		lineCounts = list.Map(pkg.Files, (*File).CountLines)
		fileLineCounts := dict.Entries(dict.Zip(fileNames, lineCounts))
		slices.SortFunc(fileLineCounts, sortDescCount)
		for _, f := range fileLineCounts {
			fileName, count := f.Tuple()
			globalRatio := percentage(count, totalLineCount)
			localRatio := percentage(count, pkgCount)
			out = append(out, fmt.Sprintf("\t\t*%4s :%4s : %6s : %s", globalRatio, localRatio, number.Comma(count), fileName))
		}
	}

	return strings.Join(out, "\n")
}
