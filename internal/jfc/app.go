package jfc

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

const (
	exitSuccess = 0
	exitDiff    = 1
	exitError   = 2
)

const initConfigContents = `ignore = ["dist", "vendor", "node_modules", "*.generated.*"]
`

var Version = "dev"

type runMode int

const (
	modePrint runMode = iota
	modeWrite
	modeCheck
	modeListDifferent
	modeDiff
)

func Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, getwd func() (string, error)) int {
	if len(args) > 0 && args[0] == "init" {
		return runInit(args[1:], stdout, stderr, getwd)
	}

	fs := flag.NewFlagSet("jfc", flag.ContinueOnError)
	fs.SetOutput(stderr)

	write := fs.Bool("write", false, "Edit files in place.")
	check := fs.Bool("check", false, "Check that files are formatted.")
	listDifferent := fs.Bool("list-different", false, "Print paths whose formatting differs.")
	diff := fs.Bool("diff", false, "Print formatting changes as a unified diff.")
	configPath := fs.String("config", "", "Path to a jfc.toml config file.")
	stdinFilepath := fs.String("stdin-filepath", "", "Treat stdin as if it came from this file path.")
	version := fs.Bool("version", false, "Print the jfc version.")

	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: jfc [--version]\n")
		fmt.Fprintf(stderr, "       jfc [--write|--check [--diff]|--list-different|--diff] [--config path] [--stdin-filepath path] [file ...]\n")
		fmt.Fprintf(stderr, "       jfc < file.json\n")
		fmt.Fprintf(stderr, "       jfc init\n")
		fmt.Fprintf(stderr, "Supported files: %s\n", supportedExtensionsText())
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitSuccess
		}
		return exitError
	}
	if *version {
		if fs.NArg() > 0 {
			fmt.Fprintln(stderr, "jfc: --version does not accept file arguments")
			return exitError
		}
		fmt.Fprintln(stdout, versionString())
		return exitSuccess
	}

	mode, err := resolveMode(*write, *check, *listDifferent, *diff)
	if err != nil {
		fmt.Fprintln(stderr, "jfc:", err)
		return exitError
	}

	cwd, err := getwd()
	if err != nil {
		fmt.Fprintln(stderr, "jfc: determine working directory:", err)
		return exitError
	}

	loader := newConfigLoader(*configPath)
	standardIgnores := newStandardIgnoreLoader()
	paths := fs.Args()
	if len(paths) == 0 || (len(paths) == 1 && paths[0] == "-") {
		return runStdin(mode, stdin, stdout, stderr, loader, cwd, *stdinFilepath)
	}
	for _, path := range paths {
		if path == "-" {
			fmt.Fprintln(stderr, "jfc: '-' cannot be combined with file paths")
			return exitError
		}
	}

	targets, err := collectTargetsWithIgnores(paths, loader, standardIgnores)
	if err != nil {
		fmt.Fprintln(stderr, "jfc:", err)
		return exitError
	}
	if len(targets) == 0 {
		fmt.Fprintln(stderr, "jfc: no supported files found")
		return exitError
	}
	if mode == modePrint && len(targets) > 1 {
		fmt.Fprintln(stderr, "jfc: multiple file arguments require --write, --check, --diff, or --list-different")
		return exitError
	}

	hadError := false
	hadDiff := false

	for _, path := range targets {
		cfg, err := loader.forFile(path)
		if err != nil {
			fmt.Fprintf(stderr, "jfc: %s\n", err)
			hadError = true
			continue
		}
		ignored, err := cfg.ignores(path)
		if err != nil {
			fmt.Fprintf(stderr, "jfc: %s\n", err)
			hadError = true
			continue
		}
		if ignored {
			continue
		}
		ignored, err = standardIgnores.ignores(path)
		if err != nil {
			fmt.Fprintf(stderr, "jfc: %s\n", err)
			hadError = true
			continue
		}
		if ignored {
			continue
		}
		format, ok := detectFormat(path)
		if !ok {
			fmt.Fprintf(stderr, "jfc: %s is not a supported file (%s)\n", path, supportedExtensionsText())
			hadError = true
			continue
		}

		input, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(stderr, "jfc: read %s: %v\n", path, err)
			hadError = true
			continue
		}

		output, err := formatDocument(input, format, cfg)
		if err != nil {
			fmt.Fprintf(stderr, "jfc: %s: %v\n", path, err)
			hadError = true
			continue
		}

		changed := !bytes.Equal(input, output)
		switch mode {
		case modePrint:
			if _, err := stdout.Write(output); err != nil {
				fmt.Fprintf(stderr, "jfc: write stdout: %v\n", err)
				return exitError
			}
		case modeWrite:
			if changed {
				if err := writeFileAtomically(path, output); err != nil {
					fmt.Fprintf(stderr, "jfc: write %s: %v\n", path, err)
					hadError = true
					continue
				}
				fmt.Fprintln(stdout, path)
			}
		case modeCheck:
			if changed {
				hadDiff = true
				fmt.Fprintln(stdout, path)
			}
		case modeListDifferent:
			if changed {
				hadDiff = true
				fmt.Fprintln(stdout, path)
			}
		case modeDiff:
			if changed {
				hadDiff = true
				if _, err := stdout.Write([]byte(unifiedDiff(path, path, input, output))); err != nil {
					fmt.Fprintf(stderr, "jfc: write stdout: %v\n", err)
					return exitError
				}
			}
		}
	}

	if hadError {
		return exitError
	}
	if hadDiff {
		return exitDiff
	}
	return exitSuccess
}

