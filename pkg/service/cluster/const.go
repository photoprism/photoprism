package cluster

type NodeType = string

const (
	Portal   NodeType = "portal"   // A Portal server for orchestrating a cluster.
	Instance NodeType = "instance" // An Instance can register with a Portal to join a cluster.
	Service  NodeType = "service"  // Additional Service with computing, sharing, or storage capabilities.
)
