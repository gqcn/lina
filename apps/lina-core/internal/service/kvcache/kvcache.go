// This file defines the distributed KV cache service component and shared value models.

package kvcache

import "github.com/gogf/gf/v2/os/gtime"

// OwnerType defines the supported cache owner types.
type OwnerType string

const (
	// OwnerTypePlugin identifies dynamic plugin-owned cache entries.
	OwnerTypePlugin OwnerType = "plugin"
	// OwnerTypeModule identifies host module-owned cache entries.
	OwnerTypeModule OwnerType = "module"
)

const (
	// ValueKindString identifies string cache values.
	ValueKindString = 1
	// ValueKindInt identifies integer cache values.
	ValueKindInt = 2
)

const (
	maxOwnerTypeBytes = 16
	maxOwnerKeyBytes  = 64
	maxNamespaceBytes = 64
	maxCacheKeyBytes  = 128
	maxValueBytes     = 4096
)

// Service provides distributed KV cache operations backed by the MEMORY cache table.
type Service struct{}

// Item defines one cache entry snapshot.
type Item struct {
	// Key is the logical cache key inside the namespace.
	Key string
	// ValueKind identifies whether the entry stores a string or integer value.
	ValueKind int
	// Value is the string payload of the cache entry.
	Value string
	// IntValue is the integer payload of the cache entry.
	IntValue int64
	// ExpireAt is the optional expiration time.
	ExpireAt *gtime.Time
}

// New creates and returns a new distributed KV cache service instance.
func New() *Service {
	return &Service{}
}

// String returns the canonical owner type value.
func (value OwnerType) String() string {
	return string(value)
}
