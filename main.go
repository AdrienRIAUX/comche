package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/pelletier/go-toml"
)

// tagFound stores the tag, path, line number, and line content
type tagFound struct {
	tag        string
	path       string
	lineNumber int
	line       string
}

// fileResults stores the file name and the tags found in the file
type fileResults struct {
	fileName  string
	tagsFound []tagFound
}

// gatherPythonFiles returns a list of all Python files in the given directory
func gatherPythonFiles(dir string) ([]string, error) {
	var pythonFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".py" {
			pythonFiles = append(pythonFiles, path)
		}
		return nil
	})
	return pythonFiles, err
}

// parseFileForComments scans a file for lines containing special tags like #TODO, #FIXME, etc.
func parseFileForComments(path string, tagPatterns map[string]*regexp.Regexp) ([]tagFound, error) {
	results := []tagFound{}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %v", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 1
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			lineNumber++
			continue
		}
		for tag, pattern := range tagPatterns {
			if pattern.MatchString(line) {
				results = append(results, tagFound{tag, path, lineNumber, line})
				break
			}
		}
		lineNumber++
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", path, err)
	}
	return results, nil
}

// inspectResult prints the tags found in the files and exits if the number of tags found is greater than the fail threshold
func inspectResult(results []fileResults, fail int) {
	counter := 0
	for _, result := range results {
		for _, tag := range result.tagsFound {
			fmt.Printf("Found %s in %s at line %d: %s\n", tag.tag, tag.path, tag.lineNumber, tag.line)
			counter++
		}
	}
	if counter > fail {
		os.Exit(1)
	}
}

// readConfigFile reads the config file and prints the tags
func readConfigFile(configPath string) (map[string]*regexp.Regexp, int) {
	tree, err := toml.LoadFile(configPath)
	if err != nil {
		fmt.Println("Error reading config file")
		os.Exit(1)
	}
	confTags := tree.Get("tool.comche.flags").([]interface{})
	tagPatterns := make(map[string]*regexp.Regexp, len(confTags))

	for _, tag := range confTags {
		tagStr, ok := tag.(string)
		if !ok {
			fmt.Println("Invalid tag format")
			os.Exit(1)
		}

		pattern := fmt.Sprintf(`# ?%s`, regexp.QuoteMeta(tagStr))
		re, err := regexp.Compile(pattern)
		if err != nil {
			fmt.Printf("Error compiling regex for tag %s: %v\n", tagStr, err)
			os.Exit(1)
		}

		tagPatterns[tagStr] = re
	}

	fails := tree.Get("tool.comche.fail").(int64)

	return tagPatterns, int(fails)
}

func main() {
	dirPtr := flag.String("dir", ".", "the root directory to scan for Python files")
	tagsPtr := flag.String("tags", "TODO-BUG-FIXME", "dash-separated list of tags to search for")
	modePtr := flag.String("mode", "commit", "mode of operation: commit or root")
	failPtr := flag.Int("fail", 0, "fail over n tags found")
	configPtr := flag.String("config", "", "path to the configuration file")

	flag.Parse()

	root := *dirPtr
	tags := strings.Split(*tagsPtr, "-")
	mode := *modePtr

	var pythonFiles []string
	var err error
	var fail int
	var tagPatterns map[string]*regexp.Regexp
	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []fileResults

	if mode == "commit" {
		pythonFiles = flag.Args()
		var filteredFiles []string
		for _, file := range pythonFiles {
			if filepath.Ext(file) == ".py" {
				filteredFiles = append(filteredFiles, file)
			}
		}
		pythonFiles = filteredFiles
	} else if mode == "root" {
		pythonFiles, err = gatherPythonFiles(root)
		if err != nil {
			fmt.Printf("Error gathering Python files: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Error: invalid mode %s. Use 'root' or 'commit'\n", mode)
		os.Exit(1)
	}

	if *configPtr != "" {
		tagPatterns, fail = readConfigFile(*configPtr)
	} else {
		fail = *failPtr
		tagPatterns = make(map[string]*regexp.Regexp, len(tags))
		for _, tag := range tags {
			pattern := fmt.Sprintf(`# ?%s`, regexp.QuoteMeta(tag))
			re, err := regexp.Compile(pattern)
			if err != nil {
				fmt.Printf("Error compiling regex for tag %s: %v\n", tag, err)
				os.Exit(1)
			}
			tagPatterns[tag] = re
		}
	}

	for _, file := range pythonFiles {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			pyResult, err := parseFileForComments(file, tagPatterns)
			if err != nil {
				fmt.Println(err)
				return
			}
			mu.Lock()
			results = append(results, fileResults{file, pyResult})
			mu.Unlock()
		}(file)
	}

	wg.Wait()

	inspectResult(results, fail)
}
