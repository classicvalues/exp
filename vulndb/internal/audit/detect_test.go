// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package audit

import (
	"fmt"
	"reflect"
	"testing"

	"golang.org/x/tools/go/packages"
	"golang.org/x/vuln/osv"
)

func moduleVulnerabilitiesToString(mv ModuleVulnerabilities) string {
	var s string
	for _, m := range mv {
		s += fmt.Sprintf("mod: %v\n", m.mod)
		for _, v := range m.vulns {
			s += fmt.Sprintf("\t%v\n", v)
		}
	}
	return s
}

func TestFilterVulns(t *testing.T) {
	mv := ModuleVulnerabilities{
		{
			mod: &packages.Module{
				Path:    "example.mod/a",
				Version: "v1.0.0",
			},
			vulns: []*osv.Entry{
				{ID: "a", Affected: []osv.Affected{
					{Ranges: osv.Affects{{Type: osv.TypeSemver, Events: []osv.RangeEvent{{Fixed: "2.0.0"}}}}},
					{Ranges: osv.Affects{{Type: osv.TypeSemver, Events: []osv.RangeEvent{{Fixed: "1.0.0"}}}}}, // should be filtered out
				}},
				{ID: "b", Affected: []osv.Affected{{Ranges: osv.Affects{{Type: osv.TypeSemver, Events: []osv.RangeEvent{{Introduced: "1.0.1"}}}}, EcosystemSpecific: osv.EcosystemSpecific{GOOS: []string{"windows", "linux"}}}}},
				{ID: "c", Affected: []osv.Affected{{Ranges: osv.Affects{{Type: osv.TypeSemver, Events: []osv.RangeEvent{{Introduced: "1.0.1"}, {Fixed: "1.0.1"}}}}, EcosystemSpecific: osv.EcosystemSpecific{GOARCH: []string{"arm64", "amd64"}}}}},
				{ID: "d", Affected: []osv.Affected{{EcosystemSpecific: osv.EcosystemSpecific{GOOS: []string{"windows"}}}}},
			},
		},
		{
			mod: &packages.Module{
				Path:    "example.mod/b",
				Version: "v1.0.0",
			},
			vulns: []*osv.Entry{
				{ID: "e", Affected: []osv.Affected{{EcosystemSpecific: osv.EcosystemSpecific{GOARCH: []string{"arm64"}}}}},
				{ID: "f", Affected: []osv.Affected{{EcosystemSpecific: osv.EcosystemSpecific{GOOS: []string{"linux"}}}}},
				{ID: "g", Affected: []osv.Affected{{EcosystemSpecific: osv.EcosystemSpecific{GOARCH: []string{"amd64"}}, Ranges: osv.Affects{{Type: osv.TypeSemver, Events: []osv.RangeEvent{{Introduced: "0.0.1"}, {Fixed: "2.0.1"}}}}}}},
				{ID: "h", Affected: []osv.Affected{{EcosystemSpecific: osv.EcosystemSpecific{GOOS: []string{"windows"}, GOARCH: []string{"amd64"}}}}},
			},
		},
		{
			mod: &packages.Module{
				Path: "example.mod/c",
			},
			vulns: []*osv.Entry{
				{ID: "i", Affected: []osv.Affected{{EcosystemSpecific: osv.EcosystemSpecific{GOARCH: []string{"amd64"}}, Ranges: osv.Affects{{Type: osv.TypeSemver, Events: []osv.RangeEvent{{Introduced: "0.0.0"}}}}}}},
				{ID: "j", Affected: []osv.Affected{{EcosystemSpecific: osv.EcosystemSpecific{GOARCH: []string{"amd64"}}, Ranges: osv.Affects{{Type: osv.TypeSemver, Events: []osv.RangeEvent{{Fixed: "3.0.0"}}}}}}},
				{ID: "k"},
			},
		},
		{
			mod: &packages.Module{
				Path:    "example.mod/d",
				Version: "v1.2.0",
			},
			vulns: []*osv.Entry{
				{ID: "l", Affected: []osv.Affected{
					{EcosystemSpecific: osv.EcosystemSpecific{GOOS: []string{"windows"}}}, // should be filtered out
					{EcosystemSpecific: osv.EcosystemSpecific{GOOS: []string{"linux"}}},
				}},
			},
		},
	}

	expected := ModuleVulnerabilities{
		{
			mod: &packages.Module{
				Path:    "example.mod/a",
				Version: "v1.0.0",
			},
			vulns: []*osv.Entry{
				{ID: "a", Affected: []osv.Affected{{Ranges: osv.Affects{{Type: osv.TypeSemver, Events: []osv.RangeEvent{{Fixed: "2.0.0"}}}}}}},
				{ID: "c", Affected: []osv.Affected{{EcosystemSpecific: osv.EcosystemSpecific{GOARCH: []string{"arm64", "amd64"}}, Ranges: osv.Affects{{Type: osv.TypeSemver, Events: []osv.RangeEvent{{Introduced: "1.0.1"}, {Fixed: "1.0.1"}}}}}}},
			},
		},
		{
			mod: &packages.Module{
				Path:    "example.mod/b",
				Version: "v1.0.0",
			},
			vulns: []*osv.Entry{
				{ID: "f", Affected: []osv.Affected{{EcosystemSpecific: osv.EcosystemSpecific{GOOS: []string{"linux"}}}}},
				{ID: "g", Affected: []osv.Affected{{EcosystemSpecific: osv.EcosystemSpecific{GOARCH: []string{"amd64"}}, Ranges: osv.Affects{{Type: osv.TypeSemver, Events: []osv.RangeEvent{{Introduced: "0.0.1"}, {Fixed: "2.0.1"}}}}}}},
			},
		},
		{
			mod: &packages.Module{
				Path: "example.mod/c",
			},
		},
		{
			mod: &packages.Module{
				Path:    "example.mod/d",
				Version: "v1.2.0",
			},
			vulns: []*osv.Entry{
				{ID: "l", Affected: []osv.Affected{{EcosystemSpecific: osv.EcosystemSpecific{GOOS: []string{"linux"}}}}},
			},
		},
	}

	filtered := mv.Filter("linux", "amd64")
	if !reflect.DeepEqual(filtered, expected) {
		t.Fatalf("Filter returned unexpected results, got:\n%s\nwant:\n%s", moduleVulnerabilitiesToString(filtered), moduleVulnerabilitiesToString(expected))
	}
}

