package main

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Tag represents a single tag.
type Tag struct {
	Name    string
	File    string
	Address string
	Type    TagType
	Fields  map[TagField]string
}

// TagField represents a single field in a tag line.
type TagField string

// Tag fields.
const (
	Access        TagField = "access"
	Signature     TagField = "signature"
	TypeField     TagField = "type"
	PrimitiveType TagField = "ptype"
	StepType      TagField = "stype"
	Line          TagField = "line"
	Language      TagField = "language"
)

// TagType represents the type of a tag in a tag line.
type TagType string

// Tag types.
const (
	Primitive    TagType = "p"
	Group        TagType = "g"
	Job          TagType = "j"
	Resource     TagType = "r"
	ResourceType TagType = "t"
	JobOutput    TagType = "o"
	JobInput     TagType = "i"
	JobTask      TagType = "k"
)

// NewTag creates a new Tag.
func NewTag(name, file, addr string, line int, tagType TagType) Tag {
	l := strconv.Itoa(line)
	return Tag{
		Name:    name,
		File:    file,
		Address: addr,
		Type:    tagType,
		Fields:  map[TagField]string{Line: l},
	}
}

// The tags file format string representation of this tag.
func (t Tag) String() string {
	var b bytes.Buffer

	b.WriteString(t.Name)
	b.WriteByte('\t')
	b.WriteString(t.File)
	b.WriteByte('\t')
	b.WriteString(t.Address)
	b.WriteString(";\"\t")
	b.WriteString(string(t.Type))
	b.WriteByte('\t')

	fields := make([]string, 0, len(t.Fields))
	i := 0
	for k, v := range t.Fields {
		if len(v) == 0 {
			continue
		}
		fields = append(fields, fmt.Sprintf("%s:%s", k, v))
		i++
	}

	sort.Sort(sort.StringSlice(fields))
	b.WriteString(strings.Join(fields, "\t"))

	return b.String()
}
