// Package idref provides JSON helpers for Tandoor reference fields that
// accept or emit either a bare integer ID or the full object, mirroring
// Tandoor's WritableNestedModelSerializer, which expands bare IDs
// server-side.
//
// Each reference type in this module embeds no generic machinery; instead it
// declares a two-method pair that delegates here:
//
//	func (u *Unit) MarshalJSON() ([]byte, error) {
//		type plain Unit
//		return idref.MarshalJSON(u.ID, u.Name, plain(*u))
//	}
//
//	func (u *Unit) UnmarshalJSON(data []byte) error {
//		type plain Unit
//		var p plain
//		if err := idref.UnmarshalJSON(data, &p.ID, &p); err != nil {
//			return err
//		}
//		*u = Unit(p)
//		return nil
//	}
//
// Go does not allow a generic struct to embed its type parameter, so the
// method pair cannot be provided by a single generic type; the pair above is
// the minimal per-type glue.
package idref

import "encoding/json"

// MarshalJSON returns data as a bare integer when id != 0 and name == "",
// otherwise marshals v (the full object).
func MarshalJSON(id int, name string, v any) ([]byte, error) {
	if id != 0 && name == "" {
		return json.Marshal(id)
	}
	return json.Marshal(v)
}

// UnmarshalJSON decodes a bare integer into id, or the full object into v.
// On the bare-integer path v is left untouched, so callers should reset v
// (e.g. by unmarshaling into a fresh local) before copying it back.
func UnmarshalJSON(data []byte, id *int, v any) error {
	if n, ok := AsID(data); ok {
		*id = n
		return nil
	}
	return json.Unmarshal(data, v)
}

// AsID reports whether data is a bare JSON integer and returns its value.
func AsID(data []byte) (int, bool) {
	var n int
	if err := json.Unmarshal(data, &n); err != nil {
		return 0, false
	}
	return n, true
}

// AsAnyID extracts an integer ID from a decoded reference field. Refs
// decoded from JSON arrive either as a bare integer (often via float64
// when unmarshaled into interface{}) or as the full object (as a map with
// an "id" key). Unset refs (nil) yield 0.
func AsAnyID(v any) int {
	switch x := v.(type) {
	case nil:
		return 0
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	case map[string]any:
		if id, ok := x["id"].(float64); ok {
			return int(id)
		}
	}
	return 0
}
