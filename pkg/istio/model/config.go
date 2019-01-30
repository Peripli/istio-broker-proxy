package model

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"time"
)

// ProtoSchema provides description of the configuration schema and its key function
// nolint: maligned
type ProtoSchema struct {
	// ClusterScoped is true for resource in cluster-level.
	ClusterScoped bool

	// Type is the config proto type.
	Type string

	// Plural is the type in plural.
	Plural string

	// Group is the config proto group.
	Group string

	// Version is the config proto version.
	Version string

	// MessageName refers to the protobuf message type name corresponding to the type
	MessageName string

	// Validate configuration as a protobuf message assuming the object is an
	// instance of the expected message type
	Validate func(name, namespace string, config proto.Message) error
}

var (
	// VirtualService describes v1alpha3 route rules
	VirtualService = ProtoSchema{
		Type:        "virtual-service",
		Plural:      "virtual-services",
		Group:       "networking",
		Version:     "v1alpha3",
		MessageName: "istio.networking.v1alpha3.VirtualService",
		Validate:    nil,
	}

	// Gateway describes a gateway (how a proxy is exposed on the network)
	Gateway = ProtoSchema{
		Type:        "gateway",
		Plural:      "gateways",
		Group:       "networking",
		Version:     "v1alpha3",
		MessageName: "istio.networking.v1alpha3.Gateway",
		Validate:    ValidateGateway,
	}

	// ServiceEntry describes service entries
	ServiceEntry = ProtoSchema{
		Type:        "service-entry",
		Plural:      "service-entries",
		Group:       "networking",
		Version:     "v1alpha3",
		MessageName: "istio.networking.v1alpha3.ServiceEntry",
		Validate:    nil,
	}

	// DestinationRule describes destination rules
	DestinationRule = ProtoSchema{
		Type:        "destination-rule",
		Plural:      "destination-rules",
		Group:       "networking",
		Version:     "v1alpha3",
		MessageName: "istio.networking.v1alpha3.DestinationRule",
		Validate:    nil,
	}
)

// ConfigMeta is metadata attached to each configuration unit.
// The revision is optional, and if provided, identifies the
// last update operation on the object.
type ConfigMeta struct {
	// Type is a short configuration name that matches the content message type
	// (e.g. "route-rule")
	Type string `json:"type,omitempty"`

	// Group is the API group of the config.
	Group string `json:"group,omitempty"`

	// Version is the API version of the Config.
	Version string `json:"version,omitempty"`

	// Name is a unique immutable identifier in a namespace
	Name string `json:"name,omitempty"`

	// Namespace defines the space for names (optional for some types),
	// applications may choose to use namespaces for a variety of purposes
	// (security domains, fault domains, organizational domains)
	Namespace string `json:"namespace,omitempty"`

	// Domain defines the suffix of the fully qualified name past the namespace.
	// Domain is not a part of the unique key unlike name and namespace.
	Domain string `json:"domain,omitempty"`

	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects.
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	Annotations map[string]string `json:"annotations,omitempty"`

	// ResourceVersion is an opaque identifier for tracking updates to the config registry.
	// The implementation may use a change index or a commit log for the revision.
	// The config client should not make any assumptions about revisions and rely only on
	// exact equality to implement optimistic concurrency of read-write operations.
	//
	// The lifetime of an object of a particular revision depends on the underlying data store.
	// The data store may compactify old revisions in the interest of storage optimization.
	//
	// An empty revision carries a special meaning that the associated object has
	// not been stored and assigned a revision.
	ResourceVersion string `json:"resourceVersion,omitempty"`

	// CreationTimestamp records the creation time
	CreationTimestamp time.Time `json:"creationTimestamp,omitempty"`
}

// Config is a configuration unit consisting of the type of configuration, the
// key identifier that is unique per type, and the content represented as a
// protobuf message.
type Config struct {
	ConfigMeta

	// Spec holds the configuration object as a protobuf message
	Spec proto.Message
}

// Port represents a network port where a service is listening for
// connections. The port should be annotated with the type of protocol
// used by the port.
type Port struct {
	// Name ascribes a human readable name for the port object. When a
	// service has multiple ports, the name field is mandatory
	Name string `json:"name,omitempty"`

	// Port number where the service can be reached. Does not necessarily
	// map to the corresponding port numbers for the instances behind the
	// service. See NetworkEndpoint definition below.
	Port int `json:"port"`

	// Protocol to be used for the port.
	Protocol Protocol `json:"protocol,omitempty"`
}

// PortList is a set of ports
type PortList []*Port

// Protocol defines network protocols for ports
type Protocol string
