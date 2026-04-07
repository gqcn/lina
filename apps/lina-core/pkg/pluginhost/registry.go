package pluginhost

import "sync"

var (
	sourcePluginRegistryMu sync.RWMutex
	sourcePluginRegistry   = make(map[string]*SourcePlugin)
)

// SourcePlugin defines one compile-time source plugin contribution.
type SourcePlugin struct {
	// ID is the stable plugin id and must match `plugin.yaml`.
	ID string
	// Hooks contains hook declarations contributed by the source plugin.
	Hooks []*HookSpec
	// Resources contains backend resource declarations contributed by the source plugin.
	Resources []*ResourceSpec
}

// HookSpec defines one hook declaration contributed by a source plugin.
type HookSpec struct {
	// Event is the published host hook event name.
	Event HookSlot
	// Action is the hook execution type. Current host only supports `insert`.
	Action HookAction
	// Table is the target table name for the hook action.
	Table string
	// Fields maps database column names to host event expressions.
	Fields map[string]string
}

// ResourceSpec defines one backend resource declaration contributed by a source plugin.
type ResourceSpec struct {
	// Key is the stable resource key used by the host API path.
	Key string
	// Type is the resource type. Current host supports `table-list`.
	Type string
	// Table is the queried table name.
	Table string
	// Fields defines selected output fields.
	Fields []*ResourceField
	// Filters defines supported query filters.
	Filters []*ResourceFilter
	// OrderBy defines the default ordering.
	OrderBy *OrderBySpec
}

// ResourceField defines one backend resource output field.
type ResourceField struct {
	// Name is the response field name.
	Name string
	// Column is the underlying database column name.
	Column string
}

// ResourceFilter defines one backend resource query filter.
type ResourceFilter struct {
	// Param is the incoming query parameter name.
	Param string
	// Column is the underlying database column name.
	Column string
	// Operator is the supported filter operator.
	Operator string
}

// OrderBySpec defines one backend resource order-by declaration.
type OrderBySpec struct {
	// Column is the ordered database column.
	Column string
	// Direction is the order direction, for example `asc` or `desc`.
	Direction string
}

// RegisterSourcePlugin registers one compile-time source plugin into the host registry.
func RegisterSourcePlugin(plugin *SourcePlugin) {
	if plugin == nil {
		panic("pluginhost: source plugin is nil")
	}
	if plugin.ID == "" {
		panic("pluginhost: source plugin id is empty")
	}

	sourcePluginRegistryMu.Lock()
	defer sourcePluginRegistryMu.Unlock()

	if _, ok := sourcePluginRegistry[plugin.ID]; ok {
		panic("pluginhost: duplicate source plugin registration: " + plugin.ID)
	}
	sourcePluginRegistry[plugin.ID] = plugin
}

// GetSourcePlugin returns one registered compile-time source plugin by id.
func GetSourcePlugin(id string) (*SourcePlugin, bool) {
	sourcePluginRegistryMu.RLock()
	defer sourcePluginRegistryMu.RUnlock()

	plugin, ok := sourcePluginRegistry[id]
	return plugin, ok
}
