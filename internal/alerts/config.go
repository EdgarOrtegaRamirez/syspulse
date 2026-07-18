// Package alerts provides alert configuration and threshold checking.
package alerts

// Config holds alert thresholds for health checks.
type Config struct {
	CPU           int     // CPU usage percentage threshold
	Memory        int     // Memory usage percentage threshold
	Disk          int     // Disk usage percentage threshold
	Load          float64 // Load average ratio to cores threshold (default: 2.0)
	ProcessZombie int     // Maximum running processes before warning (default: 0 = disabled)
}
