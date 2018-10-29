package main

import (
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"./shared"
)

func reload(
	endpointFiles map[string]shared.EndpointFile,
	vHostsPath string,
	tcpVHostsPath string,
	certScript string,
) error {
	err := rebuildConfigs(
		endpointFiles,
		vHostsPath,
		tcpVHostsPath,
		certScript,
	)
	if err != nil {
		return err
	}
	// Run Cert-Script
	println("Creating / Updating certificates")
	err = shared.Run(exec.Command("sh", certScript))
	if err != nil {
		return err
	}
	return reloadNginx()
}

func reloadNginx() error {
	// Note: If the config-file is invalid, 'reload' fails silently.
	// We check if the config-file is valid before reloading so we
	// get an error-message
	println("Reloading Nginx")
	err := shared.Run(exec.Command("nginx", "-t"))
	if err == nil {
		err = shared.Run(exec.Command("nginx", "-s", "reload"))
	}
	return err
}

func cleanConfigs(vHostsPath string, tcpVHostsPath string) error {
	err := shared.CleanDir(vHostsPath, ".conf")
	if err == nil {
		err = shared.CleanDir(tcpVHostsPath, ".conf")
	}
	return err
}

func waitForHttpPort() error {
	var tries = 10
	for {
		conn, err := net.Dial("tcp", "localhost:80")
		if err == nil {
			conn.Close()
			return nil
		}
		if tries--; tries == 0 {
			return errors.New("Timeout waiting for http socket")
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func rebuildConfigs(
	endpointFiles map[string]shared.EndpointFile,
	vHostsPath string,
	tcpVHostsPath string,
	certScript string,
) error {
	println("Rebuilding config files")

	// HTTPS V-Hosts
	err := shared.ProcessDir(
		"template.https.conf",
		vHostsPath,
		".conf",
		endpointFiles,
		func(endpoint shared.EndpointDef) bool {
			return endpoint.Protocol == "https"
		})
	if err != nil {
		return err
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
		return err
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
		return err
	}
	return os.Chmod(certScript, 755)
}

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

	// Start nginx
	if len(os.Args) == 1 {
		// There might be configs using certs we don't have any more.
		// Remove all manage configs
		println("Removing managed configs")
		cleanConfigs(vHostsPath, tcpVHostsPath)

		// Start nginx without managed configs so it can be used as webroot
		quit := make(chan struct{})
		err = shared.RunWithCallback(exec.Command("nginx", "-g", "daemon off;"), func() error {
			err := waitForHttpPort()
			if err != nil {
				return err
			}
			println("Nginx started without managed configs")
			err = reload(
				endpointFiles,
				vHostsPath,
				tcpVHostsPath,
				certScript,
			)
			if err != nil {
				return err
			}
			println("Nginx reloaded with managed configs")
			ticker := time.NewTicker(24 * 7 * time.Hour)
			go func() {
				for {
					select {
					case <-ticker.C:
						println("Renewing certs")
						err = shared.Run(exec.Command("certbot", "renew"))
						if err == nil {
							err = reloadNginx()
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
			return nil
		})
		close(quit)
	} else if os.Args[1] == "build" {
		err = rebuildConfigs(
			endpointFiles,
			vHostsPath,
			tcpVHostsPath,
			certScript,
		)
	} else if os.Args[1] == "wait-port" {
		err = waitForHttpPort()
	} else if os.Args[1] == "clean" {
		cleanConfigs(vHostsPath, tcpVHostsPath)
	} else if os.Args[1] == "get" {
		var args = []string{
			"certonly",
			"--agree-tos",
			"--non-interactive",
			"--webroot",
			"-w",
			"/webroot",
		}

		for i := 2; i < len(os.Args); i++ {
			str := os.Args[i]
			if strings.Contains(str, "@") {
				args = append(args, "--email", str)
			} else {
				args = append(args, "-d", str)
			}
		}
		err = shared.Run(exec.Command(
			"certbot",
			args...))
	} else if os.Args[1] == "reload" {
		err = reload(
			endpointFiles,
			vHostsPath,
			tcpVHostsPath,
			certScript,
		)
	}
	if err != nil {
		panic(err)
	}
}
