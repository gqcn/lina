package plugin

import "strings"

// Topology defines the cluster semantics required by plugin runtime behavior.
type Topology interface {
	// IsEnabled reports whether the host is running in clustered mode.
	IsEnabled() bool
	// IsPrimary reports whether the current node is the primary node.
	IsPrimary() bool
	// NodeID returns the stable identifier of the current node.
	NodeID() string
}

type singleNodeTopology struct{}

func (singleNodeTopology) IsEnabled() bool {
	return false
}

func (singleNodeTopology) IsPrimary() bool {
	return true
}

func (singleNodeTopology) NodeID() string {
	return "local-node"
}

func (s *Service) isClusterModeEnabled() bool {
	if s == nil || s.topology == nil {
		return false
	}
	return s.topology.IsEnabled()
}

func (s *Service) isPrimaryNode() bool {
	if s == nil || s.topology == nil {
		return true
	}
	return s.topology.IsPrimary()
}

func (s *Service) currentNodeID() string {
	if s == nil || s.topology == nil {
		return "local-node"
	}

	nodeID := strings.TrimSpace(s.topology.NodeID())
	if nodeID == "" {
		return "local-node"
	}
	return nodeID
}
