package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// loadEnvFile loads environment variables from .env file
func loadEnvFile(envPath string) error {
	if envPath == "" {
		envPath = ".env"
	}

	file, err := os.Open(envPath)
	if err != nil {
		// .env file doesn't exist, that's okay
		return nil
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Remove quotes if present
				if len(value) >= 2 {
					if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
						(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
						value = value[1 : len(value)-1]
					}
				}

				os.Setenv(key, value)
			}
		}
	}

	return nil
}

// RTRClient represents a CrowdStrike Real-Time Response client
type RTRClient struct {
	authURL      string
	baseURL      string
	clientID     string
	clientSecret string
	httpClient   *http.Client
	headers      map[string]string
}

// NewRTRClient creates a new RTRClient instance
func NewRTRClient(clientID, clientSecret, baseURL string, verifyCert bool) *RTRClient {
	if baseURL == "" {
		baseURL = "https://api.crowdstrike.com"
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	if !verifyCert {
		// Note: In production, you should properly handle certificate verification
		// This is a simplified version
		client.Transport = &http.Transport{
			// TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return &RTRClient{
		authURL:      baseURL,
		baseURL:      baseURL + "/real-time-response",
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   client,
		headers:      make(map[string]string),
	}
}

// Authenticate authenticates to CrowdStrike API using id and secret
func (c *RTRClient) Authenticate() error {
	payload := url.Values{}
	payload.Set("client_id", c.clientID)
	payload.Set("client_secret", c.clientSecret)

	req, err := http.NewRequest("POST", c.authURL+"/oauth2/token", strings.NewReader(payload.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed: %s", string(body))
	}

	var authResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return err
	}

	c.headers["Authorization"] = "Bearer " + authResp.AccessToken
	c.headers["token_type"] = "bearer"
	c.headers["Content-Type"] = "application/json"

	return nil
}

// HostSearch searches for hosts in your environment - Returns a list of agent IDs
func (c *RTRClient) HostSearch(criteria, criteriaType, rawFilter string, limit int) ([]string, error) {
	reqURL := c.authURL + "/devices/queries/devices/v1"
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	// Build query parameters
	q := req.URL.Query()
	if criteria != "" && criteriaType != "" {
		q.Set("filter", fmt.Sprintf("%s:'%s'", criteriaType, criteria))
	} else if rawFilter != "" {
		q.Set("filter", rawFilter)
	}

	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("host search failed: %s", string(body))
	}

	var result struct {
		Resources []string `json:"resources"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Resources, nil
}

// BatchInit initializes an RTR session across multiple hosts
func (c *RTRClient) BatchInit(hostIDs []string, timeout, timeoutDuration string) (string, error) {
	reqURL := c.baseURL + "/combined/batch-init-session/v1"

	// Build query parameters
	q := url.Values{}
	if timeout != "" {
		q.Set("timeout", timeout)
	}
	if timeoutDuration != "" {
		q.Set("timeout_duration", timeoutDuration)
	}
	if len(q) > 0 {
		reqURL += "?" + q.Encode()
	}

	payload := map[string]interface{}{
		"host_ids": hostIDs,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	// Set headers
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("batch init failed: %s", string(body))
	}

	var result struct {
		BatchID string `json:"batch_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.BatchID, nil
}

// BatchAdminCmd executes an RTR admin command across all hosts mapped to a batch ID
func (c *RTRClient) BatchAdminCmd(batchID, command, commandString string, timeout int, timeoutDuration string, optionalHosts []string) ([]byte, error) {
	reqURL := c.baseURL + "/combined/batch-admin-command/v1"

	// Build query parameters
	q := url.Values{}
	if timeout > 0 {
		q.Set("timeout", fmt.Sprintf("%d", timeout))
	}
	if timeoutDuration != "" {
		q.Set("timeout_duration", timeoutDuration)
	}
	if len(q) > 0 {
		reqURL += "?" + q.Encode()
	}

	payload := map[string]interface{}{
		"base_command":  command,
		"batch_id":      batchID,
		"command_string": commandString,
	}

	if len(optionalHosts) > 0 {
		payload["optional_hosts"] = optionalHosts
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Set headers
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func runcmd(rtrClient *RTRClient, host string, script string, wg *sync.WaitGroup) {
	defer wg.Done()

	hosts := []string{host}
	sessionID, err := rtrClient.BatchInit(hosts, "30", "30s")
	if err != nil {
		fmt.Printf("Error initializing batch for host %s: %v\n", host, err)
		return
	}

	cmd := "runscript -Raw=```" + script + "```"
	execResult, err := rtrClient.BatchAdminCmd(sessionID, "runscript", cmd, 30, "10m", hosts)
	if err != nil {
		fmt.Printf("Error executing command for host %s: %v\n", host, err)
		return
	}

	stdout := strings.ReplaceAll(string(execResult), "'", `"`)
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &data); err != nil {
		return
	}

	// Try to extract stdout from the nested structure
	if combined, ok := data["combined"].(map[string]interface{}); ok {
		if resources, ok := combined["resources"].(map[string]interface{}); ok {
			if hostData, ok := resources[host].(map[string]interface{}); ok {
				if stdout, ok := hostData["stdout"].(string); ok {
					fmt.Println(stdout)
				}
			}
		}
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: cli <hostname> <script>")
		os.Exit(1)
	}

	// Load environment variables from .env file
	if err := loadEnvFile(".env"); err != nil {
		fmt.Printf("Warning: Could not load .env file: %v\n", err)
	}

	clientID := os.Getenv("CLIENT_ID")
	apiKey := os.Getenv("CLIENT_SECRET")

	if clientID == "" || apiKey == "" {
		fmt.Println("Error: CLIENT_ID and CLIENT_SECRET must be set in .env file or environment variables")
		os.Exit(1)
	}

	rtrClient := NewRTRClient(clientID, apiKey, "", true)
	if err := rtrClient.Authenticate(); err != nil {
		fmt.Printf("Error authenticating: %v\n", err)
		os.Exit(1)
	}

	hosts, err := rtrClient.HostSearch(os.Args[1], "hostname", "", 5000)
	if err != nil {
		fmt.Printf("Error searching for hosts: %v\n", err)
		os.Exit(1)
	}

	script := os.Args[2]

	// Use goroutines with WaitGroup for parallel execution (similar to ThreadPoolExecutor)
	var wg sync.WaitGroup
	maxWorkers := 32
	semaphore := make(chan struct{}, maxWorkers)

	for _, host := range hosts {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore

		go func(h string) {
			defer func() { <-semaphore }() // Release semaphore
			runcmd(rtrClient, h, script, &wg)
			time.Sleep(200 * time.Millisecond) // Equivalent to time.sleep(0.2)
		}(host)
	}

	wg.Wait()
}

