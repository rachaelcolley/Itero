// Itero - Online iterative vote application
// Copyright (C) 2021 Joseph Boudou
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/JBoudou/Itero/main/services"
	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/salted"
	"github.com/JBoudou/Itero/mid/server"
	srvt "github.com/JBoudou/Itero/mid/server/servertest"
	"github.com/JBoudou/Itero/pkg/events"
	"github.com/JBoudou/Itero/pkg/events/eventstest"
	"github.com/JBoudou/Itero/pkg/ioc"
)

type createPollTest_ struct {
	Name       string
	Unlogged   bool
	Verified   bool // Whether the user is verified
	RequestFct RequestFct
	Checker    srvt.Checker
}

func CreatePollTest(c createPollTest_) *createPollTest {
	return &createPollTest{
		WithName: srvt.WithName{Name: c.Name},
		WithUser: WithUser{
			Unlogged:   c.Unlogged,
			Verified:   c.Verified,
			RequestFct: c.RequestFct,
		},
		Checker: c.Checker,
	}
}

type createPollTest struct {
	srvt.WithName
	WithUser

	Verified bool
	Checker  srvt.Checker

	pollsCreated map[uint32]bool
}

func (self *createPollTest) Prepare(t *testing.T) *ioc.Locator {
	t.Parallel()

	if checker, ok := self.Checker.(interface{ Before(t *testing.T) }); ok {
		checker.Before(t)
	}

	self.pollsCreated = make(map[uint32]bool, 2)
	locator := self.WithUser.Prepare(t).Sub()
	locator.Bind(func() events.Manager {
		return &eventstest.ManagerMock{
			T: t,
			Send_: func(evt events.Event) error {
				if vEvt, ok := evt.(services.CreatePollEvent); ok {
					self.pollsCreated[vEvt.Poll] = true
				}
				return nil
			},
		}
	})
	return locator
}

func (self *createPollTest) Check(t *testing.T, response *http.Response, request *server.Request) {
	if self.Checker != nil {
		self.Checker.Check(t, response, request)
		return
	}

	srvt.CheckStatus{http.StatusOK}.Check(t, response, request)

	const (
		qCheckPoll = `
			SELECT Title, Description, Admin, State, Start, Salt, Electorate, Hidden, ReportVote,
			       MinNbRounds, MaxNbRounds, Deadline, CurrentRoundStart,
						 ADDTIME(CurrentRoundStart, MaxRoundDuration), RoundThreshold
			  FROM Polls
			 WHERE Id = ?`
		qCheckAlternative = `SELECT Name FROM Alternatives WHERE Poll = ? ORDER BY Id ASC`
		qCleanUp          = `DELETE FROM Polls WHERE Id = ?`
	)

	var answer string
	mustt(t, json.NewDecoder(response.Body).Decode(&answer))
	pollSegment, err := salted.Decode(answer)
	mustt(t, err)
	defer func() {
		db.DB.Exec(qCleanUp, pollSegment.Id)
	}()

	query := defaultCreateQuery()
	mustt(t, request.UnmarshalJSONBody(&query))

	// Check Polls
	row := db.DB.QueryRow(qCheckPoll, pollSegment.Id)
	got := query
	var admin uint32
	var state string
	var startDate sql.NullTime
	var salt uint32
	var electorate db.Electorate
	var roundStart, roundEnd time.Time
	mustt(t, row.Scan(
		&got.Title,
		&got.Description,
		&admin,
		&state,
		&startDate,
		&salt,
		&electorate,
		&got.Hidden,
		&got.ReportVote,
		&got.MinNbRounds,
		&got.MaxNbRounds,
		&got.Deadline,
		&roundStart,
		&roundEnd,
		&got.RoundThreshold,
	))
	if salt != pollSegment.Salt {
		t.Errorf("Wrong salt. Got %d. Expect %d.", salt, pollSegment.Salt)
	}
	if admin != self.User.Id {
		t.Errorf("Wrong admin. Got %d. Expect %d.", admin, self.User.Id)
	}
	var expectState string
	if startDate.Valid {
		expectState = "Waiting"
		got.Start = startDate.Time.UTC()
	} else {
		expectState = "Active"
		got.Start = time.Time{}
	}
	if state != expectState {
		t.Errorf("Wrong state. Got %s. Expect %s.", state, expectState)
	}
	dateDiff := got.Deadline.Sub(query.Deadline)
	if (dateDiff >= 1*time.Second) || (dateDiff <= -1*time.Second) {
		t.Errorf("Deadline differ. Got %v. Expect %v. Difference %v.",
			got.Deadline, query.Deadline, dateDiff)
	}
	got.Electorate = electorateFromDB(electorate)
	got.Deadline = query.Deadline
	got.MaxRoundDuration = uint64(roundEnd.Sub(roundStart).Milliseconds())
	if !reflect.DeepEqual(got, query) {
		t.Errorf("Got %v. Expect %v.", got, query)
	}

	// Check Alternatives
	rows, err := db.DB.Query(qCheckAlternative, pollSegment.Id)
	mustt(t, err)
	defer rows.Close()
	for id, alt := range query.Alternatives {
		if !rows.Next() {
			t.Errorf("Premature end of the alternatives. Got %d. Expect %d.", id, len(query.Alternatives))
			break
		}
		var name string
		mustt(t, rows.Scan(&name))
		if name != alt.Name {
			t.Errorf("Wrong alternative %d. Got %s. Expect %s.", id, name, alt.Name)
		}
	}
	if rows.Next() {
		t.Errorf("Unexpected alternatives.")
		rows.Close()
	}

	// Check events
	_, ok := self.pollsCreated[pollSegment.Id]
	if !ok {
		t.Errorf("CreatePollEvent not sent")
	}
}