func versionString() string {
	if Version != "" {
		return Version
	}
	return "dev"
}

func runInit(args []string, stdout io.Writer, stderr io.Writer, getwd func() (string, error)) int {
	if len(args) > 0 {
		if len(args) == 1 && args[0] == "--help" {
			fmt.Fprintln(stderr, "Usage: jfc init")
			fmt.Fprintln(stderr, "Create a minimal jfc.toml in the current directory.")
			return exitSuccess
		}
		fmt.Fprintln(stderr, "jfc: init does not accept arguments")
		return exitError
	}

	cwd, err := getwd()
	if err != nil {
		fmt.Fprintln(stderr, "jfc: determine working directory:", err)
		return exitError
	}

	path := filepath.Join(cwd, defaultConfigName)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			fmt.Fprintf(stderr, "jfc: %s already exists\n", path)
			return exitError
		}
		fmt.Fprintf(stderr, "jfc: create %s: %v\n", path, err)
		return exitError
	}

	if _, err := file.WriteString(initConfigContents); err != nil {
		file.Close()
		_ = os.Remove(path)
		fmt.Fprintf(stderr, "jfc: write %s: %v\n", path, err)
		return exitError
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		fmt.Fprintf(stderr, "jfc: write %s: %v\n", path, err)
		return exitError
	}

	fmt.Fprintln(stdout, path)
	return exitSuccess
}

func resolveMode(write bool, check bool, listDifferent bool, diff bool) (runMode, error) {
	selected := 0
	mode := modePrint

	if write {
		selected++
		mode = modeWrite
	}
	if check {
		selected++
		mode = modeCheck
	}
	if listDifferent {
		selected++
		mode = modeListDifferent
	}
	if selected > 1 {
		return modePrint, fmt.Errorf("--write, --check, and --list-different are mutually exclusive")
	}
	if diff && (write || listDifferent) {
		return modePrint, fmt.Errorf("--diff can be used alone or with --check")
	}
	if diff {
		mode = modeDiff
	}
	return mode, nil
}

func runStdin(mode runMode, stdin io.Reader, stdout io.Writer, stderr io.Writer, loader *configLoader, cwd string, stdinFilepath string) int {
	if mode == modeWrite {
		fmt.Fprintln(stderr, "jfc: --write cannot be used with stdin")
		return exitError
	}

	cfg, err := loader.forStdin(cwd, stdinFilepath)
	if err != nil {
		fmt.Fprintf(stderr, "jfc: %s\n", err)
		return exitError
	}

	input, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintln(stderr, "jfc: read stdin:", err)
		return exitError
	}

	format := FormatJSON
	if stdinFilepath != "" {
		detected, ok := detectFormat(stdinFilepath)
		if !ok {
			fmt.Fprintf(stderr, "jfc: %s is not a supported file (%s)\n", stdinFilepath, supportedExtensionsText())
			return exitError
		}
		format = detected
	}

	output, err := formatDocument(input, format, cfg)
	if err != nil {
		name := "stdin"
		if stdinFilepath != "" {
			name = stdinFilepath
		}
		fmt.Fprintf(stderr, "jfc: %s: %v\n", name, err)
		return exitError
	}

	changed := !bytes.Equal(input, output)
	switch mode {
	case modePrint:
		if _, err := stdout.Write(output); err != nil {
			fmt.Fprintln(stderr, "jfc: write stdout:", err)
			return exitError
		}
		return exitSuccess
	case modeCheck, modeListDifferent:
		if changed {
			label := "stdin"
			if stdinFilepath != "" {
				label = stdinFilepath
			}
			fmt.Fprintln(stdout, label)
			return exitDiff
		}
		return exitSuccess
	case modeDiff:
		if changed {
			label := "stdin"
			if stdinFilepath != "" {
				label = stdinFilepath
			}
			if _, err := stdout.Write([]byte(unifiedDiff(label, label, input, output))); err != nil {
				fmt.Fprintln(stderr, "jfc: write stdout:", err)
				return exitError
			}
			return exitDiff
		}
		return exitSuccess
	default:
		return exitError
	}
}

