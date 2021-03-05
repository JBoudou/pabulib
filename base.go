// pubalib for Go
// Copyright (C) 2021 Joseph Boudou
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU General Public License as published by the Free Software
// Foundation, either version 3 of the License, or (at your option) any later
// version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE. See the GNU General Public License for more
// details.
//
// You should have received a copy of the GNU General Public License along with
// this program. If not, see <https://www.gnu.org/licenses/>.

package pabulib

import (
	"fmt"
	"strconv"
)

// Errors //

type MissingRequiredSection struct {
	Section string
}

func (self MissingRequiredSection) Error() string {
	return fmt.Sprintf("Missing required section %s", self.Section)
}

type MissingRequiredField struct {
	Field string
}

func (self MissingRequiredField) Error() string {
	return fmt.Sprintf("Missing required field %s", self.Field)
}

type DuplicatedMeta struct {
	Meta string
}

func (self DuplicatedMeta) Error() string {
	return fmt.Sprintf("Duplicate meta key %s", self.Meta)
}

type MissingRequiredMeta struct {
	Meta string
}

func (self MissingRequiredMeta) Error() string {
	return fmt.Sprintf("Missing required meta key %s", self.Meta)
}

// Generic types //

const (
	VoteTypeApproval = iota
	VoteTypeOrdinal
	VoteTypeCumulative
	VoteTypeScoring
	VoteTypeUnknown
)

const (
	RuleGreedy = iota
	RuleUnknown
)

type Project interface {
	Id() string
	Cost() int

	Field(name string) (string, bool)
}

type Vote interface {
	Id() string

	Field(name string) (string, bool)
}

type PB interface {
	NumProjects() int
	NumVotes() int
	Budget() int
	VoteType() int
	Rule() int

	Meta(key string) (string, bool)

	// Project returns the project with given identifier.
	Project(id string) (Project, bool)

	// ProjectByIndex returns the i'th project.
	ProjectByIndex(index int) Project

	Vote(index int) Vote
}

// Base implementation //

type fieldBased struct {
	section *Section
	line    int
}

func (self fieldBased) Field(name string) (string, bool) {
	return self.section.Cell(self.line, name)
}

func (self fieldBased) mustField(name string) (ret string) {
	var ok bool
	ret, ok = self.Field(name)
	if !ok {
		panic(MissingRequiredField{name})
	}
	return
}

type projectBase struct {
	fieldBased
}

func newProjectBase(section *Section, line int) projectBase {
	return projectBase{fieldBased: fieldBased{section: section, line: line}}
}

func (self projectBase) Id() string {
	return self.mustField("project_id")
}

func (self projectBase) Cost() (ret int) {
	str := self.mustField("cost")
	ret, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}
	return
}

type voteBase struct {
	fieldBased
}

func newVoteBase(section *Section, line int) voteBase {
	return voteBase{fieldBased: fieldBased{section: section, line: line}}
}

func (self voteBase) Id() string {
	return self.mustField("voter_id")
}

type pbBase struct {
	metaSection     *Section
	projectsSection *Section
	votesSection    *Section
	budget          int // memoized
	projectId       map[string]int
}

func firstMissingField(fields []string, indexes []int) string {
	for i, index := range indexes {
		if index < 0 {
			return fields[i]
		}
	}
	panic("Must never reach this line")
}

