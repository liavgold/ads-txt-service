package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"ads-txt-service/internal/cache"
	"ads-txt-service/internal/config"
	"ads-txt-service/internal/fetcher"
	"ads-txt-service/internal/logger"
	"ads-txt-service/internal/middleware" // Import the middleware package
	"ads-txt-service/internal/models"
	"ads-txt-service/internal/parser"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)


type AdsCache interface {
	GetAds(ctx context.Context, key string) (*models.AdsResponse, bool)
	SetAds(ctx context.Context, key string, resp *models.AdsResponse, ttl time.Duration) error
}

type AdsFetcher interface {
	FetchAdsTxt(ctx context.Context, domain string) (string, error)
}

type AdsParser interface {
	ParseAdsTxt(r io.Reader) map[string]int
}

type Server struct {
	cfg    *config.Config
	cache   AdsCache
	log    *logger.Logger
	ft      AdsFetcher
	parser AdsParser
	rl     *middleware.RateLimiter
}

func NewServer(
	cfg *config.Config,
	adsCache *cache.AdsCache,
	log *logger.Logger,
	ft *fetcher.Fetcher,
	parser *parser.Parser,
) *Server {
	rl := middleware.NewRateLimiter(cfg.LimiterMaxReq, time.Duration(cfg.LimmiterTTL)*time.Second, log)

	return &Server{
		cfg:    cfg,
		cache:  adsCache,
		log:    log,
		ft:     ft,
		parser: parser,
		rl:     rl,
	}
}

func (s *Server) Router() http.Handler {
	r := mux.NewRouter()

	r.Handle("/ads", s.rl.RateLimitMiddleware()(http.HandlerFunc(s.GetAds))).Methods(http.MethodGet)
    r.HandleFunc("/health", s.Health).Methods(http.MethodGet)

	return r
}


func (s *Server) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) GetAds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domain := strings.TrimSpace(r.URL.Query().Get("domain"))
	if domain == "" {
		http.Error(w, "missing domain", http.StatusBadRequest)
		return
	}

	if !isValidDomain(domain) {
		http.Error(w, "invalid domain", http.StatusBadRequest)
		return
	}

	if cached, found := s.cache.GetAds(ctx, domain); found {
		s.log.Infow("Cache hit", "domain", domain)
		writeJSON(w, cached)
		return
	}

	s.log.Infow("Fetching ads.txt", "domain", domain)
	content, err := s.ft.FetchAdsTxt(ctx, domain)
	if err != nil {
		s.log.Errorw("Failed to fetch ads.txt", zap.Error(err), "domain", domain)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	advertisersMap := s.parser.ParseAdsTxt(strings.NewReader(content))
	advertisers := make([]*models.Advertiser, 0, len(advertisersMap))
	for ad, count := range advertisersMap {
		advertisers = append(advertisers, &models.Advertiser{
			Domain: ad,
			Count:  count,
		})
	}

	resp := &models.AdsResponse{
		Domain:           domain,
		TotalAdvertisers: len(advertisersMap),
		Advertisers:      advertisers,
		Cached:           false,
		Timestamp:        time.Now().UTC(),
	}

	s.cache.SetAds(ctx, domain, resp, s.cfg.CacheTTL)
	writeJSON(w, resp)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		logger.L().Errorw("Failed to encode JSON", zap.Error(err))
	}
}

func isValidDomain(domain string) bool {
	if len(domain) > 255 || len(domain) < 3 {
		return false
	}
	return regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-]*\.[a-zA-Z]{2,}$`).MatchString(domain)
}