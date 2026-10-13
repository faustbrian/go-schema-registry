//go:build ignore

package main

import (
	"strings"
	"testing"
)

func TestProfileAcceptsOnlyReviewedGap(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		for _, count := range []string{"0", "1"} {
			profile := "mode: atomic\n" + exceptedBlock + " 1 " + count + "\n" + coveragePackage + "/framing.go:199.2,199.40 1 1\n"
			total, err := verifyProfile([]byte(profile), arch)
			if err != nil || total != 2 {
				t.Fatalf("valid profile: total=%d err=%v", total, err)
			}
		}
	}
	duplicate := "mode: atomic\n" + exceptedBlock + " 1 0\n" + exceptedBlock + " 1 1\n"
	if total, err := verifyProfile([]byte(duplicate), "amd64"); err != nil || total != 1 {
		t.Fatalf("consistent duplicate: total=%d err=%v", total, err)
	}
}

func TestProfileRefusesOtherGapsAndInvalidEvidence(t *testing.T) {
	good := "mode: atomic\n" + exceptedBlock + " 1 0\n"
	profiles := []string{"", "mode: atomic\n", "mode: set\n", "mode: atomic\nmalformed\n",
		"mode: atomic\n" + exceptedBlock + " 0 0\n", "mode: atomic\n" + exceptedBlock + " 2 0\n",
		"mode: atomic\n" + exceptedBlock + " 1 -1\n", "mode: atomic\n" + exceptedBlock + " x 1\n",
		"mode: atomic\nother/package/framing.go:197.3,198.1 1 0\n",
		"mode: atomic\n" + coveragePackage + "/framing.go:197.3,198.2 1 0\n",
		good + coveragePackage + "/request.go:1.1,2.1 1 0\n", good + exceptedBlock + " 2 1\n"}
	for _, profile := range profiles {
		if _, err := verifyProfile([]byte(profile), "amd64"); err == nil {
			t.Fatalf("invalid evidence accepted: %q", profile)
		}
	}
	for _, arch := range []string{"386", "arm", "unknown"} {
		if _, err := verifyProfile([]byte(good), arch); err == nil {
			t.Fatalf("exception accepted on %s", arch)
		}
	}
}

func TestBindingExpiresOnMaterialNotProseChanges(t *testing.T) {
	before, err := goMaterial([]byte("package p\nfunc value() int { return 1 }\n"))
	if err != nil {
		t.Fatal(err)
	}
	prose, err := goMaterial([]byte("package p\n// Different prose.\nfunc value() int { return 1 }\n"))
	if err != nil {
		t.Fatal(err)
	}
	changed, err := goMaterial([]byte("package p\nfunc value() int { return 2 }\n"))
	if err != nil {
		t.Fatal(err)
	}
	if digest(before) != digest(prose) || digest(before) == digest(changed) {
		t.Fatal("material/prose binding distinction changed")
	}
	directives, err := goMaterial([]byte("//go:build amd64\npackage p\nfunc value() int { return 1 }\n"))
	if err != nil {
		t.Fatal(err)
	}
	if digest(before) == digest(directives) {
		t.Fatal("build directive omitted from binding")
	}
	inputs := map[string]string{"source": digest(before)}
	binding := disposition{Package: coveragePackage, Coordinate: exceptedBlock, Statements: 1, GoVersion: "go1.27.2", Architectures: []string{"amd64", "arm64"}, Inputs: inputs}
	if err := validateBinding(binding, map[string]string{"source": digest(prose)}); err != nil {
		t.Fatal(err)
	}
	for _, input := range []map[string]string{{"source": digest(changed)}, {}, {"source": digest(before), "new-source": "new"}} {
		if err := validateBinding(binding, input); err == nil || !strings.Contains(err.Error(), "expired") {
			t.Fatal("changed/missing/added input did not expire proof")
		}
	}
	binding.Statements = 2
	if err := validateBinding(binding, inputs); err == nil {
		t.Fatal("widened exception accepted")
	}
}