func collectTargets(args []string) ([]string, error) {
	return collectTargetsWithIgnores(args, nil, nil)
}

func collectTargetsWithIgnores(args []string, loader *configLoader, standardIgnores *standardIgnoreLoader) ([]string, error) {
	seen := make(map[string]struct{})
	var targets []string

	for _, arg := range args {
		expanded, err := expandArgWithIgnores(arg, loader, standardIgnores)
		if err != nil {
			return nil, err
		}

		for _, path := range expanded {
			if err := appendTarget(path, seen, &targets, loader, standardIgnores); err != nil {
				return nil, err
			}
		}
	}

	slices.Sort(targets)
	return targets, nil
}

func expandArg(arg string) ([]string, error) {
	return expandArgWithIgnores(arg, nil, nil)
}

func expandArgWithIgnores(arg string, loader *configLoader, standardIgnores *standardIgnoreLoader) ([]string, error) {
	if !hasGlob(arg) {
		return []string{arg}, nil
	}
	if hasRecursiveGlob(arg) {
		return expandRecursiveGlobWithIgnores(arg, loader, standardIgnores)
	}

	matches, err := filepath.Glob(arg)
	if err != nil {
		return nil, fmt.Errorf("invalid glob %q: %w", arg, err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("glob %q did not match any files", arg)
	}
	return matches, nil
}

func hasRecursiveGlob(pattern string) bool {
	for _, segment := range strings.Split(filepath.ToSlash(pattern), "/") {
		if segment == "**" {
			return true
		}
	}
	return false
}

func expandRecursiveGlob(pattern string) ([]string, error) {
	return expandRecursiveGlobWithIgnores(pattern, nil, nil)
}

func expandRecursiveGlobWithIgnores(pattern string, loader *configLoader, standardIgnores *standardIgnoreLoader) ([]string, error) {
	root := recursiveGlobRoot(pattern)
	if _, err := os.Stat(root); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("glob %q did not match any files", pattern)
		}
		return nil, fmt.Errorf("stat glob root %s: %w", root, err)
	}

	var matches []string
	err := filepath.WalkDir(root, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() && entry.Name() == ".git" {
			return filepath.SkipDir
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if entry.IsDir() && filepath.Clean(current) != filepath.Clean(root) {
			ignored, err := ignoredPath(current, true, loader, standardIgnores)
			if err != nil {
				return err
			}
			if ignored {
				return filepath.SkipDir
			}
		}
		if !entry.IsDir() {
			ignored, err := ignoredPath(current, false, loader, standardIgnores)
			if err != nil {
				return err
			}
			if ignored {
				return nil
			}
		}
		matched, err := matchRecursiveGlob(pattern, current)
		if err != nil {
			return err
		}
		if matched {
			matches = append(matches, current)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, path.ErrBadPattern) {
			return nil, fmt.Errorf("invalid glob %q: %w", pattern, err)
		}
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("glob %q did not match any files", pattern)
	}
	return matches, nil
}

func recursiveGlobRoot(pattern string) string {
	clean := filepath.Clean(pattern)
	volume := filepath.VolumeName(clean)
	rest := strings.TrimPrefix(clean, volume)
	rooted := strings.HasPrefix(rest, string(filepath.Separator))
	segments := strings.Split(filepath.ToSlash(rest), "/")
	fixed := make([]string, 0, len(segments))

	for _, segment := range segments {
		if segment == "" {
			continue
		}
		if segment == "**" || hasGlob(segment) {
			break
		}
		fixed = append(fixed, segment)
	}

	if len(fixed) == 0 {
		if rooted {
			return volume + string(filepath.Separator)
		}
		if volume != "" {
			return volume
		}
		return "."
	}

	joined := filepath.FromSlash(strings.Join(fixed, "/"))
	if rooted {
		return filepath.Join(volume+string(filepath.Separator), joined)
	}
	return volume + joined
}

