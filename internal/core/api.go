package core

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
)

// RequestConfig holds configurable options for API requests
type RequestConfig struct {
	Method            string            // HTTP method (GET, POST, PUT)
	Headers           map[string]string // Additional headers to merge
	Auth              bool              // Include bearer token
	CommonHeaders     bool              // Include full CommonHeaders vs minimal
	Context           context.Context   // Request context
	StreamingResponse bool              // Return body as stream (caller closes)
	CheckStatus       bool              // Check response status with checkResponse
	ChunkedTransfer   bool              // Enable chunked transfer encoding
}

// RequestOption modifies a RequestConfig
type RequestOption func(*RequestConfig)

// WithMethod sets the HTTP method (default: POST)
func WithMethod(method string) RequestOption {
	return func(c *RequestConfig) { c.Method = method }
}

// WithHeaders adds custom headers (merged with base headers)
func WithHeaders(headers map[string]string) RequestOption {
	return func(c *RequestConfig) {
		if c.Headers == nil {
			c.Headers = make(map[string]string)
		}
		for k, v := range headers {
			c.Headers[k] = v
		}
	}
}

// WithContext sets the request context
func WithContext(ctx context.Context) RequestOption {
	return func(c *RequestConfig) { c.Context = ctx }
}

// WithAuth enables bearer token authentication
func WithAuth() RequestOption {
	return func(c *RequestConfig) { c.Auth = true }
}

// WithCommonHeaders includes full API headers (vs minimal Auth + User-Agent)
func WithCommonHeaders() RequestOption {
	return func(c *RequestConfig) { c.CommonHeaders = true }
}

// WithStreamingResponse returns body as stream instead of reading it
func WithStreamingResponse() RequestOption {
	return func(c *RequestConfig) { c.StreamingResponse = true }
}

// WithStatusCheck enables response status validation via checkResponse
func WithStatusCheck() RequestOption {
	return func(c *RequestConfig) { c.CheckStatus = true }
}

// WithChunkedTransfer enables chunked transfer encoding (ContentLength = -1)
func WithChunkedTransfer() RequestOption {
	return func(c *RequestConfig) { c.ChunkedTransfer = true }
}

// ApiConfig holds the configuration needed to create an API client
type ApiConfig struct {
	AuthData   string     // Authentication string
	Proxy      string     // Proxy URL
	Quality    string     // Default quality: "original" or "storage-saver"
	UseQuota   bool       // If true, uploaded files count against storage quota (default: false)
	TokenCache TokenCache // Optional: custom token cache (nil = use MemoryTokenCache)
}

// Api represents a Google Photos API client
type Api struct {
	AndroidAPIVersion int64
	Model             string
	Make              string
	ClientVersionCode int64
	UserAgent         string
	Language          string
	AuthData          string
	Client            *http.Client
	tokenCache        TokenCache
	authMu            sync.Mutex // Protects token refresh
	Quality           string     // Default quality: "original" or "storage-saver"
	UseQuota          bool       // If true, uploaded files count against storage quota (default: false)
}

// NewApi creates a new Google Photos API client with the given configuration
func NewApi(cfg ApiConfig) (*Api, error) {
	if cfg.AuthData == "" {
		return nil, fmt.Errorf("auth data is required")
	}

	var language string
	params, err := url.ParseQuery(cfg.AuthData)
	if err == nil {
		language = params.Get("lang")
	}

	client, err := NewHTTPClientWithProxy(cfg.Proxy)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	tokenCache := cfg.TokenCache
	if tokenCache == nil {
		tokenCache = NewMemoryTokenCache()
	}

	api := &Api{
		AndroidAPIVersion: 28,
		Model:             "Pixel XL",
		Make:              "Google",
		ClientVersionCode: 49029607,
		Language:          language,
		AuthData:          strings.TrimSpace(cfg.AuthData),
		Client:            client,
		tokenCache:        tokenCache,
		Quality:           cfg.Quality,
		UseQuota:          cfg.UseQuota,
	}

	api.UserAgent = fmt.Sprintf(
		"com.google.android.apps.photos/%d (Linux; U; Android 9; %s; %s; Build/PQ2A.190205.001; Cronet/127.0.6510.5) (gzip)",
		api.ClientVersionCode,
		api.Language,
		api.Model,
	)

	return api, nil
}

// GetAuthToken returns a valid auth token, refreshing if necessary
func (a *Api) GetAuthToken() (string, error) {
	a.authMu.Lock()
	defer a.authMu.Unlock()

	token, expiry := a.tokenCache.Get()
	if token != "" && expiry > time.Now().Unix() {
		return token, nil
	}

	token, expiry, err := a.refreshAccessToken()
	if err != nil {
		return "", fmt.Errorf("failed to refresh auth token: %w", err)
	}

	a.tokenCache.Set(token, expiry)
	return token, nil
}

