package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ads-txt-service/internal/config"
	"ads-txt-service/internal/logger"
	"ads-txt-service/internal/middleware"
	"ads-txt-service/internal/models"
)


type mockAdsCache struct {
	getFunc func(ctx context.Context, key string) (*models.AdsResponse, bool)
	setFunc func(ctx context.Context, key string, resp *models.AdsResponse, ttl time.Duration) error
}

func (m *mockAdsCache) GetAds(ctx context.Context, key string) (*models.AdsResponse, bool) {
	return m.getFunc(ctx, key)
}

func (m *mockAdsCache) SetAds(ctx context.Context, key string, resp *models.AdsResponse, ttl time.Duration) error {
	return m.setFunc(ctx, key, resp, ttl)
}

type mockAdsFetcher struct {
	fetchFunc func(ctx context.Context, domain string) (string, error)
}

func (m *mockAdsFetcher) FetchAdsTxt(ctx context.Context, domain string) (string, error) {
	return m.fetchFunc(ctx, domain)
}

type mockAdsParser struct {
	parseFunc func(r io.Reader) map[string]int
}

func (m *mockAdsParser) ParseAdsTxt(r io.Reader) map[string]int {
	return m.parseFunc(r)
}

func NewMockServer(
	cfg *config.Config,
	adsCache *mockAdsCache,
	log *logger.Logger,
	ft *mockAdsFetcher,
	parser *mockAdsParser,
) *Server {
	rl := middleware.NewRateLimiter(cfg.LimiterMaxReq, time.Duration(cfg.LimmiterTTL), log)

	return &Server{
		cfg:    cfg,
		cache:  adsCache,
		log:    log,
		ft:     ft,
		parser: parser,
		rl:     rl,
	}
}

func TestMain(m *testing.M) {
	logger.Init("Info")
	m.Run()
}

func TestServer_Health(t *testing.T) {
	server := NewMockServer(nil, &mockAdsCache{}, logger.L(), &mockAdsFetcher{}, &mockAdsParser{})
	router := server.Router()

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `{"status":"ok"}` + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestServer_GetAds(t *testing.T) {
	mockC := &mockAdsCache{}
	mockF := &mockAdsFetcher{}
	mockP := &mockAdsParser{}

	cfg := &config.Config{
		CacheTTL:      1 * time.Hour,
		LimiterMaxReq: 100,
		LimmiterTTL:   60,
	}

	s := NewMockServer(cfg, mockC, logger.L(), mockF, mockP)
	router := s.Router()

	testCases := []struct {
		name           string
		domain         string
		setupMocks     func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Successful request with cache miss",
			domain: "test.com",
			setupMocks: func() {
				mockC.getFunc = func(ctx context.Context, key string) (*models.AdsResponse, bool) {
					return nil, false 
				}
				mockF.fetchFunc = func(ctx context.Context, domain string) (string, error) {
					return "advertiser.com, pub-123, DIRECT\n", nil
				}
				mockP.parseFunc = func(r io.Reader) map[string]int {
					return map[string]int{"advertiser.com": 1}
				}
				mockC.setFunc = func(ctx context.Context, key string, resp *models.AdsResponse, ttl time.Duration) error {
					return nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"total_advertisers":1`,
		},
		{
			name:   "Successful request with cache hit",
			domain: "cached.com",
			setupMocks: func() {
				mockC.getFunc = func(ctx context.Context, key string) (*models.AdsResponse, bool) {
					return &models.AdsResponse{Domain: key, Cached: true}, true
				}
				
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"domain":"cached.com","cached":true`,
		},
		{
			name:           "Request with invalid domain",
			domain:         "invalid-domain",
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid domain\n",
		},
		{
			name:           "Missing domain parameter",
			domain:         "",
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "missing domain\n",
		},
		{
			name:   "Fetcher returns an error",
			domain: "fetcherror.com",
			setupMocks: func() {
				mockC.getFunc = func(ctx context.Context, key string) (*models.AdsResponse, bool) {
					return nil, false 
				}
				mockF.fetchFunc = func(ctx context.Context, domain string) (string, error) {
					return "", errors.New("failed to fetch") 
				}
				
			},
			expectedStatus: http.StatusBadGateway,
			expectedBody:   "failed to fetch\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			req, _ := http.NewRequest("GET", "/ads?domain="+tc.domain, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.expectedStatus)
			}
			if !strings.Contains(rr.Body.String(), tc.expectedBody) {
				t.Errorf("handler returned unexpected body; expected body to contain %q, but got %q", tc.expectedBody, rr.Body.String())
			}
		})
	}
}