func matchRecursiveGlob(pattern string, candidate string) (bool, error) {
	patternSegments := pathSegments(pattern)
	candidateSegments := pathSegments(candidate)
	return matchGlobSegments(patternSegments, candidateSegments)
}

func pathSegments(value string) []string {
	clean := filepath.Clean(value)
	return strings.Split(filepath.ToSlash(clean), "/")
}

func matchGlobSegments(patternSegments []string, candidateSegments []string) (bool, error) {
	if len(patternSegments) == 0 {
		return len(candidateSegments) == 0, nil
	}

	if patternSegments[0] == "**" {
		if matched, err := matchGlobSegments(patternSegments[1:], candidateSegments); matched || err != nil {
			return matched, err
		}
		if len(candidateSegments) == 0 {
			return false, nil
		}
		return matchGlobSegments(patternSegments, candidateSegments[1:])
	}

	if len(candidateSegments) == 0 {
		return false, nil
	}
	matched, err := path.Match(patternSegments[0], candidateSegments[0])
	if err != nil {
		return false, err
	}
	if !matched {
		return false, nil
	}
	return matchGlobSegments(patternSegments[1:], candidateSegments[1:])
}

func appendTarget(path string, seen map[string]struct{}, targets *[]string, loader *configLoader, standardIgnores *standardIgnoreLoader) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		targetInfo, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("stat %s: %w", path, err)
		}
		if targetInfo.IsDir() {
			return nil
		}
		if _, ok := detectFormat(path); !ok {
			return fmt.Errorf("%s is not a supported file (%s)", path, supportedExtensionsText())
		}
		addUnique(path, seen, targets)
		return nil
	}

	if info.IsDir() {
		return filepath.WalkDir(path, func(current string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() && entry.Name() == ".git" {
				return filepath.SkipDir
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return nil
			}
			if entry.IsDir() && filepath.Clean(current) != filepath.Clean(path) {
				ignored, err := ignoredPath(current, true, loader, standardIgnores)
				if err != nil {
					return err
				}
				if ignored {
					return filepath.SkipDir
				}
			}
			if entry.IsDir() {
				return nil
			}
			ignored, err := ignoredPath(current, false, loader, standardIgnores)
			if err != nil {
				return err
			}
			if ignored {
				return nil
			}
			if _, ok := detectFormat(entry.Name()); !ok {
				return nil
			}
			addUnique(current, seen, targets)
			return nil
		})
	}

	if _, ok := detectFormat(path); !ok {
		return fmt.Errorf("%s is not a supported file (%s)", path, supportedExtensionsText())
	}

	addUnique(path, seen, targets)
	return nil
}

func ignoredPath(path string, isDir bool, loader *configLoader, standardIgnores *standardIgnoreLoader) (bool, error) {
	if loader != nil {
		cfg, err := loader.forFile(path)
		if err != nil {
			return false, err
		}
		ignored, err := cfg.ignores(path)
		if err != nil || ignored {
			return ignored, err
		}
	}

	if standardIgnores != nil {
		return standardIgnores.ignoresPath(path, isDir)
	}
	return false, nil
}

func addUnique(path string, seen map[string]struct{}, targets *[]string) {
	clean := filepath.Clean(path)
	if _, ok := seen[clean]; ok {
		return
	}
	seen[clean] = struct{}{}
	*targets = append(*targets, clean)
}

func hasGlob(path string) bool {
	return strings.ContainsAny(path, "*?[")
}

func writeFileAtomically(path string, contents []byte) error {
	writePath := path
	if info, err := os.Lstat(path); err != nil {
		return err
	} else if info.Mode()&os.ModeSymlink != 0 {
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			return err
		}
		writePath = resolved
	}

	info, err := os.Stat(writePath)
	if err != nil {
		return err
	}

	dir := filepath.Dir(writePath)
	tempFile, err := os.CreateTemp(dir, "."+filepath.Base(writePath)+".tmp-*")
	if err != nil {
		return err
	}

	tempPath := tempFile.Name()
	cleanup := func() {
		_ = os.Remove(tempPath)
	}

	if _, err := tempFile.Write(contents); err != nil {
		tempFile.Close()
		cleanup()
		return err
	}
	if err := tempFile.Close(); err != nil {
		cleanup()
		return err
	}
	if err := os.Chmod(tempPath, info.Mode().Perm()); err != nil {
		cleanup()
		return err
	}
	if err := os.Rename(tempPath, writePath); err != nil {
		cleanup()
		return err
	}
	return nil
}
