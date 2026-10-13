//go:build ignore

// Package main enforces Confluent provider-owned native-64 coverage.
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/scanner"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const coveragePackage = "github.com/faustbrian/go-schema-registry/providers/confluent/v3"
const exceptedBlock = coveragePackage + "/framing.go:197.3,198.1"

type disposition struct {
	Package       string            `json:"package"`
	Coordinate    string            `json:"coordinate"`
	Statements    int               `json:"statements"`
	GoVersion     string            `json:"go_version"`
	Architectures []string          `json:"architectures"`
	Inputs        map[string]string `json:"inputs"`
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// Ignore prose/formatting, but retain build and compiler directives.
func goMaterial(data []byte) ([]byte, error) {
	var scan scanner.Scanner
	var scanErr error
	scan.Init(token.NewFileSet().AddFile("source.go", -1, len(data)), data,
		func(_ token.Position, message string) { scanErr = errors.New(message) }, scanner.ScanComments)
	var result strings.Builder
	for {
		_, kind, literal := scan.Scan()
		if kind == token.EOF {
			break
		}
		if kind == token.COMMENT && !strings.HasPrefix(literal, "//go:") && !strings.HasPrefix(literal, "// +build") {
			continue
		}
		fmt.Fprintf(&result, "%s:%q;", kind, literal)
	}
	return []byte(result.String()), scanErr
}

func textMaterial(data []byte, comments bool) []byte {
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		if comments {
			line = strings.SplitN(line, " #", 2)[0]
			if strings.HasPrefix(strings.TrimSpace(line), "#") {
				continue
			}
		}
		if value := strings.Join(strings.Fields(line), " "); value != "" {
			lines = append(lines, value)
		}
	}
	return []byte(strings.Join(lines, "\n"))
}

