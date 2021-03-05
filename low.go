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
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"
)

var (
	WrongFormat = errors.New("Wrong format")
)

type Section struct {
	Fields []string
	Lines  [][]string
}

type File struct {
	sections map[string]*Section
}

func ReadFile(in io.Reader) (ret *File, err error) {
	ret = &File{sections: make(map[string]*Section)}
	scan := bufio.NewScanner(in)
	sectionTitle := ""

	for true {
		for sectionTitle == "" {
			if !scan.Scan() {
				err = scan.Err()
				return
			}
			sectionTitle = strings.TrimSpace(scan.Text())
		}

		var nextTitle string
		var section *Section
		section, nextTitle, err = newSection(scan)
		if err != nil {
			return
		}
		ret.sections[sectionTitle] = section
		sectionTitle = nextTitle
	}

	return
}

var (
	spliterSemicolon = regexp.MustCompile("\\s*;\\s*")
	spliterComma     = regexp.MustCompile("\\s*,\\s*")
)

func splitScanned(scan *bufio.Scanner) []string {
	return spliterSemicolon.Split(strings.TrimSpace(scan.Text()), -1)
}

func newSection(scan *bufio.Scanner) (section *Section, nextTitle string, err error) {
	if !scan.Scan() {
		err = scan.Err()
		if err == nil {
			err = WrongFormat
		}
		return
	}
	section = &Section{Fields: splitScanned(scan)}
	nbFields := len(section.Fields)
	if nbFields == 1 {
		err = WrongFormat
		return
	}

	for scan.Scan() {
		line := splitScanned(scan)
		lineLen := len(line)
		if lineLen != nbFields {
			if lineLen == 1 {
				nextTitle = line[0]
				return
			}
			err = WrongFormat
		}
		section.Lines = append(section.Lines, line)
	}

	err = scan.Err()
	return
}

func (self *File) Get(sectionName string) (section *Section, ok bool) {
	section, ok = self.sections[sectionName]
	return
}

// FieldIndexes searches the indexes corresponding to the given field names.
// The indexes are returned in the same order as the given field names.
// Negative values indicate that field names have not been found.
// The returned boolean is true only if all field names have been found.
func (self *Section) FieldIndexes(fields []string) (indexes []int, ok bool) {
	count := len(fields)
	indexes = make([]int, count)
	posMap := make(map[string]int, count)
	for i, field := range fields {
		indexes[i] = -1
		posMap[field] = i
	}

	found := 0
	for i, field := range self.Fields {
		pos, tmpOk := posMap[field]
		if tmpOk {
			if indexes[pos] < 0 {
				found += 1
			}
			indexes[pos] = i
			if found == count {
				break
			}
		}
	}

	ok = found == count
	return
}

func (self *Section) Cell(line int, field string) (string, bool) {
	if line >= len(self.Lines) {
		return "", false
	}

	index, ok := self.FieldIndexes([]string{field})
	if !ok {
		return "", false
	}

	return self.Lines[line][index[0]], true
}
