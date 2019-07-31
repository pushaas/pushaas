package workers

import (
	"fmt"

	"github.com/spf13/viper"
)

type (
	ProvisionWorker interface {
		DispatchWorker()
	}

	provisionWorker struct {
		enabled bool
		workersEnabled bool
	}
)

func (w *provisionWorker) startWorker() {
	fmt.Println("### starting provision worker")
}

func (w *provisionWorker) DispatchWorker() {
	if w.workersEnabled && w.enabled {
		go w.startWorker()
	}
}

func NewProvisionWorker(config *viper.Viper) ProvisionWorker {
	enabled := config.GetBool("workers.provision.enabled")
	workersEnabled := config.GetBool("workers.enabled")

	return &provisionWorker{
		enabled: enabled,
		workersEnabled: workersEnabled,
	}
}
