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
	"testing"
)

func mustt(t *testing.T, err error) {
	if err == nil {
		return
	}
	t.Fatal(err)
}

type namedSection struct {
	name    string
	section Section
}

func makeFile(repr []namedSection) *File {
	m := make(map[string]*Section, len(repr))
	for i := range repr {
		m[repr[i].name] = &repr[i].section
	}
	return &File{sections: m}
}

func makePBBase(repr []namedSection) (*pbBase, error) {
	return newPbBase(makeFile(repr))
}

type projectRepr struct {
	id   string
	cost int
}

func (self *projectRepr) check(t *testing.T, index int, project Project) {
	if got := project.Id(); got != self.id {
		t.Errorf("Wrong Id for project %d. Got %s. Expect %s.", index, got, self.id)
	}
	if got := project.Cost(); got != self.cost {
		t.Errorf("Wrong Cost for project %d. Got %d. Expect %d.", index, got, self.cost)
	}
}

func TestPBBase(t *testing.T) {
	tests := []struct {
		name        string
		repr        []namedSection
		numProjects int
		numVotes    int
		budget      int
		voteType    int
		rule        int
		projects    []projectRepr
		votes       []string
	}{
		{
			name: "simple",
			repr: []namedSection{
				{
					name: "META",
					section: Section{
						Fields: []string{"key", "value"},
						Lines: [][]string{
							{"num_projects", "2"},
							{"num_votes", "3"},
							{"budget", "1000"},
							{"vote_type", "approval"},
							{"rule", "greedy"},
						},
					},
				},
				{
					name: "PROJECTS",
					section: Section{
						Fields: []string{"project_id", "cost"},
						Lines: [][]string{
							{"1", "999"},
							{"2", "998"},
						},
					},
				},
				{
					name: "VOTES",
					section: Section{
						Fields: []string{"voter_id", "vote"},
						Lines: [][]string{
							{"0", "1"},
							{"1", "0"},
							{"2", "0"},
						},
					},
				},
			},
			numProjects: 2,
			numVotes:    3,
			budget:      1000,
			voteType:    VoteTypeApproval,
			rule:        RuleGreedy,
			projects: []projectRepr{
				{id: "1", cost: 999},
				{id: "2", cost: 998},
			},
			votes: []string{"0", "1", "2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb, err := makePBBase(tt.repr)
			mustt(t, err)

			if got := pb.NumProjects(); got != tt.numProjects {
				t.Errorf("Wrong NumProjects. Got %d. Expect %d.", got, tt.numProjects)
			}
			if got := pb.NumVotes(); got != tt.numVotes {
				t.Errorf("Wrong NumVotes. Got %d. Expect %d.", got, tt.numVotes)
			}
			if got := pb.Budget(); got != tt.budget {
				t.Errorf("Wrong Budget. Got %d. Expect %d.", got, tt.budget)
			}
			if got := pb.VoteType(); got != tt.voteType {
				t.Errorf("Wrong VoteType. Got %d. Expect %d.", got, tt.voteType)
			}
			if got := pb.Rule(); got != tt.rule {
				t.Errorf("Wrong Rule. Got %d. Expect %d.", got, tt.rule)
			}

			for i, expect := range tt.projects {
				expect.check(t, i, pb.ProjectByIndex(i))
				
				project, found := pb.Project(expect.id)
				if !found {
					t.Errorf("Project with id %s not found.", expect.id)
				} else {
					expect.check(t, i, project)
				}
			}

			for i, expect := range tt.votes {
				vote := pb.Vote(i)
				if got := vote.Id(); got != expect {
					t.Errorf("Wrong Id for vote %d. Got %s. Expect %s.", i, got, expect)
				}
			}

		})
	}
}
