package batch

import (
	"time"
)

// String represents batch edit form value.
type String struct {
	Value  string `json:"value"`
	Mixed  bool   `json:"mixed"`
	Action Action `json:"action"`
}

// Bool represents batch edit form value.
type Bool struct {
	Value  bool   `json:"value"`
	Mixed  bool   `json:"mixed"`
	Action Action `json:"action"`
}

// Time represents batch edit form value.
type Time struct {
	Value  time.Time `json:"value"`
	Mixed  bool      `json:"mixed"`
	Action Action    `json:"action"`
}

// Int represents batch edit form value.
type Int struct {
	Value  int    `json:"value"`
	Mixed  bool   `json:"mixed"`
	Action Action `json:"action"`
}

// UInt represents batch edit form value.
type UInt struct {
	Value  uint   `json:"value"`
	Mixed  bool   `json:"mixed"`
	Action Action `json:"action"`
}

// Float32 represents batch edit form value.
type Float32 struct {
	Value  float32 `json:"value"`
	Mixed  bool    `json:"mixed"`
	Action Action  `json:"action"`
}

// Float64 represents batch edit form value.
type Float64 struct {
	Value  float64 `json:"value"`
	Mixed  bool    `json:"mixed"`
	Action Action  `json:"action"`
}
