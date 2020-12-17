// Itero - Online iterative vote application
// Copyright (C) 2020 Joseph Boudou
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

package main

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/server"
)

type NuDate sql.NullTime

func (self NuDate) MarshalJSON() ([]byte, error) {
	if !self.Valid {
		return []byte(`"⋅"`), nil
	} else {
		return self.Time.MarshalJSON()
	}
}

func (self *NuDate) UnmarshalJSON(raw []byte) (err error) {
	if string(raw) == `"⋅"` {
		self.Valid = false
		return
	}
	err = self.Time.UnmarshalJSON(raw)
	self.Valid = err == nil
	return
}

type listAnswerEntry struct {
	Segment      string `json:"s"`
	Title        string `json:"t"`
	CurrentRound uint8  `json:"c"`
	MaxRound     uint8  `json:"m"`
	Deadline     NuDate `json:"d"`
	Action       string `json:"a"` // TODO Use an "enum" ?
}

func ListHandler(ctx context.Context, response server.Response, request *server.Request) {
	reply := make([]listAnswerEntry, 0, 16)

	if request.User == nil {
		// TODO change that
		response.SendError(ctx, server.NewHttpError(http.StatusNotImplemented, "Unimplemented", ""))
		return
	}

	const query = `
	   SELECT p.Id, p.Salt, p.Title, p.CurrentRound, p.MaxNbRounds,
	          addtime(p.CurrentRoundStart, p.MaxRoundDuration) AS Deadline,
	          CASE WHEN a.User IS NULL THEN 'Part'
	               WHEN a.LastRound >= p.CurrentRound THEN 'Modi'
	               ELSE 'Vote' END AS Action
	     FROM Polls AS p LEFT OUTER JOIN (
	              SELECT Poll, User, LastRound
	               FROM Participants
	              WHERE User = ?
	          ) AS a ON p.Id = a.Poll
	    WHERE p.Active
	      AND ((p.CurrentRound = 0 AND p.Publicity <= ?) OR a.User IS NOT NULL)
	 ORDER BY Action DESC, Deadline ASC`
	rows, err := db.DB.QueryContext(ctx, query, request.User.Id, db.PollPublicityPublicRegistered)
	if err != nil {
		response.SendError(ctx, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var listAnswerEntry listAnswerEntry
		var segment PollSegment
		var deadline sql.NullTime

		err = rows.Scan(&segment.Id, &segment.Salt, &listAnswerEntry.Title,
			&listAnswerEntry.CurrentRound, &listAnswerEntry.MaxRound, &deadline,
			&listAnswerEntry.Action)
		if err != nil {
			response.SendError(ctx, err)
			return
		}

		listAnswerEntry.Deadline = NuDate(deadline)
		listAnswerEntry.Segment, err = segment.Encode()
		if err != nil {
			response.SendError(ctx, err)
			return
		}

		reply = append(reply, listAnswerEntry)
	}

	response.SendJSON(ctx, reply)
	return
}
