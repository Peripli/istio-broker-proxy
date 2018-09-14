package integration

// These structures are extracts from the Kube API
// We are using these structures because importing "k8s.io/api/core/v1" will lead to panics related to
// flag redefined: log_dir

type PodList struct {
	Items []Pod `json:"items" protobuf:"bytes,2,rep,name=items"`
}

type Pod struct {
	ObjectMeta `json:"metadata,omitempty"`
}

type ObjectMeta struct {
	Name string `json:"name,omitempty"`
}
