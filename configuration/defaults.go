package configuration

import "time"

const (
	defaultLogLevel             = "info"
	defaultHTTPAddress          = "0.0.0.0:8080"
	defaultMetricsHTTPAddress   = "0.0.0.0:8080"
	defaultDeveloperModeEnabled = false

	defaultPostgresHost                 = "localhost"
	defaultPostgresPort                 = 5432
	defaultPostgresUser                 = "postgres"
	defaultPostgresDatabase             = "postgres"
	defaultPostgresPassword             = "mysecretpassword"
	defaultPostgresSSLMode              = "disable"
	defaultPostgresConnectionTimeout    = 5
	defaultPostgresTransactionTimeout   = 5 * time.Minute
	defaultPostgresConnectionRetrySleep = time.Second
	defaultPostgresConnectionMaxIdle    = -1
	defaultPostgresConnectionMaxOpen    = -1
	defaultProxyURL                     = "http://localhost:9091"
	defaultMonitorIPDuration            = 15 * time.Minute
)
