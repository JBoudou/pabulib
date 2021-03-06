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
	"errors"
	"reflect"
	"strings"
	"testing"
)

type fileTester interface {
	testFile(t *testing.T, file *File, err error)
}

type sectionTester interface {
	testSection(t *testing.T, section *Section)
}

type ftError struct {
	err error
}

func (self ftError) testFile(t *testing.T, _ *File, err error) {
	if err == nil {
		t.Errorf("Expect error %v.", self.err)
	} else if !errors.Is(err, self.err) {
		t.Errorf("Got error %v. Expect error %v.", err, self.err)
	}
}

type ftHasSection struct {
	name  string
	tests []sectionTester
}

func (self *ftHasSection) testFile(t *testing.T, file *File, err error) {
	if err != nil {
		t.Fatalf("Unexpected error %v.", err)
		return
	}
	section, ok := file.Get(self.name)
	if !ok {
		t.Errorf("No section %s.", self.name)
		return
	}
	for _, tst := range self.tests {
		tst.testSection(t, section)
	}
}

type stHasFields struct {
	fields []string
}

func (self stHasFields) testSection(t *testing.T, section *Section) {
	if len(section.Fields) != len(self.fields) {
		t.Errorf("Different fields length. Got %d. Expect %d.", len(section.Fields), len(self.fields))
	}
	for i, got := range section.Fields {
		if expect := self.fields[i]; got != expect {
			t.Errorf("Wrong field %d. Got %s. Expect %s.", i, got, expect)
		}
	}
}

type stHasValues struct {
	values [][]string
}

func (self stHasValues) testSection(t *testing.T, section *Section) {
	if gotLen, expectLen := len(section.Lines), len(self.values); gotLen != expectLen {
		t.Errorf("Different number of lines. Got %d. Expect %d.", gotLen, expectLen)
	}
	for i, gotLine := range section.Lines {
		expectLine := self.values[i]
		if gotLen, expectLen := len(gotLine), len(expectLine); gotLen != expectLen {
			t.Errorf("Different line %d length. Got %d. Expect %d.", i, gotLen, expectLen)
		}
		for j, got := range gotLine {
			if expect := expectLine[j]; got != expect {
				t.Errorf("Wrong value %d at line %d. Got %s. Expect %s.", j, i, got, expect)
			}
		}
	}
}

func TestReadFile(t *testing.T) {
	tests := []struct {
		name  string
		data  string
		tests []fileTester
	}{
		{
			name:  "No fields",
			data:  "SECTION\n",
			tests: []fileTester{ftError{WrongFormat}},
		},
		{
			name:  "One field",
			data:  "foo\nbar\n",
			tests: []fileTester{ftError{WrongFormat}},
		},
		{
			name:  "Wrong number of fields",
			data:  "foo\nbar;baz\none\n",
			tests: []fileTester{ftError{WrongFormat}},
		},
		{
			name: "One section",
			data: "foo \n key  ; value \n bar; baz\nflu;blu\n",
			tests: []fileTester{&ftHasSection{
				name: "foo",
				tests: []sectionTester{
					stHasFields{[]string{"key", "value"}},
					stHasValues{[][]string{{"bar", "baz"}, {"flu", "blu"}}},
				},
			}},
		},
		{
			name: "Two sections",
			data: "First \n key  ; value \n bar; baz\n Second \nha;hb\na;b",
			tests: []fileTester{
				&ftHasSection{
					name: "First",
					tests: []sectionTester{
						stHasFields{[]string{"key", "value"}},
						stHasValues{[][]string{{"bar", "baz"}}},
					},
				},
				&ftHasSection{
					name: "Second",
					tests: []sectionTester{
						stHasFields{[]string{"ha", "hb"}},
						stHasValues{[][]string{{"a", "b"}}},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := ReadFile(strings.NewReader(tt.data))
			for _, tst := range tt.tests {
				tst.testFile(t, file, err)
			}
		})
	}
}

func TestSection_FieldIndexes(t *testing.T) {
	tests := []struct {
		name    string
		section Section
		fields  []string
		indexes []int
		ok      bool
	}{
		{
			name: "Found",
			section: Section{
				Fields: []string{"foo", "bar", "baz"},
				Lines:  [][]string{{"a", "b", "c"}, {"d", "e", "f"}},
			},
			fields:  []string{"baz", "foo"},
			indexes: []int{2, 0},
			ok:      true,
		},
		{
			name: "Missing",
			section: Section{
				Fields: []string{"foo", "bar", "baz"},
				Lines:  [][]string{{"a", "b", "c"}, {"d", "e", "f"}},
			},
			fields:  []string{"bar", "flu"},
			indexes: []int{1, -1},
			ok:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIndexes, gotOk := tt.section.FieldIndexes(tt.fields)
			if gotOk != tt.ok {
				t.Errorf("Wrong ok result. Got %t. Expect %t.", gotOk, tt.ok)
			}
			if !reflect.DeepEqual(gotIndexes, tt.indexes) {
				t.Errorf("Wrong indexes. Got %v. Expect %v.", gotIndexes, tt.indexes)
			}
		})
	}
}

func TestSection_Cell(t *testing.T) {
	tests := []struct {
		name    string
		section Section
		line    int
		field   string
		value   string
		ok      bool
	}{
		{
			name: "Found",
			section: Section{
				Fields: []string{"foo", "bar", "baz"},
				Lines:  [][]string{{"a", "b", "c"}, {"d", "e", "f"}},
			},
			line:  1,
			field: "foo",
			value: "d",
			ok:    true,
		},
		{
			name: "Missing",
			section: Section{
				Fields: []string{"foo", "bar", "baz"},
				Lines:  [][]string{{"a", "b", "c"}, {"d", "e", "f"}},
			},
			line:  0,
			field: "flu",
			value: "",
			ok:    false,
		},
		{
			name: "Out of bounds",
			section: Section{
				Fields: []string{"foo", "bar", "baz"},
				Lines:  [][]string{{"a", "b", "c"}, {"d", "e", "f"}},
			},
			line:  2,
			field: "bar",
			value: "",
			ok:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOk := tt.section.Cell(tt.line, tt.field)
			if gotOk != tt.ok {
				t.Errorf("Wrong ok result. Got %t. Expect %t.", gotOk, tt.ok)
			}
			if gotValue != tt.value {
				t.Errorf("Wrong value. Got %s. Expect %s.", gotValue, tt.value)
			}
		})
	}
}
