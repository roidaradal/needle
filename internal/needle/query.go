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
