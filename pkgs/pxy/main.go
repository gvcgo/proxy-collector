package main

import (
	"os"

	"github.com/gvcgo/proxy-collector/pkgs/confs"
	"github.com/gvcgo/proxy-collector/pkgs/versions"
)

func main() {
	// sites.TestEDomains()
	// sites.TestEDCollector()
	// sites.TestTDomains()

	// app := NewApp()
	// app.Run()

	cfg := confs.NewCollectorConf()
	// gl := versions.NewGolang(cfg)
	// gl.FetchAll()
	// gl.Upload()

	// jdk := versions.NewJDK(cfg)
	// jdk.FetchAll()
	// jdk.Upload()

	os.Setenv(confs.ToEnableProxyEnvName, "true")
	// gra := versions.NewGradle(cfg)
	// gra.FetchAll()
	// gra.Upload()

	// maven := versions.NewMaven(cfg)
	// maven.FetchAll()
	// maven.Upload()

	// node := versions.NewNodejs(cfg)
	// node.FetchAll()
	// node.Upload()

	// zig := versions.NewZig(cfg)
	// zig.FetchAll()
	// zig.Upload()

	// julia := versions.NewJulia(cfg)
	// julia.FetchAll()
	// julia.Upload()

	flutter := versions.NewFlutter(cfg)
	flutter.FetchAll()
	flutter.Upload()
}
