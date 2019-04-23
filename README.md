# pushaas
Pushstream as a Service

## ecs methods

CreateService(*ecs.CreateServiceInput) (*ecs.CreateServiceOutput, error)
CreateTaskSet(*ecs.CreateTaskSetInput) (*ecs.CreateTaskSetOutput, error)
DeleteService(*ecs.DeleteServiceInput) (*ecs.DeleteServiceOutput, error)
DeleteTaskSet(*ecs.DeleteTaskSetInput) (*ecs.DeleteTaskSetOutput, error)
DeregisterContainerInstance(*ecs.DeregisterContainerInstanceInput) (*ecs.DeregisterContainerInstanceOutput, error)
DeregisterTaskDefinition(*ecs.DeregisterTaskDefinitionInput) (*ecs.DeregisterTaskDefinitionOutput, error)
DescribeContainerInstances(*ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error)
DescribeServices(*ecs.DescribeServicesInput) (*ecs.DescribeServicesOutput, error)
DescribeTaskDefinition(*ecs.DescribeTaskDefinitionInput) (*ecs.DescribeTaskDefinitionOutput, error)
DescribeTaskSets(*ecs.DescribeTaskSetsInput) (*ecs.DescribeTaskSetsOutput, error)
DescribeTasks(*ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error)
ListContainerInstances(*ecs.ListContainerInstancesInput) (*ecs.ListContainerInstancesOutput, error)
ListContainerInstancesPages(*ecs.ListContainerInstancesInput, func(*ecs.ListContainerInstancesOutput, bool) bool) error
ListServices(*ecs.ListServicesInput) (*ecs.ListServicesOutput, error)
ListTaskDefinitionFamilies(*ecs.ListTaskDefinitionFamiliesInput) (*ecs.ListTaskDefinitionFamiliesOutput, error)
ListTaskDefinitions(*ecs.ListTaskDefinitionsInput) (*ecs.ListTaskDefinitionsOutput, error)
ListTasks(*ecs.ListTasksInput) (*ecs.ListTasksOutput, error)
RegisterContainerInstance(*ecs.RegisterContainerInstanceInput) (*ecs.RegisterContainerInstanceOutput, error)
RegisterTaskDefinition(*ecs.RegisterTaskDefinitionInput) (*ecs.RegisterTaskDefinitionOutput, error)
RunTask(*ecs.RunTaskInput) (*ecs.RunTaskOutput, error)
StartTask(*ecs.StartTaskInput) (*ecs.StartTaskOutput, error)
StopTask(*ecs.StopTaskInput) (*ecs.StopTaskOutput, error)
WaitUntilServicesInactive(*ecs.DescribeServicesInput) error
WaitUntilServicesStable(*ecs.DescribeServicesInput) error
WaitUntilTasksRunning(*ecs.DescribeTasksInput) error
WaitUntilTasksStopped(*ecs.DescribeTasksInput) error
