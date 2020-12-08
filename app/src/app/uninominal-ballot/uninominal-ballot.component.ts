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


import { HttpClient } from '@angular/common/http';
import { Component, Input, OnInit } from '@angular/core';
import { FormBuilder, FormGroup } from '@angular/forms';

import { PollSubComponent } from '../poll/common';
import { PollAlternative, UninomialBallotAnswer } from '../api';

@Component({
  selector: 'app-uninominal-ballot',
  templateUrl: './uninominal-ballot.component.html',
  styleUrls: ['./uninominal-ballot.component.sass']
})
export class UninominalBallotComponent implements OnInit, PollSubComponent {

  @Input() pollSegment: string;

  answer: UninomialBallotAnswer;

  form = this.formBuilder.group({
    Choice: [''],
  });

  constructor(
    private http: HttpClient,
    private formBuilder: FormBuilder,
  ) { }

  ngOnInit(): void {
    this.http.get<UninomialBallotAnswer>('/a/ballot/uninominal/' + this.pollSegment).subscribe({
      next: (answer: UninomialBallotAnswer) => {
        this.answer = answer;
      }
    });
  }

  hasPrevious(): boolean {
    return this.answer !== undefined && this.answer.Previous !== undefined;
  }

  previous(): string|null {
    return this.nameOf(this.answer.Previous);
  }

  hasCurrent(): boolean {
    return this.answer != undefined && this.answer.Current !== undefined;
  }

  current(): string|null {
    return this.nameOf(this.answer.Current);
  }

  private nameOf(id: number|undefined): string|null {
    if (id === undefined) {
      return null;
    }
    var alternative: PollAlternative;
    for (alternative of this.answer.Alternatives) {
      if (alternative.Id == id!) {
        return alternative.Name;
      }
    }
    return null;
  }

}
