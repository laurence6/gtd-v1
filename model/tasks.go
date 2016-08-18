package model

import (
	"bytes"
	"encoding/json"
)

// Tasks is a slice of Task.
type Tasks []Task

// MarshalJSON implements json.Marshaler interface.
func (tasks Tasks) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteString("[")
	for i := 0; i < len(tasks); i++ {
		if i > 0 {
			buf.WriteString(",")
		}
		j, err := json.Marshal(tasks[i])
		if err != nil {
			return nil, err
		}
		buf.Write(j)
	}
	buf.WriteString("]")
	return buf.Bytes(), nil
}