// refreshAccessToken fetches a new auth token from Google (expensive operation)
func (a *Api) refreshAccessToken() (authToken string, expiry int64, err error) {
	authDataValues, err := url.ParseQuery(a.AuthData)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse auth data: %w", err)
	}

	authRequestData := url.Values{
		"androidId":                    {authDataValues.Get("androidId")},
		"app":                          {"com.google.android.apps.photos"},
		"client_sig":                   {authDataValues.Get("client_sig")},
		"callerPkg":                    {"com.google.android.apps.photos"},
		"callerSig":                    {authDataValues.Get("callerSig")},
		"device_country":               {authDataValues.Get("device_country")},
		"Email":                        {authDataValues.Get("Email")},
		"google_play_services_version": {authDataValues.Get("google_play_services_version")},
		"lang":                         {authDataValues.Get("lang")},
		"oauth2_foreground":            {authDataValues.Get("oauth2_foreground")},
		"sdk_version":                  {authDataValues.Get("sdk_version")},
		"service":                      {authDataValues.Get("service")},
		"Token":                        {authDataValues.Get("Token")},
	}

	headers := map[string]string{
		"Accept-Encoding": "gzip",
		"app":             "com.google.android.apps.photos",
		"Connection":      "Keep-Alive",
		"Content-Type":    "application/x-www-form-urlencoded",
		"device":          authRequestData.Get("androidId"),
		"User-Agent":      "GoogleAuth/1.4 (Pixel XL PQ2A.190205.001); gzip",
	}

	req, err := http.NewRequest(
		"POST",
		"https://android.googleapis.com/auth",
		strings.NewReader(authRequestData.Encode()),
	)

	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	fmt.Println("Request URL:", req.URL.String())
	resp, err := a.Client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return "", 0, err
	}

	bodyBytes, err := readGzipBody(resp)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the key=value response format
	parsedAuthResponse := make(map[string]string)
	for _, line := range strings.Split(string(bodyBytes), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			parsedAuthResponse[parts[0]] = parts[1]
		}
	}

	// Validate we got the required fields
	if parsedAuthResponse["Auth"] == "" || parsedAuthResponse["Expiry"] == "" {
		return "", 0, errors.New("auth response missing Auth or Expiry token")
	}

	expiryInt, err := strconv.ParseInt(parsedAuthResponse["Expiry"], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse expiry: %w", err)
	}

	return parsedAuthResponse["Auth"], expiryInt, nil
}

// CommonHeaders returns the standard headers for Google Photos API requests
func (a *Api) CommonHeaders() map[string]string {
	return map[string]string{
		"Accept-Encoding":          "gzip",
		"Accept-Language":          a.Language,
		"Content-Type":             "application/x-protobuf",
		"User-Agent":               a.UserAgent,
		"x-goog-ext-173412678-bin": "CgcIAhClARgC",
		"x-goog-ext-174067345-bin": "CgIIAg==",
	}
}

// DeviceInfo returns the current device model and make info
func (a *Api) DeviceInfo() (model, make string, apiVersion int64) {
	return a.Model, a.Make, a.AndroidAPIVersion
}

// SetModel updates the device model (used for quality settings)
func (a *Api) SetModel(model string) {
	a.Model = model
}

// checkResponse checks if the HTTP response status is successful (2xx).
// Returns an error with the response body if status is not 2xx.
func checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	// Try to read and decompress the error response
	body, err := readGzipBody(resp)
	if err != nil {
		return fmt.Errorf("request failed with status %d (could not read response: %v)", resp.StatusCode, err)
	}
	// Use errors.New to avoid interpreting % characters in the response body
	bodyStr := string(body)
	return fmt.Errorf("request failed with status %d: %w", resp.StatusCode, errors.New(bodyStr))
}

// readGzipBody reads the response body, handling gzip decompression if needed.
func readGzipBody(resp *http.Response) ([]byte, error) {
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gr, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gr.Close()
		reader = gr
	}
	return io.ReadAll(reader)
}

// DoRequest executes an HTTP request with full lifecycle management.
// Returns body bytes, http.Response (for headers), and error.
// For streaming (WithStreamResponse), body is nil and caller must close resp.Body.
func (a *Api) DoRequest(url string, body io.Reader, opts ...RequestOption) ([]byte, *http.Response, error) {
	cfg := &RequestConfig{
		Method:        "POST",
		Headers:       make(map[string]string),
		Auth:          false,
		CommonHeaders: false,
		Context:       context.Background(),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	// Build headers based on config
	allHeaders := make(map[string]string)
	if cfg.CommonHeaders {
		for k, v := range a.CommonHeaders() {
			allHeaders[k] = v
		}
	}
	if cfg.Auth {
		authToken, err := a.GetAuthToken()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get bearer token: %w", err)
		}
		allHeaders["Authorization"] = "Bearer " + authToken
		allHeaders["User-Agent"] = a.UserAgent
	}

	// Merge custom headers (custom headers override defaults)
	for k, v := range cfg.Headers {
		allHeaders[k] = v
	}

	// Create request
	req, err := http.NewRequestWithContext(cfg.Context, cfg.Method, url, body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Enable chunked transfer if requested
	if cfg.ChunkedTransfer {
		req.ContentLength = -1
	}

	// Apply headers
	for k, v := range allHeaders {
		req.Header.Set(k, v)
	}

	// Execute request
	resp, err := a.Client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}

	// Validate response status if requested
	if cfg.CheckStatus {
		if err := checkResponse(resp); err != nil {
			resp.Body.Close()
			return nil, nil, err
		}
	}

	// For streaming responses, return without reading body
	if cfg.StreamingResponse {
		return nil, resp, nil
	}

	// Read body (handling gzip)
	bodyBytes, err := readGzipBody(resp)
	resp.Body.Close()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return bodyBytes, resp, nil
}

// DoProtoRequest marshals a protobuf request, sends it, and optionally unmarshals the response.
// If respMsg is nil, the response body is not unmarshaled (fire-and-forget).
func (a *Api) DoProtoRequest(url string, reqMsg proto.Message, respMsg proto.Message, opts ...RequestOption) error {
	serializedData, err := proto.Marshal(reqMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf: %w", err)
	}

	bodyBytes, _, err := a.DoRequest(url, bytes.NewReader(serializedData), opts...)
	if err != nil {
		return err
	}

	if respMsg != nil {
		if err := proto.Unmarshal(bodyBytes, respMsg); err != nil {
			return fmt.Errorf("failed to unmarshal protobuf: %w", err)
		}
	}

	return nil
}
