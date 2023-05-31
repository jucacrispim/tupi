// Copyright 2020 Juca Crispim <juca@poraodojuca.net>

// This file is part of tupi.

// tupi is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// tupi is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with tupi. If not, see <http://www.gnu.org/licenses/>.

package tupi

import "testing"

func TestLock(t *testing.T) {
	AcquireLock("something")
	if !IsLocked("something") {
		t.Errorf("Lock didn't lock!")
	}

	if IsLocked("otherthing") {
		t.Errorf("Lock locked wrong thing!")
	}

	ReleaseLock("something")

	if IsLocked("something") {
		t.Errorf("Lock was not released")
	}
}
