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

import { Component, OnInit } from '@angular/core';

import { SessionService, SessionInfo } from '../session.service';

@Component({
  selector: 'app-navtitle',
  templateUrl: './navtitle.component.html',
  styleUrls: ['./navtitle.component.sass']
})
export class NavtitleComponent implements OnInit {

  logged = false
  sessionUser = ''

  constructor(private session: SessionService
             ) { }

  ngOnInit(): void {
    this.session.observable.subscribe({
      next: (info: SessionInfo) => {
        this.logged = info.registered;
        this.sessionUser = info.user;
      }
    });
  }

}
