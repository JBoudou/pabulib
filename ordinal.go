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

import ()

type OrdinalVote struct {
	// An ordered list of project identifiers.
	Vote []string

	voteBase
}

type OrdinalPB struct {
	*pbBase
}

func newOrdinalVote(section *Section, line int) (ret OrdinalVote) {
	voteStr, ok := section.Cell(line, "vote")
	if !ok {
		panic(MissingRequiredField{"vote"})
	}
	return OrdinalVote{
		voteBase: newVoteBase(section, line),
		Vote:     spliterComma.Split(voteStr, -1),
	}
}

func newOrdinalPB(file *File) (ret OrdinalPB, err error) {
	ret = OrdinalPB{}
	ret.pbBase, err = newPbBase(file)
	return
}

func (self OrdinalPB) Vote(index int) Vote {
	return newOrdinalVote(self.votesSection, index)
}

func (self OrdinalPB) MinLength() int {
	return self.defaultMetaInt("min_length", 1)
}

func (self OrdinalPB) MaxLength() int {
	return self.defaultMetaInt("max_length", self.NumProjects())
}

func (self OrdinalPB) ScoringFn() string {
	return self.defaultMeta("scoring_fn", "Borda")
}
