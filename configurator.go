package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"./shared"
)

func main() {
	nginxPath, _ := os.LookupEnv("NGINX")
	if len(nginxPath) == 0 {
		nginxPath = "/etc/nginx"
	}
	certbotPath, _ := os.LookupEnv("CERTBOT")
	if len(certbotPath) == 0 {
		certbotPath = "/etc/letsencrypt"
	}
	backendsPath, _ := os.LookupEnv("BACKENDS")
	if len(backendsPath) == 0 {
		backendsPath = "/backends"
	}
	// Since we use this path directly, normalize it
	backendsPath = filepath.Join(backendsPath)
	vHostsPath := filepath.Join(nginxPath, "conf.d")
	tcpVHostsPath := filepath.Join(nginxPath, "stream-conf.d")
	certScript := filepath.Join(certbotPath, "cert.sh")

	println("vHostsPath    ", vHostsPath)
	println("tcpVHostsPath ", tcpVHostsPath)
	println("certScript    ", certScript)
	println("backendsPath  ", backendsPath)

	// Load endpoint-files
	endpointFiles, err := shared.LoadEndpointFiles(backendsPath)
	if err != nil {
		panic(err)
	}

	// HTTPS V-Hosts
	err = shared.ProcessDir(
		"template.https.conf",
		vHostsPath,
		".conf",
		endpointFiles,
		func(endpoint shared.EndpointDef) bool {
			return endpoint.Protocol == "https"
		})
	if err != nil {
		panic(err)
	}

	// Stream V-Hosts
	err = shared.ProcessDir(
		"template.stream.conf",
		tcpVHostsPath,
		".conf",
		endpointFiles,
		func(endpoint shared.EndpointDef) bool {
			return endpoint.Protocol != "https"
		})
	if err != nil {
		panic(err)
	}

	// Cert-Script
	err = shared.ProcessFile(
		"template.cert.sh",
		certScript,
		endpointFiles,
		func(endpoint shared.EndpointDef) bool {
			return true
		})
	if err != nil {
		panic(err)
	}
	os.Chmod(certScript, 755)

	if len(os.Args) != 1 && os.Args[1] == "dry-run" {
		println("Dry-run complete")
		os.Exit(0)
	}

	// Run Cert-Script
	err = shared.Run(exec.Command("sh", certScript))
	if err != nil {
		panic(err)
	}

	// Start nginx
	if len(os.Args) == 1 {
		ticker := time.NewTicker(time.Minute)
		quit := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker.C:
					println("Renewing certs")
					err = shared.Run(exec.Command("certbot", "renew"))
					if err == nil {
						err = shared.Run(exec.Command("nginx", "-t"))
						if err == nil {
							err = shared.Run(exec.Command("nginx", "-s", "reload"))
						}
					}
					if err != nil {
						println("Renewing certs failed", err)
					}
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}()
		err = shared.Run(exec.Command("nginx", "-g", "daemon off;"))
		close(quit)
	} else if os.Args[1] == "dry-run" {
		println("Dry-run complete")
	} else if os.Args[1] == "reload" {
		err = shared.Run(exec.Command("nginx", "-t"))
		if err != nil {
			panic(err)
		}
		err = shared.Run(exec.Command("nginx", "-s", "reload"))
	}
	if err != nil {
		panic(err)
	}
}
