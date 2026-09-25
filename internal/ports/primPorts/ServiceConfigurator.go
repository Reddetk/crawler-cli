package primports

import "time"

type ServiceConfigurator interface {
	ConfigureService(depth int, reqTimeout time.Duration) error
}
