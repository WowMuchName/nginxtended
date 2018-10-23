package shared

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ClientDef holds Client information
type ClientDef struct {
	CommonName string
}

// EndpointDef holds the definition of the endpoint
type EndpointDef struct {
	Version  string
	Domain   string
	URL      string
	Protocol string
	Aliases  []string
	Admin    string
	Port     int16
	Clients  []ClientDef
	KeyAuth  bool
}

// EndpointFile holds information about a Endpoint and the file that contained that information
type EndpointFile struct {
	Fullpath string
	Endpoint EndpointDef
	Name     string
}

func loadEndpoint(path string) (EndpointDef, error) {
	// Load the Endpoint
	endpointBody, err := ioutil.ReadFile(path)
	var endpoint EndpointDef
	if err != nil {
		return endpoint, err
	}
	err = json.Unmarshal(endpointBody, &endpoint)

	// Check required fields
	if len(endpoint.Domain) == 0 {
		return endpoint, errors.New("Domain is a required field")
	}
	if len(endpoint.URL) == 0 {
		return endpoint, errors.New("URL is a required field")
	}

	// Apply defaults
	if len(endpoint.Version) == 0 {
		endpoint.Version = "1.0"
	}
	if len(endpoint.Protocol) == 0 {
		endpoint.Protocol = "https"
	}
	if len(endpoint.Admin) == 0 {
		endpoint.Admin = "admin@" + endpoint.Domain
	}
	if endpoint.Port == 0 {
		endpoint.Port = 443
	}
	if !endpoint.KeyAuth {
		endpoint.KeyAuth = len(endpoint.Clients) > 0
	}

	// Check Client subobject
	for _, client := range endpoint.Clients {
		if len(client.CommonName) == 0 {
			return endpoint, errors.New("Client.CommonName is a required field")
		}
	}

	// Enforce constraints after defaults have been applied
	if endpoint.Version != "1.0" {
		return endpoint, errors.New("Unsupported Version " + endpoint.Version)
	}
	if endpoint.Protocol != "https" && endpoint.Protocol != "tls" {
		return endpoint, errors.New("Unsupported Protocol " + endpoint.Protocol)
	}
	return endpoint, err
}

func LoadEndpointFiles(InputDir string) (map[string]EndpointFile, error) {
	var fileMap = make(map[string]EndpointFile)
	files, err := ioutil.ReadDir(InputDir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(name, ".json") {
			var current EndpointFile
			current.Fullpath = filepath.Join(InputDir, name)
			current.Name = strings.TrimSuffix(name, ".json")
			current.Endpoint, err = loadEndpoint(current.Fullpath)
			if err != nil {
				return nil, err
			}
			fileMap[current.Name] = current
		}
	}
	return fileMap, nil
}

func ProcessDir(
	templatePath string,
	outputDir string,
	extension string,
	endpoints map[string]EndpointFile,
	filter func(endpoint EndpointDef) bool) error {
	template, err := loadTemplate(templatePath)
	if err != nil {
		return err
	}
	files, err := ioutil.ReadDir(outputDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(name, extension) && strings.HasPrefix(name, derivedPrefix) {
			if _, ok := endpoints[strings.TrimPrefix(strings.TrimSuffix(name, extension), derivedPrefix)]; !ok {
				out := filepath.Join(outputDir, name)
				fmt.Println("Remove File", out)
				err = os.Remove(out)
				if err != nil {
					return err
				}
			}
		}
	}
	for name, endpointFile := range endpoints {
		if filter(endpointFile.Endpoint) {
			out := filepath.Join(outputDir, derivedPrefix+name+extension)
			fmt.Println("Transforming Template", endpointFile.Fullpath, "->", out)
			file, err := os.OpenFile(out, os.O_RDWR|os.O_CREATE, 0755)
			file.Truncate(0)
			err = template.Execute(file, endpointFile.Endpoint)
			if err != nil {
				return err
			}
			file.Close()
		}
	}
	return nil
}

func ProcessFile(
	templatePath string,
	outputFile string,
	endpoints map[string]EndpointFile,
	filter func(endpoint EndpointDef) bool) error {
	template, err := loadTemplate(templatePath)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	file.Truncate(0)
	for _, endpointFile := range endpoints {
		if filter(endpointFile.Endpoint) {
			fmt.Println("Transforming Template", endpointFile.Fullpath, "->", outputFile)
			err = template.Execute(file, endpointFile.Endpoint)
			if err != nil {
				return err
			}
		}
	}
	return file.Close()
}

func loadTemplate(path string) (*template.Template, error) {
	templateBody, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	template, err := template.New("Template").Parse(string(templateBody))
	if err != nil {
		panic(err)
	}
	return template, nil
}

func Run(cmd *exec.Cmd) error {
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

const derivedPrefix = "derived_"

