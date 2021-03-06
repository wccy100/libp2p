package config

import (
	"time"

	"github.com/smallnest/libp2p/log"
)

const (
	ClientVersion    = "libp2p/0.0.1"
	MinClientVersion = "0.0.1"
)

// ConfigValues specifies  default values for node config params.
var (
	ConfigValues      = DefaultConfig()
	TimeConfigValues  = ConfigValues.TimeConfig
	SwarmConfigValues = ConfigValues.SwarmConfig
)

func init() {
	// set default config params based on runtime here
}

func duration(duration string) (dur time.Duration) {
	dur, err := time.ParseDuration(duration)
	if err != nil {
		log.Error("could not parse duration string returning 0, error:", err)
	}
	return dur
}

// Config defines the configuration options for the Spacemesh peer-to-peer networking layer
type Config struct {
	PrivateKey      string        `mapstructure:"private-key"`
	FastSync        bool          `mapstructure:"fast-sync"`
	TCPPort         int           `mapstructure:"tcp-port"`
	DialTimeout     time.Duration `mapstructure:"dial-timeout"`
	ConnKeepAlive   time.Duration `mapstructure:"conn-keepalive"`
	NetworkID       uint32        `mapstructure:"network-id"`
	ResponseTimeout time.Duration `mapstructure:"response-timeout"`
	SwarmConfig     SwarmConfig   `mapstructure:"swarm"`
	TimeConfig      TimeConfig
}

// SwarmConfig specifies swarm config params.
type SwarmConfig struct {
	Gossip                 bool     `mapstructure:"gossip"`
	Bootstrap              bool     `mapstructure:"bootstrap"`
	RoutingTableBucketSize int      `mapstructure:"bucketsize"`
	RoutingTableAlpha      int      `mapstructure:"alpha"`
	RandomConnections      int      `mapstructure:"randcon"`
	BootstrapNodes         []string `mapstructure:"bootnodes"`
}

// TimeConfig specifies the timesync params for ntp.
type TimeConfig struct {
	MaxAllowedDrift       time.Duration `mapstructure:"max-allowed-time-drift"`
	NtpQueries            int           `mapstructure:"ntp-queries"`
	DefaultTimeoutLatency time.Duration `mapstructure:"default-timeout-latency"`
	RefreshNtpInterval    time.Duration `mapstructure:"ntp-refresh-interval"`
}

// DefaultConfig deines the default p2p configuration
func DefaultConfig() Config {

	// TimeConfigValues defines default values for all time and ntp related params.
	var TimeConfigValues = TimeConfig{
		MaxAllowedDrift:       duration("10s"),
		NtpQueries:            5,
		DefaultTimeoutLatency: duration("10s"),
		RefreshNtpInterval:    duration("30m"),
	}

	// SwarmConfigValues defines default values for swarm config params.
	var SwarmConfigValues = SwarmConfig{
		Gossip:                 false,
		Bootstrap:              false,
		RoutingTableBucketSize: 20,
		RoutingTableAlpha:      3,
		RandomConnections:      5,
		BootstrapNodes:         []string{ // these should be the spacemesh foundation bootstrap nodes
		},
	}

	return Config{
		FastSync:        true,
		TCPPort:         7513,
		DialTimeout:     duration("1m"),
		ConnKeepAlive:   duration("48h"),
		NetworkID:       0,
		ResponseTimeout: duration("15s"),
		SwarmConfig:     SwarmConfigValues,
		TimeConfig:      TimeConfigValues,
	}
}
