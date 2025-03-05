package common

import (
	"encoding/json"
	"io"
)

func ToJSON(v interface{}, w io.Writer) error {
	return json.NewEncoder(w).Encode(v)
}