func electorateFromDB(electorate db.Electorate) CreatePollElectorate {
	switch electorate {
	case db.ElectorateAll:
		return CreatePollElectorateAll
	case db.ElectorateVerified:
		return CreatePollElectorateVerified
	default:
		return CreatePollElectorateLogged
	}
}

func TestCreateHandler(t *testing.T) {
	precheck(t)

	makeBody := func(innerBody string, alternatives []string) string {
		pollAlternatives := make([]SimpleAlternative, len(alternatives))
		for id, name := range alternatives {
			pollAlternatives[id] = SimpleAlternative{Name: name, Cost: 1.}
		}
		encoded, err := json.Marshal(pollAlternatives)
		mustt(t, err)
		return `{"Title":"Test",` + innerBody + `"Alternatives": ` + string(encoded) + "}"
	}

	tests := []srvt.Test{
		CreatePollTest(createPollTest_{
			Name:       "No user",
			RequestFct: RFPostNoSession(makeBody("", []string{"No", "Yes"})),
			Checker:    srvt.CheckStatus{http.StatusForbidden},
		}),
		CreatePollTest(createPollTest_{
			Name: "GET",
			RequestFct: func(user *server.User) *srvt.Request {
				return &srvt.Request{
					UserId: &user.Id,
					Method: "GET",
					Body: `{
						"Title": "Test",
						"Alternatives": [{"Name":"No", "Cost":1}, {"Name":"Yes", "Cost":1}]
					}`,
				}
			},
			Checker: srvt.CheckStatus{http.StatusForbidden},
		}),
		CreatePollTest(createPollTest_{
			Name: "Duplicate",
			RequestFct: RFPostSession(`{
					"Title": "Test",
					"Alternatives": [{"Name":"Yip", "Cost":1}, {"Name":"Yip", "Cost":1}]
				}`),
			Checker: srvt.CheckAnyErrorStatus,
		}),
		CreatePollTest(createPollTest_{
			Name:       "Success",
			RequestFct: RFPostSession(makeBody("", []string{"No", "Yes"})),
		}),
		CreatePollTest(createPollTest_{
			Name:       "Hidden",
			RequestFct: RFPostSession(makeBody(`"Hidden": true,`, []string{"First", "Second"})),
		}),
		CreatePollTest(createPollTest_{
			Name:       "ReportVote",
			RequestFct: RFPostSession(makeBody(`"ReportVote": false,`, []string{"First", "Second"})),
		}),
		CreatePollTest(createPollTest_{
			Name:       "Start later",
			RequestFct: RFPostSession(makeBody(`"Start": "3000-01-01T12:12:12Z",`, []string{"First", "Second"})),
		}),
		CreatePollTest(createPollTest_{
			Name:       "Unlogged",
			Unlogged:   true,
			RequestFct: RFPostSession(makeBody("", []string{"First", "Second"})),
			Checker:    srvt.CheckStatus{http.StatusForbidden},
		}),
		CreatePollTest(createPollTest_{
			Name:       "Unverified",
			Verified:   false,
			RequestFct: RFPostSession(makeBody(`"Electorate": 1,`, []string{"First", "Second"})),
			Checker:    srvt.CheckStatus{http.StatusBadRequest},
		}),
		CreatePollTest(createPollTest_{
			Name:       "Verified",
			Verified:   true,
			RequestFct: RFPostSession(makeBody(`"Electorate": 1,`, []string{"First", "Second"})),
		}),
		CreatePollTest(createPollTest_{
			Name:       "ElectorateAll",
			RequestFct: RFPostSession(makeBody(`"Electorate": -1,`, []string{"First", "Second"})),
		}),
	}

	srvt.Run(t, tests, CreateHandler)
}
