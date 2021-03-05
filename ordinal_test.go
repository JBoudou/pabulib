// pabulib for Go
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
	"reflect"
	"testing"
)

type ordinalVoteRepr struct {
	id   string
	vote []string
}

func TestOrdinalPB(t *testing.T) {
	tests := []struct {
		name      string
		repr      []namedSection
		minLength int
		maxLength int
		scoringFn string
		votes     []ordinalVoteRepr
	}{
		{
			name: "full",
			repr: []namedSection{
				{
					name: "META",
					section: Section{
						Fields: []string{"key", "value"},
						Lines: [][]string{
							{"num_projects", "4"},
							{"num_votes", "3"},
							{"budget", "1000"},
							{"vote_type", "ordinal"},
							{"min_length", "2"},
							{"max_length", "3"},
							{"scoring_fn", "none"},
							{"rule", "Condorcet"},
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
							{"3", "997"},
							{"4", "996"},
						},
					},
				},
				{
					name: "VOTES",
					section: Section{
						Fields: []string{"voter_id", "vote"},
						Lines: [][]string{
							{"0", "1,2"},
							{"1", "2, 1 ,4"},
							{"2", "2,3"},
						},
					},
				},
			},
			minLength: 2,
			maxLength: 3,
			scoringFn: "none",
			votes: []ordinalVoteRepr{
				{id: "0", vote: []string{"1", "2"}},
				{id: "1", vote: []string{"2", "1", "4"}},
				{id: "2", vote: []string{"2", "3"}},
			},
		},
		{
			name: "default",
			repr: []namedSection{
				{
					name: "META",
					section: Section{
						Fields: []string{"key", "value"},
						Lines: [][]string{
							{"num_projects", "4"},
							{"num_votes", "4"},
							{"budget", "1000"},
							{"vote_type", "ordinal"},
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
							{"3", "997"},
							{"4", "996"},
						},
					},
				},
				{
					name: "VOTES",
					section: Section{
						Fields: []string{"voter_id", "vote"},
						Lines: [][]string{
							{"0", "1"},
							{"1", "2, 1 ,4, 3"},
							{"2", "2,3"},
							{"3", "4,1,3"},
						},
					},
				},
			},
			minLength: 1,
			maxLength: 4,
			scoringFn: "Borda",
			votes: []ordinalVoteRepr{
				{id: "0", vote: []string{"1"}},
				{id: "1", vote: []string{"2", "1", "4", "3"}},
				{id: "2", vote: []string{"2", "3"}},
				{id: "3", vote: []string{"4", "1", "3"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb, err := newOrdinalPB(makeFile(tt.repr))
			mustt(t, err)

			if got := pb.MinLength(); got != tt.minLength {
				t.Errorf("Wrong MinLength. Got %d. Expect %d.", got, tt.minLength)
			}
			if got := pb.MaxLength(); got != tt.maxLength {
				t.Errorf("Wrong MaxLength. Got %d. Expect %d.", got, tt.maxLength)
			}
			if got := pb.ScoringFn(); got != tt.scoringFn {
				t.Errorf("Wrong ScoringFn. Got %s. Expect %s.", got, tt.scoringFn)
			}

			for i, expect := range tt.votes {
				vote, ok := pb.Vote(i).(OrdinalVote)
				if !ok {
					t.Errorf("Vote %v not of type OrdinalVote", pb.Vote(i))
					continue
				}
				if got := vote.Id(); got != expect.id {
					t.Errorf("Wrong Id for vote %d. Got %s. Expect %s.", i, got, expect.id)
				}
				if !reflect.DeepEqual(vote.Vote, expect.vote) {
					t.Errorf("Wrong Vote for vote %d. Got %v. Expect %v.", i, vote.Vote, expect.vote)
				}
			}
		})
	}
}