func materialInputs(root string, policy []byte) (map[string]string, error) {
	module := filepath.Join(root, "providers/confluent")
	var files []string
	err := filepath.WalkDir(module, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (entry.Name() == ".verification" || entry.Name() == ".git") {
			return filepath.SkipDir
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, name := range []string{"go.mod", "go.sum", "verification/package.mk", "scripts/check-coverage-exception.go", "scripts/check-coverage-exception_test.go", "scripts/check-coverage-exception.sh"} {
		files = append(files, filepath.Join(module, name))
	}
	inputs := make(map[string]string)
	for _, path := range files {
		// #nosec G304 -- Paths are enumerated only beneath the trusted owning repository, not selected by profile or network input.
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(path, ".go") {
			data, err = goMaterial(data)
			if err != nil {
				return nil, err
			}
		} else {
			if filepath.Base(path) == "go.mod" {
				data = []byte(regexp.MustCompile(`(?m)//[^\n]*`).ReplaceAllString(string(data), ""))
			}
			data = textMaterial(data, strings.HasSuffix(path, ".mk") || strings.HasSuffix(path, ".sh"))
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return nil, err
		}
		inputs[filepath.ToSlash(relative)] = digest(data)
	}
	data := policy
	var config map[string]json.RawMessage
	if err = json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	selected := map[string]any{"tool_version": config["tool_version"], "tool_checksums_sha256": config["tool_checksums_sha256"]}
	var operations []map[string]any
	if err = json.Unmarshal(config["operations"], &operations); err != nil {
		return nil, err
	}
	var ownOperations []map[string]any
	for _, operation := range operations {
		if operation["module"] == "providers/confluent" {
			ownOperations = append(ownOperations, operation)
		}
	}
	selected["operations"] = ownOperations
	var coverage struct {
		Modules []map[string]any `json:"modules"`
	}
	if err = json.Unmarshal(config["coverage"], &coverage); err != nil {
		return nil, err
	}
	for _, policy := range coverage.Modules {
		if policy["module"] == "providers/confluent" {
			selected["coverage"] = policy
		}
	}
	data, err = json.Marshal(selected)
	if err != nil {
		return nil, err
	}
	inputs["native-confluent-policy"] = digest(data)
	// #nosec G304 -- root is the owning repository working directory and the inventory filename is fixed.
	data, err = os.ReadFile(filepath.Join(root, "modules.json"))
	if err != nil {
		return nil, err
	}
	var catalog struct {
		Modules []map[string]json.RawMessage `json:"modules"`
	}
	if err = json.Unmarshal(data, &catalog); err != nil {
		return nil, err
	}
	for _, module := range catalog.Modules {
		var directory string
		if err = json.Unmarshal(module["directory"], &directory); err != nil {
			return nil, err
		}
		if directory == "providers/confluent" {
			policy := make(map[string]json.RawMessage)
			for _, key := range []string{"module_path", "go_version", "packages", "gates", "test_tags", "build_tags", "required_services"} {
				policy[key] = module[key]
			}
			data, err = json.Marshal(policy)
			if err != nil {
				return nil, err
			}
			inputs["confluent-inventory"] = digest(data)
		}
	}
	if _, exists := inputs["confluent-inventory"]; !exists {
		return nil, errors.New("missing Confluent inventory")
	}
	// #nosec G304 -- Read the fixed required-proof workflow under the trusted repository root.
	data, err = os.ReadFile(filepath.Join(root, ".github/workflows/ci.yml"))
	if err != nil {
		return nil, err
	}
	_, job, exists := strings.Cut(string(data), "  framing-386:\n")
	if !exists {
		return nil, errors.New("missing required 386 proof")
	}
	inputs["required-386-policy"] = digest(textMaterial([]byte(job), true))
	return inputs, nil
}

func validateBinding(binding disposition, inputs map[string]string) error {
	if binding.Package != coveragePackage || binding.Coordinate != exceptedBlock || binding.Statements != 1 || binding.GoVersion != "go1.27.2" || strings.Join(binding.Architectures, ",") != "amd64,arm64" {
		return errors.New("coverage exception scope/compiler changed")
	}
	expected, err := json.Marshal(binding.Inputs)
	if err != nil {
		return err
	}
	actual, err := json.Marshal(inputs)
	if err != nil {
		return err
	}
	if !bytes.Equal(expected, actual) {
		return errors.New("coverage exception expired: material inputs changed")
	}
	return nil
}

func verifyProfile(data []byte, architecture string) (int, error) {
	if architecture != "amd64" && architecture != "arm64" {
		return 0, errors.New("native-64 exception does not apply to target")
	}
	scan := bufio.NewScanner(bytes.NewReader(data))
	scan.Buffer(make([]byte, 4096), 1<<20)
	if !scan.Scan() || scan.Text() != "mode: atomic" {
		return 0, errors.New("missing atomic profile")
	}
	digits := regexp.MustCompile(`^[0-9]+$`)
	coordinate := regexp.MustCompile(`^` + regexp.QuoteMeta(coveragePackage) + `/[A-Za-z0-9_]+\.go:[0-9]+\.[0-9]+,[0-9]+\.[0-9]+$`)
	type block struct {
		statements int
		covered    bool
	}
	blocks := make(map[string]block)
	for scan.Scan() {
		fields := strings.Fields(scan.Text())
		if len(fields) != 3 || !coordinate.MatchString(fields[0]) || !digits.MatchString(fields[1]) || !digits.MatchString(fields[2]) {
			return 0, errors.New("invalid profile coordinate/package")
		}
		statements, err := strconv.Atoi(fields[1])
		if err != nil || statements < 0 {
			return 0, errors.New("invalid statement count")
		}
		count, err := strconv.ParseUint(fields[2], 10, 64)
		if err != nil {
			return 0, errors.New("invalid execution count")
		}
		prior, exists := blocks[fields[0]]
		if exists && prior.statements != statements {
			return 0, errors.New("inconsistent duplicate block")
		}
		if fields[0] == exceptedBlock && statements != 1 {
			return 0, errors.New("excepted statement count changed")
		}
		blocks[fields[0]] = block{statements, prior.covered || count > 0}
	}
	if err := scan.Err(); err != nil {
		return 0, err
	}
	total := 0
	for location, value := range blocks {
		if value.statements > int(^uint(0)>>1)-total {
			return 0, errors.New("statement count overflow")
		}
		total += value.statements
		if value.statements > 0 && !value.covered && location != exceptedBlock {
			return 0, errors.New("uncovered non-excepted production block: " + location)
		}
	}
	if total == 0 {
		return 0, errors.New("missing executable package coverage")
	}
	return total, nil
}

func run() error {
	policy := flag.String("policy", "", "fresh configured native policy JSON")
	profile := flag.String("profile", "", "fresh compiler profile; omission checks inputs only")
	architecture := flag.String("arch", "", "actual Go target from owning collector")
	version := flag.String("version", "", "actual compiler version from owning collector")
	emit := flag.Bool("inputs", false, "print material inputs for explicit reviewed renewal")
	flag.Parse()
	if *version != "go1.27.2" || (*architecture != "amd64" && *architecture != "arm64") {
		return errors.New("native-64 exception compiler/target changed")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		return err
	}
	data, err := os.ReadFile(*policy)
	if err != nil {
		return err
	}
	inputs, err := materialInputs(root, data)
	if err != nil {
		return err
	}
	if *emit {
		return json.NewEncoder(os.Stdout).Encode(inputs)
	}
	data, err = os.ReadFile("verification/coverage-exception.json")
	if err != nil {
		return err
	}
	var binding disposition
	if err = json.Unmarshal(data, &binding); err != nil {
		return err
	}
	if err = validateBinding(binding, inputs); err != nil {
		return err
	}
	if *profile == "" {
		return nil
	}
	data, err = os.ReadFile(*profile)
	if err != nil {
		return err
	}
	total, err := verifyProfile(data, *architecture)
	if err != nil {
		return err
	}
	fmt.Printf("Confluent native-64 coverage: %d statements checked; only reviewed 32-bit refusal may be uncovered\n", total)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
