package cmd

// Version is set at build time via -ldflags="-X rootinfo/cmd.Version=x.y".
var Version = "dev"

// BuildDate is set at build time via -ldflags="-X rootinfo/cmd.BuildDate=YYYY-MM-DD".
var BuildDate = "unknown"
