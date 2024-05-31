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
)

// Create a struct to store the tag, path, line number and line content
type tagFound struct {
	tag        string
	path       string
	lineNumber int
	line       string
}

type fileResults struct {
	fileName  string
	tagsFound []tagFound
}

// gatherPythonFiles returns a list of all python files in the given directory
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
	results := make([]tagFound, 0)

	file, err := os.Open(path)
	if err != nil {
		return results, fmt.Errorf("error opening file %s: %v", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 1

	for scanner.Scan() {
		line := scanner.Text()
		// Check if the line contains any of the tags
		for tag, pattern := range tagPatterns {
			// When line is empty, skip the line
			if len(line) == 0 {
				continue
			}

			// If the line contains the tag, print the line and the tag
			if pattern.MatchString(line) {
				line = strings.TrimSpace(line)
				//fmt.Printf("Found %s in %s at line %d: %s\n", tag, path, lineNumber, line)
				found := tagFound{tag, path, lineNumber, line}
				results = append(results, found)
			}
		}
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return results, fmt.Errorf("error reading file %s: %v", path, err)
	}
	return results, nil
}

func main() {
	// Define command-line flags
	dirPtr := flag.String("dir", ".", "the root directory to scan for Python files")
	tagsPtr := flag.String("tags", "TODO,BUG,FIXME", "comma-separated list of tags to search for")
	modePtr := flag.String("mode", "commit", "mode of operation: commit or root")

	// Parse the command-line flags
	flag.Parse()

	root := *dirPtr
	tags := strings.Split(*tagsPtr, ",")
	mode := *modePtr

	// Compile regex patterns for the tags and store them in a map
	tagPatterns := make(map[string]*regexp.Regexp, len(tags))
	for _, tag := range tags {
		pattern := fmt.Sprintf(`# ?%s`, regexp.QuoteMeta(tag))
		re, err := regexp.Compile(pattern)
		if err != nil {
			fmt.Printf("Error compiling regex for tag %s: %v\n", tag, err)
			os.Exit(1)
		}
		tagPatterns[tag] = re
	}

	var pythonFiles []string
	var err error

	// Gather the Python files to scan from the command-line arguments
	if mode == "commit" {
		pythonFiles = flag.Args()
		// Filter out non-Python files
		var filteredFiles []string
		for _, file := range pythonFiles {
			if filepath.Ext(file) == ".py" {
				filteredFiles = append(filteredFiles, file)
			}
		}
		pythonFiles = filteredFiles
	} else if mode == "root" {
		// Gather all Python files in the root directory
		pythonFiles, err = gatherPythonFiles(root)
		if err != nil {
			fmt.Printf("Error gathering Python files: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Error: invalid mode %s. Use 'auto' or 'commit'\n", mode)
		os.Exit(1)
	}

	// Use a wait group to process files concurrently
	var wg sync.WaitGroup
	var results = make([]fileResults, 0)
	for _, file := range pythonFiles {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			pyResult, err := parseFileForComments(file, tagPatterns)
			if err != nil {
				fmt.Println(err)
			}
			results = append(results, fileResults{file, pyResult})
		}(file)
	}
	// Wait for all goroutines to complete
	wg.Wait()
	fmt.Print(results)
}
