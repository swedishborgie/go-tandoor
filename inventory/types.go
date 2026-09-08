// Package inventory contains types for Tandoor Inventory resources.
package inventory

import "time"

// Location represents a physical location for storing food
// (e.g. pantry, freezer, refrigerator).
type Location struct {
	ID        int    `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	IsFreezer bool   `json:"is_freezer,omitempty"`
	Household *int   `json:"household,omitempty"`
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
