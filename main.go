package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

var (
	config     Config
	namespaces map[string]*NSConfig
)

type NSConfig struct {
	Alias     string `yaml:"alias"`
	Namespace string `yaml:"namespace"`
	Token     string `yaml:"token"`
}

type Config struct {
	BindAddr           string     `yaml:"bind_addr"`
	DirektivAddr       string     `yaml:"direktiv_addr"`
	InsecureSkipVerify bool       `yaml:"insecure_skip_verify"`
	Namespaces         []NSConfig `yaml:"routes"`
}

func loadConfig() error {
	if len(os.Args) < 2 {
		return errors.New("usage: ./dproxy CONFIG_FILE")
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	if len(config.Namespaces) == 0 {
		return errors.New("config file defines no routes")
	}

	if len(config.DirektivAddr) == 0 {
		return errors.New("config file defines no direktiv_addr")
	}

	namespaces = make(map[string]*NSConfig)

	for idx := range config.Namespaces {
		nsconf := &config.Namespaces[idx]
		key := nsconf.Alias
		if key == "" {
			key = nsconf.Namespace
		}
		namespaces[key] = nsconf
	}

	return nil
}

func fail(err error) {
	fmt.Println(err.Error())
	os.Exit(1)
}

func main() {
	err := loadConfig()
	if err != nil {
		fail(err)
		return
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/dproxy/n/:namespace/w/:workflow", handler)

	fmt.Printf("Listening on %s\n", config.BindAddr)

	router.Run(config.BindAddr)
}

func handler(c *gin.Context) {
	ns := c.Param("namespace")
	wf := c.Param("workflow")
	ctype := c.Query("ctype")
	field := c.Query("field")
	rawOutput := c.Query("raw-output")

	nsconf, exists := namespaces[ns]
	if !exists {
		fmt.Printf("Aborting handler for unknown namespace '%s'.\n", ns)
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("namespace '%s' not found", ns)})
		return
	}

	fmt.Printf("Handling request for namespace '%s' and workflow '%s'.\n", ns, wf)
	url := fmt.Sprintf("https://%s/api/namespaces/%s/tree/%s?op=wait", config.DirektivAddr, nsconf.Namespace, wf)
	if ctype != "" {
		url += fmt.Sprintf("&ctype=%s", ctype)
	}
	if field != "" {
		url += fmt.Sprintf("&field=%s", field)
	}
	if rawOutput != "" {
		url += fmt.Sprintf("&raw-output=%s", rawOutput)
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: config.InsecureSkipVerify}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Printf("Aborting handler after failing to craft request: %v\n", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	req.Header.Add("Direktiv-Token", nsconf.Token)

	fmt.Printf("Handling request: %s\n", url)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Aborting handler after failing to execute request: %v\n", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	code := resp.StatusCode
	ctype = resp.Header.Get("Content-Type")
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Aborting handler after failing to receive response: %v\n", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Data(code, ctype, data)
}
