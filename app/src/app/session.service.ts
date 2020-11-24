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

import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, Subject } from 'rxjs';
import { map } from 'rxjs/operators';

import { LoginInfo } from './api'

export class SessionInfo {
  registered: boolean;
  user: string;
}

@Injectable({
  providedIn: 'root'
})
export class SessionService {

  sessionId: string = '';

  observable = new Subject<SessionInfo>()

  login(info: LoginInfo): Observable<LoginInfo> {
    return this.http.post('/a/login', info).pipe(
      map((data: string) => {
        this.sessionId = data;
        this.observable.next({registered: true, user: info.User})
        localStorage.setItem("SessionId", this.sessionId);
        localStorage.setItem("User", info.User);
        return info;
      })
    );
  }

  logoff() {
    this.sessionId = ''
    localStorage.removeItem("SessionId");
    localStorage.removeItem("User");
    document.cookie = "s=; Path=/; Max-Age=-1";
    this.observable.next({registered: false, user: ''});
  }

  makeURL(base: string): string {
    if (this.sessionId == '') {
      return base;
    }

    var sep: string = "?";
    if (base.includes(sep)) {
      sep = "&";
    }
    return base + sep + "s=" + this.sessionId;
  }

  constructor(public http: HttpClient) {
  }

  checkSession() {
    if (this.sessionId !== '') {
      return;
    }

    this.sessionId = localStorage.getItem("SessionId");
    if (!!this.sessionId) {
      this.observable.next({registered: true, user: localStorage.getItem("User")});
    }
  }
}