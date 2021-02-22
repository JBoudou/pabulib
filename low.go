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

package pubalib

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
	Headers []string
	Lines   [][]string
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
	spliter = regexp.MustCompile("\\s*;\\s*")
)

func splitScanned(scan *bufio.Scanner) []string {
	return spliter.Split(strings.TrimSpace(scan.Text()), -1)
}

func newSection(scan *bufio.Scanner) (section *Section, nextTitle string, err error) {
	if !scan.Scan() {
		err = scan.Err()
		if err == nil {
			err = WrongFormat
		}
		return
	}
	section = &Section{Headers: splitScanned(scan)}
	nbFields := len(section.Headers)
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
