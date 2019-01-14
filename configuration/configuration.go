package configuration

import (
	"fmt"
	"os"
	"strings"
	"time"

	errs "github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	// Constants for viper variable names. Will be used to set
	// default values as well as to get each value
	varCleanTestDataEnabled = "clean.test.data"
	varDBLogsEnabled        = "enable.db.logs"
	varDeveloperModeEnabled = "developer.mode.enabled"
	varDiagnoseHTTPAddress  = "diagnose.http.address"
	varEnvironment          = "environment"
	varHTTPAddress          = "http.address"
	varLogJSON              = "log.json"
	varLogLevel             = "log.level"
	varMetricsHTTPAddress   = "metrics.http.address"
	varSentryDSN            = "sentry.dsn"

	// Postgres
	varPostgresHost                 = "postgres.host"
	varPostgresPort                 = "postgres.port"
	varPostgresUser                 = "postgres.user"
	varPostgresDatabase             = "postgres.database"
	varPostgresPassword             = "postgres.password"
	varPostgresSSLMode              = "postgres.sslmode"
	varPostgresConnectionTimeout    = "postgres.connection.timeout"
	varPostgresTransactionTimeout   = "postgres.transaction.timeout"
	varPostgresConnectionRetrySleep = "postgres.connection.retrysleep"
	varPostgresConnectionMaxIdle    = "postgres.connection.maxidle"
	varPostgresConnectionMaxOpen    = "postgres.connection.maxopen"
	varMonitorIPDuration            = "monitor.ip.duration"

	// ProxyURL
	varProxyURL = "proxy.url"
)

// New creates a configuration reader object using a configurable configuration
// file path.
func New(configFilePath string) (*Config, error) {
	c := Config{
		v: viper.New(),
	}
	c.v.SetEnvPrefix("F8")
	c.v.AutomaticEnv()
	c.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.v.SetTypeByDefaultValue(true)
	c.setConfigDefaults()

	if configFilePath != "" {
		c.v.SetConfigType("yaml")
		c.v.SetConfigFile(configFilePath)
		err := c.v.ReadInConfig() // Find and read the config file
		if err != nil {           // Handle errors reading the config file
			return nil, errs.Errorf("Fatal error config file: %s \n", err)
		}
	}
	return &c, nil
}

// Config encapsulates the Viper configuration registry which stores the
// configuration data in-memory.
type Config struct {
	v *viper.Viper
}

// GetConfig is a wrapper over NewConfigurationData which reads configuration file path
// from the environment variable.
func GetConfig() (*Config, error) {
	return New(getMainConfigFile())
}

func getMainConfigFile() string {
	// This was either passed as a env var or set inside main.go from --config
	envConfigPath, _ := os.LookupEnv("WEBHOOK_CONFIG_FILE_PATH")
	return envConfigPath
}

func (c *Config) setConfigDefaults() {
	c.v.SetTypeByDefaultValue(true)

	c.v.SetDefault(varLogLevel, defaultLogLevel)
	c.v.SetDefault(varHTTPAddress, defaultHTTPAddress)
	c.v.SetDefault(varMetricsHTTPAddress, defaultMetricsHTTPAddress)
	c.v.SetDefault(varDeveloperModeEnabled,
		defaultDeveloperModeEnabled)
	c.v.SetDefault(varCleanTestDataEnabled, true)
	c.v.SetDefault(varDBLogsEnabled, false)

	//---------
	// Postgres
	//---------
	c.v.SetDefault(varPostgresHost, defaultPostgresHost)
	c.v.SetDefault(varPostgresPort, defaultPostgresPort)
	c.v.SetDefault(varPostgresUser, defaultPostgresUser)
	c.v.SetDefault(varPostgresDatabase, defaultPostgresDatabase)
	c.v.SetDefault(varPostgresPassword, defaultPostgresPassword)
	c.v.SetDefault(varPostgresSSLMode, defaultPostgresSSLMode)
	c.v.SetDefault(varPostgresConnectionTimeout,
		defaultPostgresConnectionTimeout)
	c.v.SetDefault(varPostgresConnectionMaxIdle,
		defaultPostgresConnectionMaxIdle)
	c.v.SetDefault(varPostgresConnectionMaxOpen,
		defaultPostgresConnectionMaxOpen)
	// Number of seconds to wait before trying to connect again
	c.v.SetDefault(varPostgresConnectionRetrySleep,
		defaultPostgresConnectionRetrySleep)

	// Timeout of a transaction in minutes
	c.v.SetDefault(varPostgresTransactionTimeout,
		defaultPostgresConnectionTimeout)

	// ProxyURL to forward webhook request
	c.v.SetDefault(varProxyURL, defaultProxyURL)
	// Monitor IP Duration for duration between job to update IP
	c.v.SetDefault(varMonitorIPDuration, defaultMonitorIPDuration)
}

// DeveloperModeEnabled returns `true` if development related features (as set via default, config file, or environment variable),
// e.g. token generation endpoint are enabled
func (c *Config) DeveloperModeEnabled() bool {
	return c.v.GetBool(varDeveloperModeEnabled)
}

// GetEnvironment returns the current environment application is deployed in
// like 'production', 'prod-preview', 'local', etc as the value of environment variable
// `F8_ENVIRONMENT` is set.
func (c *Config) GetEnvironment() string {
	if c.v.IsSet(varEnvironment) {
		return c.v.GetString(varEnvironment)
	}
	return "local"
}

