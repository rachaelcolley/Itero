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

import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ReactiveFormsModule, FormBuilder, FormsModule } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';

import { EMPTY } from 'rxjs';

import { MatIconModule } from '@angular/material/icon';

import { SimpleAlternativesComponent } from './simple-alternatives.component';

import { CreateService } from '../create.service';

import { ActivatedRouteStub } from '../../../testing/activated-route-stub'

describe('SimpleAlternativesComponent', () => {
  let component: SimpleAlternativesComponent;
  let fixture: ComponentFixture<SimpleAlternativesComponent>;
  let serviceSpy: jasmine.SpyObj<CreateService>;
  let activatedRouteStub: ActivatedRouteStub;

  beforeEach(async () => {
    serviceSpy = jasmine.createSpyObj('CreateService', {register: {}}, { query$: EMPTY });
    activatedRouteStub = new ActivatedRouteStub();
    
    await TestBed.configureTestingModule({
      declarations: [ SimpleAlternativesComponent ],
      imports: [ ReactiveFormsModule, FormsModule, MatIconModule ],
      providers: [
        FormBuilder,
        { provide: CreateService, useValue: serviceSpy },
        { provide: ActivatedRoute, useValue: activatedRouteStub },
      ],
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(SimpleAlternativesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
