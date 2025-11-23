package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds application configuration
type Config struct {
	Trainer   TrainerConfig   `mapstructure:"trainer"`
	Shifter   ShifterConfig   `mapstructure:"shifter"`
	Bike      BikeConfig      `mapstructure:"bike"`
	Bluetooth BluetoothConfig `mapstructure:"bluetooth"`
	Display   DisplayConfig   `mapstructure:"display"`
	Controls  ControlsConfig  `mapstructure:"controls"`
}

// BluetoothConfig holds Bluetooth connection settings
type BluetoothConfig struct {
	TrainerAddress string `mapstructure:"trainer_address"`
}

type TrainerConfig struct {
	DeviceID string `mapstructure:"device_id"`
}

type ShifterConfig struct {
	DeviceID string `mapstructure:"device_id"`
}

type BikeConfig struct {
	Preset             string  `mapstructure:"preset"`
	Chainrings         []int   `mapstructure:"chainrings"`
	Cassette           []int   `mapstructure:"cassette"`
	WheelCircumference float64 `mapstructure:"wheel_circumference"`
	RiderWeight        float64 `mapstructure:"rider_weight"`
}

type DisplayConfig struct {
	GraphWindowMinutes      int     `mapstructure:"graph_window_minutes"`
	ClimbGradientThreshold  float64 `mapstructure:"climb_gradient_threshold"`
	ClimbElevationThreshold float64 `mapstructure:"climb_elevation_threshold"`
}

type ControlsConfig struct {
	ShiftUp        string `mapstructure:"shift_up"`
	ShiftDown      string `mapstructure:"shift_down"`
	ResistanceUp   string `mapstructure:"resistance_up"`
	ResistanceDown string `mapstructure:"resistance_down"`
	Pause          string `mapstructure:"pause"`
	ToggleView     string `mapstructure:"toggle_view"`
}

// Load reads config from file with defaults
func Load(configDir string) (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(configDir)

	// Set defaults
	setDefaults(v)

	// Try to read config file (ignore if not found)
	_ = v.ReadInConfig()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	// Bike defaults
	v.SetDefault("bike.preset", "road-2x11")
	v.SetDefault("bike.chainrings", []int{50, 34})
	v.SetDefault("bike.cassette", []int{11, 12, 13, 14, 15, 17, 19, 21, 24, 28})
	v.SetDefault("bike.wheel_circumference", 2.1)
	v.SetDefault("bike.rider_weight", 75.0)

	// Display defaults
	v.SetDefault("display.graph_window_minutes", 5)
	v.SetDefault("display.climb_gradient_threshold", 3.0)
	v.SetDefault("display.climb_elevation_threshold", 30.0)

	// Controls defaults
	v.SetDefault("controls.shift_up", "Up")
	v.SetDefault("controls.shift_down", "Down")
	v.SetDefault("controls.resistance_up", "Right")
	v.SetDefault("controls.resistance_down", "Left")
	v.SetDefault("controls.pause", "Space")
	v.SetDefault("controls.toggle_view", "Tab")
}

// DefaultConfigDir returns the default config directory
func DefaultConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "goc")
}

// Save writes config to file
func Save(cfg *Config, configDir string) error {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	v := viper.New()
	v.SetConfigType("toml")

	v.Set("trainer.device_id", cfg.Trainer.DeviceID)
	v.Set("shifter.device_id", cfg.Shifter.DeviceID)
	v.Set("bluetooth.trainer_address", cfg.Bluetooth.TrainerAddress)
	v.Set("bike.preset", cfg.Bike.Preset)
	v.Set("bike.chainrings", cfg.Bike.Chainrings)
	v.Set("bike.cassette", cfg.Bike.Cassette)
	v.Set("bike.wheel_circumference", cfg.Bike.WheelCircumference)
	v.Set("bike.rider_weight", cfg.Bike.RiderWeight)
	v.Set("display.graph_window_minutes", cfg.Display.GraphWindowMinutes)
	v.Set("display.climb_gradient_threshold", cfg.Display.ClimbGradientThreshold)
	v.Set("display.climb_elevation_threshold", cfg.Display.ClimbElevationThreshold)
	v.Set("controls.shift_up", cfg.Controls.ShiftUp)
	v.Set("controls.shift_down", cfg.Controls.ShiftDown)
	v.Set("controls.resistance_up", cfg.Controls.ResistanceUp)
	v.Set("controls.resistance_down", cfg.Controls.ResistanceDown)
	v.Set("controls.pause", cfg.Controls.Pause)
	v.Set("controls.toggle_view", cfg.Controls.ToggleView)

	configPath := filepath.Join(configDir, "config.toml")
	return v.WriteConfigAs(configPath)
}
