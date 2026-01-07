package utils

import "encoding/json"

// MustJSON помощник для быстрой сериализации в []byte; паника маловероятна при корректных структурах.
func MustJSON(v any) []byte {
    b, _ := json.Marshal(v)
    return b
}


