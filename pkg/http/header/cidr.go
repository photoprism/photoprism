package header

const (
	// CidrPodInternal covers internal pod traffic ranges.
	CidrPodInternal = "10.0.0.0/8"
	// CidrDockerInternal covers default Docker internal ranges.
	CidrDockerInternal = "172.16.0.0/12"
	// CidrCalicoInternal covers Calico internal ranges.
	CidrCalicoInternal = "192.168.0.0/16"
)