// IsLogJSON returns if we should log json format (as set via config file or environment variable)
func (c *Config) IsLogJSON() bool {
	if c.v.IsSet(varLogJSON) {
		return c.v.GetBool(varLogJSON)
	}
	if c.DeveloperModeEnabled() {
		return false
	}
	return true
}

// GetHTTPAddress returns the HTTP address (as set via default, config file, or environment variable)
// that the wit server binds to (e.g. "0.0.0.0:8080")
func (c *Config) GetHTTPAddress() string {
	return c.v.GetString(varHTTPAddress)
}

// GetMetricsHTTPAddress returns the address the /metrics endpoing will be mounted.
// By default GetMetricsHTTPAddress is the same as GetHTTPAddress
func (c *Config) GetMetricsHTTPAddress() string {
	return c.v.GetString(varMetricsHTTPAddress)
}

// GetDiagnoseHTTPAddress returns the address of where to start the gops handler.
// By default GetDiagnoseHTTPAddress is 127.0.0.1:0 in devMode, but turned off in prod mode
// unless explicitly configured
func (c *Config) GetDiagnoseHTTPAddress() string {
	if c.v.IsSet(varDiagnoseHTTPAddress) {
		return c.v.GetString(varDiagnoseHTTPAddress)
	} else if c.DeveloperModeEnabled() {
		return "127.0.0.1:0"
	}
	return ""
}

// GetLogLevel returns the loggging level (as set via config file or environment variable)
func (c *Config) GetLogLevel() string {
	return c.v.GetString(varLogLevel)
}

// GetPostgresHost returns the postgres host as set via default, config file, or environment variable
func (c *Config) GetPostgresHost() string {
	return c.v.GetString(varPostgresHost)
}

// GetPostgresPort returns the postgres port as set via default, config file, or environment variable
func (c *Config) GetPostgresPort() int64 {
	return c.v.GetInt64(varPostgresPort)
}

// GetPostgresUser returns the postgres user as set via default, config file, or environment variable
func (c *Config) GetPostgresUser() string {
	return c.v.GetString(varPostgresUser)
}

// GetPostgresDatabase returns the postgres database as set via default, config file, or environment variable
func (c *Config) GetPostgresDatabase() string {
	return c.v.GetString(varPostgresDatabase)
}

// GetPostgresPassword returns the postgres password as set via default, config file, or environment variable
func (c *Config) GetPostgresPassword() string {
	return c.v.GetString(varPostgresPassword)
}

// GetPostgresSSLMode returns the postgres sslmode as set via default, config file, or environment variable
func (c *Config) GetPostgresSSLMode() string {
	return c.v.GetString(varPostgresSSLMode)
}

// GetPostgresConnectionTimeout returns the postgres connection timeout as set via default, config file, or environment variable
func (c *Config) GetPostgresConnectionTimeout() int64 {
	return c.v.GetInt64(varPostgresConnectionTimeout)
}

// GetPostgresConnectionRetrySleep returns the number of seconds (as set via default, config file, or environment variable)
// to wait before trying to connect again
func (c *Config) GetPostgresConnectionRetrySleep() time.Duration {
	return c.v.GetDuration(varPostgresConnectionRetrySleep)
}

// GetPostgresTransactionTimeout returns the number of minutes to timeout a transaction
func (c *Config) GetPostgresTransactionTimeout() time.Duration {
	return c.v.GetDuration(varPostgresTransactionTimeout)
}

// GetPostgresConnectionMaxIdle returns the number of connections that should be keept alive in the database connection pool at
// any given time. -1 represents no restrictions/default behavior
func (c *Config) GetPostgresConnectionMaxIdle() int {
	return c.v.GetInt(varPostgresConnectionMaxIdle)
}

// GetPostgresConnectionMaxOpen returns the max number of open connections that should be open in the database connection pool.
// -1 represents no restrictions/default behavior
func (c *Config) GetPostgresConnectionMaxOpen() int {
	return c.v.GetInt(varPostgresConnectionMaxOpen)
}

// GetPostgresConfigString returns a ready to use string for usage in sql.Open()
func (c *Config) GetPostgresConfigString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		c.GetPostgresHost(),
		c.GetPostgresPort(),
		c.GetPostgresUser(),
		c.GetPostgresPassword(),
		c.GetPostgresDatabase(),
		c.GetPostgresSSLMode(),
		c.GetPostgresConnectionTimeout(),
	)
}

// GetProxyURL returns URL to forward Webhook
func (c *Config) GetProxyURL() string {
	return c.v.GetString(varProxyURL)
}

// GetMonitorIPDuration Return duration between monitoring call to
// update ip ranges from source
func (c *Config) GetMonitorIPDuration() time.Duration {
	return c.v.GetDuration(varMonitorIPDuration)
}

// IsCleanTestDataEnabled return true if we want to have clean data enabled
func (c *Config) IsCleanTestDataEnabled() bool {
	return c.v.GetBool(varCleanTestDataEnabled)
}

func (c *Config) GetSentryDSN() string {
	return c.v.GetString(varSentryDSN)
}

// Return True if we want to have DB Logs Enabled
func (c *Config) IsDBLogsEnabled() bool {
	return c.v.GetBool(varDBLogsEnabled)
}
