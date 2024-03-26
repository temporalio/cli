package temporalcli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/temporalio/cli/temporalcli/internal/printer"
)

func (c *TemporalCloudLoginCommand) run(cctx *CommandContext, _ []string) error {
	// Set defaults
	if c.Domain == "" {
		c.Domain = "https://login.tmprl.cloud"
	}
	if c.Audience == "" {
		c.Audience = "https://saas-api.tmprl.cloud"
	}
	if c.ClientId == "" {
		c.ClientId = "d7V5bZMLCbRLfRVpqC567AqjAERaWHhl"
	}

	// Get device code
	var codeResp CloudOAuthDeviceCodeResponse
	err := c.postToLogin(
		cctx,
		"/oauth/device/code",
		url.Values{"client_id": {c.ClientId}, "scope": {"openid profile user"}, "audience": {c.Audience}},
		&codeResp,
	)
	if err != nil {
		return fmt.Errorf("failed getting device code: %w", err)
	}

	// Confirm URL same as domain URL
	if domainURL, err := url.Parse(c.Domain); err != nil {
		return fmt.Errorf("failed parsing domain URL: %w", err)
	} else if verifURL, err := url.Parse(codeResp.VerificationURI); err != nil {
		return fmt.Errorf("failed parsing verification URL: %w", err)
	} else if domainURL.Hostname() != verifURL.Hostname() {
		return fmt.Errorf("domain URL %q does not match verification URL %q in response",
			domainURL.Hostname(), verifURL.Hostname())
	}

	if c.DisablePopUp {
		cctx.Printer.Printlnf("Login via this URL: %v", codeResp.VerificationURIComplete)
	} else {
		cctx.Printer.Printlnf("Attempting to open browser to: %v", codeResp.VerificationURIComplete)
		if err := cctx.openBrowser(codeResp.VerificationURIComplete); err != nil {
			cctx.Logger.Debug("Failed opening browser", "error", err)
			cctx.Printer.Println("Failed opening browser, visit URL manually")
		}
	}

	// According to RFC, we should set a default polling interval if not provided.
	// https://tools.ietf.org/html/draft-ietf-oauth-device-flow-07#section-3.5
	if codeResp.Interval == 0 {
		codeResp.Interval = 10
	}

	// Poll for token
	tokenResp, err := c.pollForToken(cctx, codeResp.DeviceCode, time.Duration(codeResp.Interval)*time.Second)
	if err != nil {
		return fmt.Errorf("failed polling for token response: %w", err)
	}
	if c.NoPersist {
		return cctx.Printer.PrintStructured(tokenResp, printer.StructuredOptions{})
	} else if file := defaultCloudLoginTokenFile(); file == "" {
		return fmt.Errorf("unable to find home directory for token file")
	} else if err := writeCloudLoginTokenFile(file, tokenResp); err != nil {
		return fmt.Errorf("failed writing token file: %w", err)
	}
	cctx.Printer.Println("Login successful")
	return nil
}

func (c *TemporalCloudLogoutCommand) run(cctx *CommandContext, _ []string) error {
	// Set defaults
	if c.Domain == "" {
		c.Domain = "https://login.tmprl.cloud"
	}
	// Delete file then do browser logout
	if file := defaultCloudLoginTokenFile(); file != "" {
		if err := deleteCloudLoginTokenFile(file); err != nil {
			return fmt.Errorf("failed deleting cloud token: %w", err)
		}
	}
	logoutURL := c.Domain + "/v2/logout"
	if c.DisablePopUp {
		cctx.Printer.Printlnf("Logout via this URL: %v", logoutURL)
	} else {
		cctx.Printer.Printlnf("Attempting to open browser to: %v", logoutURL)
		if err := cctx.openBrowser(logoutURL); err != nil {
			cctx.Logger.Debug("Failed opening browser", "error", err)
			cctx.Printer.Println("Failed opening browser, visit URL manually")
		}
	}
	return nil
}

type CloudOAuthDeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type CloudOAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func (c *TemporalCloudLoginCommand) postToLogin(
	cctx *CommandContext,
	path string,
	form url.Values,
	resJSON any,
	allowedStatusCodes ...int,
) error {
	req, err := http.NewRequestWithContext(
		cctx,
		"POST",
		strings.TrimRight(c.Domain, "/")+"/"+strings.TrimLeft(path, "/"),
		strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	} else if resp.StatusCode != 200 && !slices.Contains(allowedStatusCodes, resp.StatusCode) {
		return fmt.Errorf("HTTP call failed, status: %v, body: %s", resp.StatusCode, b)
	}
	return json.Unmarshal(b, resJSON)
}

func (c *TemporalCloudLoginCommand) pollForToken(
	cctx *CommandContext,
	deviceCode string,
	interval time.Duration,
) (*CloudOAuthTokenResponse, error) {
	var tokenResp CloudOAuthTokenResponse
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-cctx.Done():
			return nil, cctx.Err()
		case <-ticker.C:
		}
		err := c.postToLogin(
			cctx,
			"/oauth/token",
			url.Values{
				"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
				"device_code": {deviceCode},
				"client_id":   {c.ClientId},
			},
			&tokenResp,
			// 403 is returned while polling
			http.StatusForbidden,
		)
		if err != nil {
			return nil, err
		} else if len(tokenResp.AccessToken) > 0 {
			return &tokenResp, nil
		}
	}
}
