package jfc

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type standardIgnoreLoader struct {
	cache map[string][]standardIgnoreRule
}

type standardIgnoreRule struct {
	source        string
	baseDir       string
	pattern       string
	negated       bool
	rooted        bool
	directoryOnly bool
}

func newStandardIgnoreLoader() *standardIgnoreLoader {
	return &standardIgnoreLoader{
		cache: make(map[string][]standardIgnoreRule),
	}
}

func (l *standardIgnoreLoader) ignores(filePath string) (bool, error) {
	rules, err := l.rulesFor(filePath)
	if err != nil {
		return false, err
	}

	ignored := false
	for _, rule := range rules {
		matched, err := rule.matches(filePath, false)
		if err != nil {
			return false, err
		}
		if matched {
			ignored = !rule.negated
		}
	}
	return ignored, nil
}

func (l *standardIgnoreLoader) rulesFor(filePath string) ([]standardIgnoreRule, error) {
	dirs := ancestorDirs(filepath.Dir(filePath))
	var rules []standardIgnoreRule
	for _, dir := range dirs {
		dirRules, err := l.rulesForDir(dir)
		if err != nil {
			return nil, err
		}
		rules = append(rules, dirRules...)
	}
	return rules, nil
}

func (l *standardIgnoreLoader) rulesForDir(dir string) ([]standardIgnoreRule, error) {
	dir = filepath.Clean(dir)
	if rules, ok := l.cache[dir]; ok {
		return rules, nil
	}

	var rules []standardIgnoreRule
	for _, name := range []string{".ignore", ".gitignore"} {
		fileRules, err := readStandardIgnoreFile(filepath.Join(dir, name), dir)
		if err != nil {
			return nil, err
		}
		rules = append(rules, fileRules...)
	}

	gitExclude := filepath.Join(dir, ".git", "info", "exclude")
	fileRules, err := readStandardIgnoreFile(gitExclude, dir)
	if err != nil {
		return nil, err
	}
	rules = append(rules, fileRules...)

	l.cache[dir] = rules
	return rules, nil
}

func ancestorDirs(startDir string) []string {
	dir := filepath.Clean(startDir)
	var reversed []string
	for {
		reversed = append(reversed, dir)
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	dirs := make([]string, len(reversed))
	for i := range reversed {
		dirs[i] = reversed[len(reversed)-1-i]
	}
	return dirs
}

func readStandardIgnoreFile(filePath string, baseDir string) ([]standardIgnoreRule, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read ignore source %s: %w", filePath, err)
	}
	defer file.Close()

	var rules []standardIgnoreRule
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		rule, ok := parseStandardIgnoreRule(filePath, baseDir, scanner.Text())
		if !ok {
			continue
		}
		if err := validateStandardIgnoreRule(rule); err != nil {
			return nil, fmt.Errorf("%s:%d: %w", filePath, lineNumber, err)
		}
		rules = append(rules, rule)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read ignore source %s: %w", filePath, err)
	}
	return rules, nil
}

func parseStandardIgnoreRule(source string, baseDir string, line string) (standardIgnoreRule, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return standardIgnoreRule{}, false
	}
	if strings.HasPrefix(line, `\#`) || strings.HasPrefix(line, `\!`) {
		line = line[1:]
	} else if strings.HasPrefix(line, "#") {
		return standardIgnoreRule{}, false
	}

	rule := standardIgnoreRule{
		source:  source,
		baseDir: filepath.Clean(baseDir),
	}
	if strings.HasPrefix(line, "!") {
		rule.negated = true
		line = strings.TrimPrefix(line, "!")
	}
	if line == "" {
		return standardIgnoreRule{}, false
	}

	if strings.HasSuffix(line, "/") {
		rule.directoryOnly = true
		line = strings.TrimRight(line, "/")
	}
	if strings.HasPrefix(line, "/") {
		rule.rooted = true
		line = strings.TrimLeft(line, "/")
	}

	line = filepath.ToSlash(filepath.Clean(line))
	if line == "." {
		return standardIgnoreRule{}, false
	}
	rule.pattern = line
	return rule, true
}

func validateStandardIgnoreRule(rule standardIgnoreRule) error {
	for _, segment := range strings.Split(rule.pattern, "/") {
		if segment == "" || segment == "**" {
			continue
		}
		if _, err := path.Match(segment, ""); err != nil {
			return fmt.Errorf("invalid ignore pattern %q: %w", rule.pattern, err)
		}
	}
	return nil
}

func (r standardIgnoreRule) matches(filePath string, isDir bool) (bool, error) {
	rel, err := filepath.Rel(r.baseDir, filePath)
	if err != nil || rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return false, nil
	}
	rel = filepath.ToSlash(filepath.Clean(rel))

	if r.directoryOnly {
		return r.matchesDirectory(rel, isDir)
	}
	if r.rooted || strings.Contains(r.pattern, "/") {
		return matchRecursiveGlob(r.pattern, rel)
	}

	for _, segment := range pathSegments(rel) {
		matched, err := path.Match(r.pattern, segment)
		if err != nil {
			return false, fmt.Errorf("invalid ignore pattern %q from %s: %w", r.pattern, r.source, err)
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

func (r standardIgnoreRule) matchesDirectory(rel string, isDir bool) (bool, error) {
	if r.rooted || strings.Contains(r.pattern, "/") {
		matched, err := matchRecursiveGlob(r.pattern, rel)
		if err != nil || matched {
			return matched, err
		}
		for _, prefix := range directoryPrefixes(rel, isDir) {
			matched, err := matchRecursiveGlob(r.pattern, prefix)
			if err != nil || matched {
				return matched, err
			}
		}
		return false, nil
	}

	for _, segment := range directorySegments(rel, isDir) {
		matched, err := path.Match(r.pattern, segment)
		if err != nil {
			return false, fmt.Errorf("invalid ignore pattern %q from %s: %w", r.pattern, r.source, err)
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

func directorySegments(rel string, isDir bool) []string {
	segments := pathSegments(rel)
	if isDir {
		return segments
	}
	if len(segments) == 0 {
		return nil
	}
	return segments[:len(segments)-1]
}

func directoryPrefixes(rel string, isDir bool) []string {
	segments := directorySegments(rel, isDir)
	prefixes := make([]string, 0, len(segments))
	for i := range segments {
		prefixes = append(prefixes, strings.Join(segments[:i+1], "/"))
	}
	return prefixes
}