func newPbBase(file *File) (ret *pbBase, err error) {
	ret = &pbBase{}

	// Sections
	var ok bool
	if ret.metaSection, ok = file.Get("META"); !ok {
		return ret, MissingRequiredSection{"META"}
	}
	if ret.projectsSection, ok = file.Get("PROJECTS"); !ok {
		return ret, MissingRequiredSection{"PROJECTS"}
	}
	if ret.votesSection, ok = file.Get("VOTES"); !ok {
		return ret, MissingRequiredSection{"VOTES"}
	}

	// Meta
	var strBudget string
	if strBudget, ok = ret.Meta("budget"); !ok {
		return ret, MissingRequiredMeta{"budget"}
	}
	if ret.budget, err = strconv.Atoi(strBudget); err != nil {
		return
	}
	if err = ret.hasAllMeta([]string{"num_projects", "num_votes", "vote_type", "rule"}); err != nil {
		return
	}

	// Fields
	var (
		projectFields = []string{"project_id", "cost"}
		votesFields   = []string{"voter_id", "vote"}
	)
	var projectIndexes []int
	if projectIndexes, ok = ret.projectsSection.FieldIndexes(projectFields); !ok {
		return ret, MissingRequiredField{firstMissingField(projectFields, projectIndexes)}
	}
	if indexes, ok := ret.votesSection.FieldIndexes(votesFields); !ok {
		return ret, MissingRequiredField{firstMissingField(votesFields, indexes)}
	}

	// Projects
	ret.projectId = make(map[string]int, ret.NumProjects())
	for i, project := range ret.projectsSection.Lines {
		ret.projectId[project[projectIndexes[0]]] = i
	}

	return
}

func (self *pbBase) hasAllMeta(meta []string) error {
	count := len(meta)
	metaMap := make(map[string]bool, count)
	for _, m := range meta {
		metaMap[m] = false
	}

	found := 0
	for _, line := range self.metaSection.Lines {
		dup, ok := metaMap[line[0]]
		if !ok {
			continue
		}
		if dup {
			return DuplicatedMeta{line[0]}
		}
		metaMap[line[0]] = true
		found += 1
		if found == count {
			return nil
		}
	}

	for key, fnd := range metaMap {
		if !fnd {
			return MissingRequiredField{key}
		}
	}
	panic("Must never reach this line")
}

func (self *pbBase) NumProjects() int {
	return self.mustMetaInt("num_projects")
}

func (self *pbBase) NumVotes() int {
	return self.mustMetaInt("num_votes")
}

func (self *pbBase) Budget() int {
	return self.budget
}

func (self *pbBase) VoteType() int {
	switch self.mustMeta("vote_type") {
	case "approval":
		return VoteTypeApproval
	case "ordinal":
		return VoteTypeOrdinal
	case "cumulative":
		return VoteTypeCumulative
	case "scoring":
		return VoteTypeScoring
	default:
		return VoteTypeUnknown
	}
}

func (self *pbBase) Rule() int {
	switch self.mustMeta("rule") {
	case "greedy":
		return RuleGreedy
	default:
		return RuleUnknown
	}
}

func (self *pbBase) Meta(key string) (string, bool) {
	for _, line := range self.metaSection.Lines {
		if line[0] == key {
			return line[1], true
		}
	}
	return key, false
}

func (self *pbBase) mustMeta(key string) (ret string) {
	var ok bool
	if ret, ok = self.Meta(key); !ok {
		panic(MissingRequiredMeta{key})
	}
	return
}

func (self *pbBase) mustMetaInt(key string) (ret int) {
	var err error
	if ret, err = strconv.Atoi(self.mustMeta(key)); err != nil {
		panic(err)
	}
	return
}

func (self *pbBase) defaultMeta(key string, _default string) (ret string) {
	var ok bool
	if ret, ok = self.Meta(key); !ok {
		return _default
	}
	return
}

func (self *pbBase) defaultMetaInt(key string, _default int) (ret int) {
	str, ok := self.Meta(key)
	if ok {
		var err error
		ret, err = strconv.Atoi(str)
		ok = err == nil
	}
	if !ok {
		return _default
	}
	return
}

func (self *pbBase) ProjectByIndex(index int) Project {
	return newProjectBase(self.projectsSection, index)
}

func (self *pbBase) Project(id string) (Project, bool) {
	index, ok := self.projectId[id]
	if !ok {
		return nil, false
	}
	return self.ProjectByIndex(index), true
}

func (self *pbBase) Vote(index int) Vote {
	return newVoteBase(self.votesSection, index)
}
