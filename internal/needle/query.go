package needle

import (
	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/list"
)

// Count module total file count from packages
func (mod StatsModule) CountFiles() int {
	return list.Sum(list.Map(mod.Packages, func(pkg *Package) int {
		return len(pkg.Files)
	}))
}

// Count module total file count from packages
func (mod CodeModule) CountFiles() int {
	return list.Sum(list.Map(mod.Packages, func(pkg *Package) int {
		return len(pkg.Files)
	}))
}

// Count module total test file count from packages
func (mod StatsModule) CountTestFiles() int {
	return list.Sum(list.Map(mod.Packages, (*Package).CountTestFiles))
}

// Count package test files
func (pkg *Package) CountTestFiles() int {
	return len(pkg.TestFiles())
}

// Get package test files
func (pkg *Package) TestFiles() []*File {
	return list.Filter(pkg.Files, func(f *File) bool {
		return f.Type == FILE_TEST
	})
}

// Count module total line count from packages
func (mod StatsModule) CountLines() int {
	return list.Sum(list.Map(mod.Packages, (*Package).CountLines))
}

// Count module total line count from packages
func (mod CodeModule) CountLines() int {
	return list.Sum(list.Map(mod.Packages, (*Package).CountLines))
}

// Count module total line count from package test files
func (mod StatsModule) CountTestLines() int {
	return list.Sum(list.Map(mod.Packages, (*Package).CountTestLines))
}

// Count package total line count from files
func (pkg *Package) CountLines() int {
	return list.Sum(list.Map(pkg.Files, (*File).CountLines))
}

// Count package total line count from test files
func (pkg *Package) CountTestLines() int {
	return list.Sum(list.Map(pkg.TestFiles(), (*File).CountLines))
}

// Return line count for File
func (f *File) CountLines() int {
	return len(f.Lines)
}

// Return package names from Module
func (mod StatsModule) PackageNames() []string {
	return list.Map(mod.Packages, (*Package).GetCode)
}

// Return package names from Module
func (mod CodeModule) PackageNames() []string {
	return list.Map(mod.Packages, (*Package).GetCode)
}

// Return file names from Package
func (pkg *Package) FileNames() []string {
	return list.Map(pkg.Files, func(f *File) string {
		return f.Name
	})
}

// Count library packages in Module
func (mod StatsModule) CountLibPackages() int {
	return len(list.Filter(mod.Packages, func(pkg *Package) bool {
		return pkg.Type == PKG_LIB
	}))
}

// Count main packages in Module
func (mod StatsModule) CountMainPackages() int {
	return len(list.Filter(mod.Packages, func(pkg *Package) bool {
		return pkg.Type == PKG_MAIN
	}))
}

// Count total chars in module packages
func (mod StatsModule) CountChars() int {
	return list.Sum(list.Map(mod.Packages, (*Package).CountChars))
}

// Count total chars in module packages
func (mod CodeModule) CountChars() int {
	return list.Sum(list.Map(mod.Packages, (*Package).CountChars))
}

// Count module total chars from package test files
func (mod StatsModule) CountTestChars() int {
	return list.Sum(list.Map(mod.Packages, (*Package).CountTestChars))
}

// Count total chars in package files
func (pkg *Package) CountChars() int {
	return list.Sum(list.Map(pkg.Files, (*File).CountChars))
}

// Count package total chars from test files
func (pkg *Package) CountTestChars() int {
	return list.Sum(list.Map(pkg.TestFiles(), (*File).CountChars))
}

// Count total chars in file
func (f *File) CountChars() int {
	return list.Sum(list.Map(f.Lines, func(line *Line) int {
		return line.Length
	}))
}

// Compute module line usage from package files
func (mod CodeModule) BreakdownLines() map[LineType]int {
	return dict.MergeCounts(list.Map(mod.Packages, (*Package).BreakdownLines))
}

// Compute package line usage from files
func (pkg *Package) BreakdownLines() map[LineType]int {
	return dict.MergeCounts(list.Map(pkg.Files, (*File).BreakdownLines))
}

// Compute file line usage
func (f *File) BreakdownLines() map[LineType]int {
	return dict.CounterFunc(f.Lines, func(line *Line) LineType {
		return line.Type
	})
}

// Compute module char usage from package files
func (mod CodeModule) BreakdownChars() map[LineType]int {
	return dict.MergeCounts(list.Map(mod.Packages, (*Package).BreakdownChars))
}

// Compute package char usage from files
func (pkg *Package) BreakdownChars() map[LineType]int {
	return dict.MergeCounts(list.Map(pkg.Files, (*File).BreakdownChars))
}

// Compute file char usage
func (f *File) BreakdownChars() map[LineType]int {
	usage := make(map[LineType]int)
	for _, line := range f.Lines {
		usage[line.Type] += line.Length
	}
	return usage
}

// Compute module block breakdown
func (mod CodeModule) BreakdownBlocks() map[BlockType]int {
	return dict.MergeCounts(list.Map(mod.Packages, (*Package).BreakdownBlocks))
}

// Compute package block breakdown
func (pkg *Package) BreakdownBlocks() map[BlockType]int {
	return dict.MergeCounts(list.Map(pkg.Files, func(f *File) map[BlockType]int {
		return f.Block
	}))
}

// Count total block type from module packages
func (mod CodeModule) CountBlocks(blockType BlockType) int {
	return list.Sum(list.Map(mod.Packages, func(pkg *Package) int {
		return pkg.CountBlocks(blockType)
	}))
}

// Count total block type from package files
func (pkg *Package) CountBlocks(blockType BlockType) int {
	return list.Sum(list.Map(pkg.Files, func(f *File) int {
		return f.Block[blockType]
	}))
}

// Compute module code breakdown
func (mod CodeModule) BreakdownCodes() map[CodeType]int {
	return dict.MergeCounts(list.Map(mod.Packages, (*Package).BreakdownCodes))
}

// Compute package code breakdown
func (pkg *Package) BreakdownCodes() map[CodeType]int {
	return dict.MergeCounts(list.Map(pkg.Files, func(f *File) map[CodeType]int {
		return f.Code
	}))
}

// Count total code type from module packages
func (mod CodeModule) CountTypes(codeType CodeType) int {
	return list.Sum(list.Map(mod.Packages, func(pkg *Package) int {
		return pkg.CountTypes(codeType)
	}))
}

// Count total code type from package files
func (pkg *Package) CountTypes(codeType CodeType) int {
	return list.Sum(list.Map(pkg.Files, func(f *File) int {
		return f.Code[codeType]
	}))
}
