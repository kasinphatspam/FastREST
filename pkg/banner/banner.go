package banner

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"fastrest/constant"
)

const art = `
███████╗ █████╗ ███████╗████████╗██████╗ ███████╗███████╗████████╗
██╔════╝██╔══██╗██╔════╝╚══██╔══╝██╔══██╗██╔════╝██╔════╝╚══██╔══╝
█████╗  ███████║███████╗   ██║   ██████╔╝█████╗  ███████╗   ██║
██╔══╝  ██╔══██║╚════██║   ██║   ██╔══██╗██╔══╝  ╚════██║   ██║
██║     ██║  ██║███████║   ██║   ██║  ██║███████╗███████║   ██║
╚═╝     ╚═╝  ╚═╝╚══════╝   ╚═╝   ╚═╝  ╚═╝╚══════╝╚══════╝   ╚═╝
`

type Config struct {
	Addr        string
	HealthCheck bool
	HealthPath  string
	Metrics     bool
	Routes      int
	Env         string
}

func Print(cfg *Config) {
	fmt.Print(constant.ColorCyan)
	fmt.Print(art)
	fmt.Print(constant.ColorReset)

	hostname, _ := os.Hostname()

	env := cfg.Env
	if env == "" {
		env = "development"
	}

	fmt.Println()
	fmt.Printf("  %s%s%s %s\n", constant.ColorGreen, "●", constant.ColorReset, "FastREST server started")
	fmt.Println()

	printItem("Server", cfg.Addr)
	printItem("Environment", env)
	printItem("Routes", fmt.Sprintf("%d", cfg.Routes))
	fmt.Println()

	printItem("Hostname", hostname)
	printItem("OS/Arch", runtime.GOOS+"/"+runtime.GOARCH)
	printItem("Go", runtime.Version())
	printItem("PID", fmt.Sprintf("%d", os.Getpid()))
	printItem("CPUs", fmt.Sprintf("%d", runtime.NumCPU()))
	fmt.Println()

	features := []string{}
	if cfg.HealthCheck {
		features = append(features, "health "+cfg.HealthPath)
	}
	if cfg.Metrics {
		features = append(features, "metrics")
	}
	if len(features) > 0 {
		printItem("Features", strings.Join(features, ", "))
	}

	printItem("Started", time.Now().Format("15:04:05"))
	fmt.Println()
}

func printItem(label, value string) {
	fmt.Printf("  %s%-14s%s %s\n", constant.ColorGray, label, constant.ColorReset, value)
}
