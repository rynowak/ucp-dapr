package reconciler

var workflowsByOperationType = map[string]string{
	"APPLICATIONS.CORE/CONTAINERS|PUT":    "ContainerPut",
	"APPLICATIONS.CORE/CONTAINERS|DELETE": "ContainerDelete",
}
