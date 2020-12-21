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

package alarm

import (
	"time"
)

type Event struct {
	Time time.Time
	Data interface{}
}

// Alarm resends events at the requested time, or later.
// Events are always received in the order they've been sent.
// Closing the Send channel asks the alarm to terminate, which is notified by closing Receive.
type Alarm struct {
	Send    chan<- Event
	Receive <-chan Event
}

func run(rcv <-chan Event, send chan<- Event) {
	for true {
		var next Event

		next, ok := <-rcv
		if !ok {
			break
		}

		duration := time.Until(next.Time)
		if duration > 0 {
			time.Sleep(duration)
		}

		send <- next
	}

	close(send)
}

// New creates a new Alarm with the given size for Send.
// Receive is always unbuffered.
func New(chanSize int) Alarm {
	in := make(chan Event, chanSize)
	out := make(chan Event)
	go run(in, out)
	return Alarm{Send: in, Receive: out}
}