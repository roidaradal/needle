package needle

import (
	"fmt"
	"slices"
	"strings"

	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/ds"
	"github.com/roidaradal/fn/io"
	"github.com/roidaradal/fn/lang"
	"github.com/roidaradal/fn/list"
	"github.com/roidaradal/fn/number"
)

// Build new StatsModule for Go module at path
func NewStatsModule(cfg *Config) (*StatsModule, error) {
	// Initialize stats module
	baseMod, err := baseModule(cfg)
	if err != nil {
		return nil, err
	}
	mod := newStatsModule(baseMod)

	// Run concurrently
	taskCfg := &taskConfig[TreeEntry, *Package]{
		Task: func(entry TreeEntry) (*Package, error) {
			folder, node := entry.Tuple()
			return newPackage(mod.Module, folder, node.Files)
		},
		Receive: func(pkg *Package) {
			mod.Packages = append(mod.Packages, pkg)
		},
	}
	entries := mod.ValidTreeEntries()
	err = runConcurrent(entries, taskCfg)
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
			return newFile(pkg, path)
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
func newFile(pkg *Package, path string) (*File, error) {
	file := &File{
		Name: getFilename(path),
		Type: getFileType(path),
	}
	lines, err := io.ReadRawLines(path)
	if err != nil {
		return nil, err
	}

	for _, rawLine := range lines {
		var line *Line
		cleanLine := strings.TrimSpace(rawLine)
		rawCount := len(rawLine)
		if cleanLine == "" {
			line = &Line{Type: LINE_SPACE, Length: 1}
		} else if strings.HasPrefix(cleanLine, "// ") {
			line = &Line{Type: LINE_COMMENT, Length: rawCount}
		} else if strings.HasPrefix(cleanLine, "package ") {
			line = &Line{Type: LINE_HEAD, Length: rawCount}
			if name, ok := getLinePart(cleanLine, 1); ok {
				pkg.Type = lang.Ternary(name == "main", PKG_MAIN, PKG_LIB)
			}
		} else {
			line = &Line{Type: LINE_CODE, Length: rawCount}
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
		mod.PerPackageFilesCharCount(),
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
	totalFileCount := list.Sum(fileCounts)
	out = append(out, fmt.Sprintf("Pkg: %d, TotalFiles: %s", len(pkgNames), number.Comma(totalFileCount)))
	out = append(out, fmt.Sprintf("LibPkg: %d, MainPkg: %d", mod.CountLibPackages(), mod.CountMainPackages()))

	// Check if has test files
	totalTestCount := mod.CountTestFiles()
	hasTest := false
	if totalTestCount > 0 {
		hasTest = true
		totalCodeCount := totalFileCount - totalTestCount
		codeRatio := percentage(totalCodeCount, totalFileCount)
		testRatio := percentage(totalTestCount, totalFileCount)
		out = append(out, fmt.Sprintf("- Code: %d (%s), Test: %d (%s)", totalCodeCount, codeRatio, totalTestCount, testRatio))
	}

	if !mod.IsCompact {
		lookup := ds.NewLookupCode(mod.Packages)
		pkgFileCounts := dict.Entries(dict.Zip(pkgNames, fileCounts))
		slices.SortFunc(pkgFileCounts, sortDescCount)
		for _, e := range pkgFileCounts {
			pkgName, count := e.Tuple()
			pkg := lookup[pkgName]
			testCount := pkg.CountTestFiles()
			ratio := percentage(count, totalFileCount)
			if hasTest {
				codeCount := count - testCount
				out = append(out, fmt.Sprintf("\t%4s : %2d | %2d | %2d : %4s: %s", ratio, count, codeCount, testCount, pkg.Type, pkgName))
			} else {
				out = append(out, fmt.Sprintf("\t%4s : %2d : %4s: %s", ratio, count, pkg.Type, pkgName))
			}
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
	globalAvg := number.Ratio(totalLineCount, totalFileCount)
	out = append(out, fmt.Sprintf("TotalFiles: %d, TotalLines: %s, AvgLinePerFile: %.1f", totalFileCount, number.Comma(totalLineCount), globalAvg))

	testLineCount := mod.CountTestLines()
	if testLineCount > 0 {
		testFileCount := mod.CountTestFiles()
		codeLineCount := totalLineCount - testLineCount
		codeFileCount := totalFileCount - testFileCount
		codeRatio := percentage(codeLineCount, totalLineCount)
		testRatio := percentage(testLineCount, totalLineCount)
		codeALPF := number.Ratio(codeLineCount, codeFileCount)
		testALPF := number.Ratio(testLineCount, testFileCount)
		out = append(out, fmt.Sprintf("- CodeFiles: %d, CodeLines: %s (%s), CodeALPF: %.1f", codeFileCount, number.Comma(codeLineCount), codeRatio, codeALPF))
		out = append(out, fmt.Sprintf("- TestFiles: %d, TestLines: %s (%s), TestALPF: %.1f", testFileCount, number.Comma(testLineCount), testRatio, testALPF))
	}

	if !mod.IsCompact {
		// Sort packages by total lines
		lookup := ds.NewLookupCode(mod.Packages)
		pkgNames := mod.PackageNames()
		lineCounts := list.Map(mod.Packages, (*Package).CountLines)
		pkgLineCounts := dict.Entries(dict.Zip(pkgNames, lineCounts))
		slices.SortFunc(pkgLineCounts, sortDescCount)
		for _, e := range pkgLineCounts {
			pkgName, pkgCount := e.Tuple()
			pkg := lookup[pkgName]
			fileNames := pkg.FileNames()
			pkgRatio := percentage(pkgCount, totalLineCount)
			countStr := number.Comma(pkgCount) + " lines"
			alpf := number.Ratio(pkgCount, len(fileNames))
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
	}

	return strings.Join(out, "\n")
}

// Per package: list files x char count per file
func (mod StatsModule) PerPackageFilesCharCount() string {
	out := make([]string, 0)

	// Total chars, average char per file, average char per line
	totalCharCount := mod.CountChars()
	totalFileCount := mod.CountFiles()
	totalLineCount := mod.CountLines()
	charPerFile := number.Ratio(totalCharCount, totalFileCount)
	charPerLine := number.Ratio(totalCharCount, totalLineCount)
	a, b := number.Comma(totalCharCount), number.FloatComma(charPerFile, 1)
	out = append(out, fmt.Sprintf("TotalChars: %s, AvgCharPerFile: %s, AvgCharPerLine: %.1f", a, b, charPerLine))

	testCharCount := mod.CountTestChars()
	if testCharCount > 0 {
		testFileCount := mod.CountTestFiles()
		testLineCount := mod.CountTestLines()
		codeFileCount := totalFileCount - testFileCount
		codeLineCount := totalLineCount - testLineCount
		codeCharCount := totalCharCount - testCharCount
		codeRatio := percentage(codeCharCount, totalCharCount)
		testRatio := percentage(testCharCount, totalCharCount)
		codeACPF := number.Ratio(codeCharCount, codeFileCount)
		codeACPL := number.Ratio(codeCharCount, codeLineCount)
		testACPF := number.Ratio(testCharCount, testFileCount)
		testACPL := number.Ratio(testCharCount, testLineCount)
		a, b, c := number.Comma(codeCharCount), codeRatio, number.FloatComma(codeACPF, 1)
		out = append(out, fmt.Sprintf("- CodeChars: %s (%s), CodeACPF: %s, CodeACPL: %.1f", a, b, c, codeACPL))
		a, b, c = number.Comma(testCharCount), testRatio, number.FloatComma(testACPF, 1)
		out = append(out, fmt.Sprintf("- TestChars: %s (%s), TestACPF: %s, TestACPL: %.1f", a, b, c, testACPL))
	}

	if !mod.IsCompact {
		// Sort packages by total chars
		lookup := ds.NewLookupCode(mod.Packages)
		pkgNames := mod.PackageNames()
		charCounts := list.Map(mod.Packages, (*Package).CountChars)
		pkgCharCounts := dict.Entries(dict.Zip(pkgNames, charCounts))
		slices.SortFunc(pkgCharCounts, sortDescCount)
		for _, e := range pkgCharCounts {
			pkgName, pkgCount := e.Tuple()
			pkg := lookup[pkgName]
			fileNames := pkg.FileNames()
			pkgRatio := percentage(pkgCount, totalCharCount)
			countStr := number.Comma(pkgCount) + " chars"
			acpf := number.FloatComma(number.Ratio(pkgCount, len(fileNames)), 1)
			acpl := number.Ratio(pkgCount, pkg.CountLines())
			out = append(out, fmt.Sprintf("\t%4s : %-12s : %s (%s) (%.1f)", pkgRatio, countStr, pkgName, acpf, acpl))
			// Sort files by char count
			charCounts = list.Map(pkg.Files, (*File).CountChars)
			fileCharCounts := dict.Entries(dict.Zip(fileNames, charCounts))
			slices.SortFunc(fileCharCounts, sortDescCount)
			for _, f := range fileCharCounts {
				fileName, count := f.Tuple()
				globalRatio := percentage(count, totalCharCount)
				localRatio := percentage(count, pkgCount)
				out = append(out, fmt.Sprintf("\t\t*%4s :%4s : %6s : %s", globalRatio, localRatio, number.Comma(count), fileName))
			}
		}
	}

	return strings.Join(out, "\n")
}
