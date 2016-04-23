package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// tagParser contains the data needed while parsing.
type tagParser struct {
	file     string
	tags     []Tag    // list of created tags
	types    []string // all types we encounter, used to determine the constructors
	relative bool     // should filenames be relative to basepath
	basepath string   // output file directory
}

// Parse parses the source in filename and returns a list of tags. If relative
// is true, the filenames in the list of tags are relative to basepath.
func Parse(filename string, relative bool, basepath string) ([]Tag, error) {
	p := &tagParser{
		tags:     []Tag{},
		types:    make([]string, 0),
		relative: relative,
		basepath: basepath,
		file:     filename,
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	// declarations
	p.parse(f)

	return p.tags, nil
}

// parseDeclarations creates a tag for each function, type or value declaration.
func (p *tagParser) parse(f io.Reader) {
	r := bufio.NewReader(f)
	var currentSection TagType
	var currentPrimitive string
	var currentJob, currentStep *Tag
	lineNum := 0

	for {
		lineNum++
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "couldn't read input file: %s\n", err)
			return
		}
		sectionRegexp := regexp.MustCompile("^(groups|jobs|resources|resource_types):")
		sectionMatch := sectionRegexp.FindStringSubmatch(line)
		if len(sectionMatch) > 0 {
			name := sectionMatch[1]
			switch name {
			case "groups":
				currentSection = Group
			case "jobs":
				currentSection = Job
			case "resources":
				currentSection = Resource
			case "resource_types":
				currentSection = ResourceType
			}
			currentPrimitive = name

			addr := fmt.Sprintf("/\\%%%dl\\%%%dc/", lineNum, strings.Index(string(line), name)+1)
			tag := p.createTag(name, addr, lineNum, Primitive)
			tag.Fields[TypeField] = "section"
			p.tags = append(p.tags, tag)
			continue
		}

		lineRegexp := regexp.MustCompile("^- name: (.*)")
		lineMatch := lineRegexp.FindStringSubmatch(line)
		if len(lineMatch) > 0 {
			name := lineMatch[1]
			addr := fmt.Sprintf("/\\%%%dl\\%%%dc/", lineNum, strings.Index(string(line), name)+1)
			tag := p.createTag(name, addr, lineNum, currentSection)
			tag.Fields[PrimitiveType] = currentPrimitive
			tag.Fields[TypeField] = currentPrimitive
			if currentSection == Job {
				tag.Fields[Access] = "private"
			}
			currentJob = &tag
			p.tags = append(p.tags, tag)
			continue
		}

		lineRegexp = regexp.MustCompile("^  type: (.*)")
		lineMatch = lineRegexp.FindStringSubmatch(line)
		if len(lineMatch) > 0 {
			t := lineMatch[1]
			currentJob.Fields[TypeField] = t
			continue
		}

		lineRegexp = regexp.MustCompile("^  public: (.*)")
		lineMatch = lineRegexp.FindStringSubmatch(line)
		if len(lineMatch) > 0 {
			value := lineMatch[1]
			if value == "true" {
				currentJob.Fields[Access] = "public"
			} else {
				currentJob.Fields[Access] = "private"
			}
			continue
		}

		lineRegexp = regexp.MustCompile("trigger: (.*)")
		lineMatch = lineRegexp.FindStringSubmatch(line)
		if len(lineMatch) > 0 {
			value := lineMatch[1]
			if value == "true" {
				currentStep.Fields[Access] = "public"
				currentStep.Fields[TypeField] = "[trigger]"
			}
			continue
		}

		lineRegexp = regexp.MustCompile("passed: \\[(.*)\\]")
		lineMatch = lineRegexp.FindStringSubmatch(line)
		if len(lineMatch) > 0 {
			value := lineMatch[1]
			currentStep.Fields[Signature] = fmt.Sprintf("(passed: [%s])", value)
			continue
		}

		lineRegexp = regexp.MustCompile("^\\s*-? (get|put|task): (.*)")
		lineMatch = lineRegexp.FindStringSubmatch(line)
		if len(lineMatch) > 0 {
			t := lineMatch[1]
			name := lineMatch[2]
			var kind TagType
			switch t {
			case "get":
				kind = JobInput
			case "put":
				kind = JobOutput
			case "task":
				kind = JobTask
			}
			tag := p.createTag(name, "", lineNum, kind)
			tag.Fields[StepType] = currentPrimitive + "." + currentJob.Name
			currentStep = &tag
			p.tags = append(p.tags, tag)
			continue
		}
	}
}

// createTag creates a new tag, using pos to find the filename and set the line number.
func (p *tagParser) createTag(name, addr string, line int, tagType TagType) Tag {
	f := p.file
	if p.relative {
		if abs, err := filepath.Abs(f); err != nil {
			fmt.Fprintf(os.Stderr, "could not determine absolute path: %s\n", err)
		} else if rel, err := filepath.Rel(p.basepath, abs); err != nil {
			fmt.Fprintf(os.Stderr, "could not determine relative path: %s\n", err)
		} else {
			f = rel
		}
	}
	return NewTag(name, f, addr, line, tagType)
}