func vulnsToString(vulns []*osv.Entry) string {
	var s string
	for _, v := range vulns {
		s += fmt.Sprintf("\t%v\n", v)
	}
	return s
}

func TestVulnsForPackage(t *testing.T) {
	mv := ModuleVulnerabilities{
		{
			mod: &packages.Module{
				Path:    "example.mod/a",
				Version: "v1.0.0",
			},
			vulns: []*osv.Entry{
				{ID: "a", Affected: []osv.Affected{{Package: osv.Package{Name: "example.mod/a/b/c"}}}},
			},
		},
		{
			mod: &packages.Module{
				Path:    "example.mod/a/b",
				Version: "v1.0.0",
			},
			vulns: []*osv.Entry{
				{ID: "b", Affected: []osv.Affected{{Package: osv.Package{Name: "example.mod/a/b/c"}}}},
			},
		},
		{
			mod: &packages.Module{
				Path:    "example.mod/d",
				Version: "v0.0.1",
			},
			vulns: []*osv.Entry{
				{ID: "d", Affected: []osv.Affected{{Package: osv.Package{Name: "example.mod/d"}}}},
			},
		},
	}

	filtered := mv.VulnsForPackage("example.mod/a/b/c")
	expected := []*osv.Entry{
		{ID: "b", Affected: []osv.Affected{{Package: osv.Package{Name: "example.mod/a/b/c"}}}},
	}

	if !reflect.DeepEqual(filtered, expected) {
		t.Fatalf("VulnsForPackage returned unexpected results, got:\n%s\nwant:\n%s", vulnsToString(filtered), vulnsToString(expected))
	}
}

func TestVulnsForPackageReplaced(t *testing.T) {
	mv := ModuleVulnerabilities{
		{
			mod: &packages.Module{
				Path:    "example.mod/a",
				Version: "v1.0.0",
			},
			vulns: []*osv.Entry{
				{ID: "a", Affected: []osv.Affected{{Package: osv.Package{Name: "example.mod/a/b/c"}}}},
			},
		},
		{
			mod: &packages.Module{
				Path: "example.mod/a/b",
				Replace: &packages.Module{
					Path: "example.mod/b",
				},
				Version: "v1.0.0",
			},
			vulns: []*osv.Entry{
				{ID: "c", Affected: []osv.Affected{{Package: osv.Package{Name: "example.mod/b/c"}}}},
			},
		},
	}

	filtered := mv.VulnsForPackage("example.mod/a/b/c")
	expected := []*osv.Entry{
		{ID: "c", Affected: []osv.Affected{{Package: osv.Package{Name: "example.mod/b/c"}}}},
	}

	if !reflect.DeepEqual(filtered, expected) {
		t.Fatalf("VulnsForPackage returned unexpected results, got:\n%s\nwant:\n%s", vulnsToString(filtered), vulnsToString(expected))
	}
}

func TestVulnsForSymbol(t *testing.T) {
	mv := ModuleVulnerabilities{
		{
			mod: &packages.Module{
				Path:    "example.mod/a",
				Version: "v1.0.0",
			},
			vulns: []*osv.Entry{
				{ID: "a", Affected: []osv.Affected{{Package: osv.Package{Name: "example.mod/a/b/c"}}}},
			},
		},
		{
			mod: &packages.Module{
				Path:    "example.mod/a/b",
				Version: "v1.0.0",
			},
			vulns: []*osv.Entry{
				{ID: "b", Affected: []osv.Affected{{Package: osv.Package{Name: "example.mod/a/b/c"}, EcosystemSpecific: osv.EcosystemSpecific{Symbols: []string{"a"}}}}},
				{ID: "c", Affected: []osv.Affected{{Package: osv.Package{Name: "example.mod/a/b/c"}, EcosystemSpecific: osv.EcosystemSpecific{Symbols: []string{"b"}}}}},
			},
		},
	}

	filtered := mv.VulnsForSymbol("example.mod/a/b/c", "a")
	expected := []*osv.Entry{
		{ID: "b", Affected: []osv.Affected{{Package: osv.Package{Name: "example.mod/a/b/c"}, EcosystemSpecific: osv.EcosystemSpecific{Symbols: []string{"a"}}}}},
	}

	if !reflect.DeepEqual(filtered, expected) {
		t.Fatalf("VulnsForPackage returned unexpected results, got:\n%s\nwant:\n%s", vulnsToString(filtered), vulnsToString(expected))
	}
}
