package network

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// QoSLevel is an MQTT quality-of-service level (0, 1 or 2).
//
// The editor renders QoS as a select control, and a select's value is a string
// ("0", "1", "2"), while flows built programmatically or imported from JSON use
// numbers. Decoding into a plain byte accepts only the numeric form, so a node
// configured through the UI failed to initialise with "cannot unmarshal string
// into Go struct field ... of type uint8". Accepting both forms here keeps the
// node usable from either source.
type QoSLevel byte

// UnmarshalJSON decodes a QoS level from either a JSON number or a numeric
// string. An empty string means "unset" and yields QoS 0.
func (q *QoSLevel) UnmarshalJSON(data []byte) error {
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		*q = QoSLevel(num)
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("qos must be a number or a numeric string, got %s", data)
	}
	if s == "" {
		*q = 0
		return nil
	}

	v, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("invalid qos %q: must be 0, 1 or 2", s)
	}
	*q = QoSLevel(v)
	return nil
}

// MarshalJSON emits the numeric form so a round-trip stays canonical.
func (q QoSLevel) MarshalJSON() ([]byte, error) {
	return json.Marshal(byte(q))
}
