// Package inventory contains types for Tandoor Inventory resources.
package inventory

import (
	"encoding/json"
	"time"

	"github.com/swedishborgie/go-tandoor/idref"
)

// Location represents a physical location for storing food
// (e.g. pantry, freezer, refrigerator).
type Location struct {
	ID        int    `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	IsFreezer bool   `json:"is_freezer,omitempty"`
	Household *int   `json:"household,omitempty"`
}

// MarshalJSON marshals as a bare integer when only an ID is set, otherwise
// as the full object (mirrors Tandoor's writable nested handling).
func (l *Location) MarshalJSON() ([]byte, error) {
	type plain Location
	return idref.MarshalJSON(l.ID, l.Name, plain(*l))
}

// UnmarshalJSON accepts either a bare integer or the full object. The
// household field arrives as a nested object in responses but is written as
// a bare ID, so it is handled separately.
func (l *Location) UnmarshalJSON(data []byte) error {
	l.Household = nil
	if id, ok := idref.AsID(data); ok {
		l.ID = id
		return nil
	}
	var aux struct {
		ID        int             `json:"id,omitempty"`
		Name      string          `json:"name,omitempty"`
		IsFreezer bool            `json:"is_freezer,omitempty"`
		Household json.RawMessage `json:"household"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	l.ID, l.Name, l.IsFreezer = aux.ID, aux.Name, aux.IsFreezer
	if len(aux.Household) > 0 {
		if id, ok := idref.AsID(aux.Household); ok {
			l.Household = &id
		} else {
			var h struct {
				ID int `json:"id"`
			}
			if err := json.Unmarshal(aux.Household, &h); err == nil && h.ID > 0 {
				l.Household = &h.ID
			}
		}
	}
	return nil
}

// Entry represents a tracked food item in inventory.
type Entry struct {
	ID int `json:"id,omitempty"`

	// Location is where the item is stored.
	Location *Location `json:"inventory_location,omitempty"`

	// SubLocation is a human-readable sub-location description
	// (e.g. "back shelf", "door compartment").
	SubLocation string `json:"sub_location,omitempty"`

	// Code is a unique identifier (auto-generated if not provided).
	Code string `json:"code,omitempty"`

	// Food is the food this entry tracks.
	Food *int `json:"food,omitempty"`

	// Unit is the measurement unit (optional).
	Unit *int `json:"unit,omitempty"`

	// Amount is the current quantity of this item.
	Amount float64 `json:"amount,omitempty"`

	// Expires is the expiration date (optional).
	Expires *time.Time `json:"expires,omitempty"`

	// Note is a free-text note.
	Note string `json:"note,omitempty"`

	// Label is a computed display string (read-only).
	Label string `json:"label,omitempty"`

	CreatedAt time.Time `json:"created_at,omitempty"`

	// CreatedBy is the user who created the entry (read-only).
	CreatedBy int `json:"created_by,omitempty"`
}

// UnmarshalJSON accepts food and unit as bare IDs or full objects (the API
// returns nested objects); writes send bare IDs.
func (e *Entry) UnmarshalJSON(data []byte) error {
	type plain Entry
	var p plain
	aux := struct {
		*plain
		Food json.RawMessage `json:"food"`
		Unit json.RawMessage `json:"unit"`
	}{plain: &p}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*e = Entry(p)
	if id, ok := refID(aux.Food); ok {
		e.Food = &id
	}
	if id, ok := refID(aux.Unit); ok {
		e.Unit = &id
	}
	return nil
}

// refID decodes a reference field that may be a bare integer or a nested
// object with an id; ok is false for null/missing.
func refID(raw json.RawMessage) (int, bool) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, false
	}
	if id, ok := idref.AsID(raw); ok {
		return id, true
	}
	var o struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(raw, &o); err != nil {
		return 0, false
	}
	return o.ID, true
}

// Log records a change to an inventory entry.
// Read-only — cannot be created/updated via the API.
type Log struct {
	ID int `json:"id"`

	// Entry is the inventory entry that was changed.
	Entry Entry `json:"entry"`

	// BookingType indicates the type of change:
	//   - "B_ADD": item added
	//   - "B_REMOVE": quantity removed
	//   - "B_MOVE": item moved to different location
	BookingType string `json:"booking_type"`

	// OldAmount is the amount before the change.
	OldAmount float64 `json:"old_amount"`

	// NewAmount is the amount after the change.
	NewAmount float64 `json:"new_amount"`

	// OldInventoryLocation is the location before the change.
	OldLocation Location `json:"old_inventory_location"`

	// NewInventoryLocation is the location after the change.
	NewLocation Location `json:"new_inventory_location"`

	// Note is a free-text note about the change.
	Note string `json:"note"`

	// CreatedAt is when the change was made.
	CreatedAt time.Time `json:"created_at"`
}
