package needle

import (
	"github.com/roidaradal/fn/list"
)

// Count module total file count from packages
func (mod StatsModule) CountFiles() int {
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
	return list.Map(mod.Packages, func(pkg *Package) string {
		return pkg.Name
	})
}

// Return file names from Package
func (pkg *Package) FileNames() []string {
	return list.Map(pkg.Files, func(f *File) string {
		return f.Name
	})
}
