package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/twigman/fshare/src/config"
	"github.com/twigman/fshare/src/store"
)

type RESTService struct {
	config          *config.Config
	apiKeyService   *store.APIKeyService
	resourceService *store.ResourceService
	env             *config.Env
}

func NewRESTService(config *config.Config, as *store.APIKeyService, rs *store.ResourceService) *RESTService {
	return &RESTService{
		config:          config,
		apiKeyService:   as,
		resourceService: rs,
	}
}

func (s *RESTService) authorizeBearer(w http.ResponseWriter, r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeJSONStatus(w, http.StatusUnauthorized, "Missing Authorization header")
		return "", fmt.Errorf("missing authorization")
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		writeJSONStatus(w, http.StatusUnauthorized, "Invalid Authorization scheme")
		return "", fmt.Errorf("invalid scheme")
	}

	apiKey := strings.TrimPrefix(authHeader, prefix)
	keyUUID, err := s.apiKeyService.GetUUIDForAPIKey(apiKey)
	if err != nil || keyUUID == "" {
		writeJSONStatus(w, http.StatusUnauthorized, "Authorization failed")
		return "", fmt.Errorf("invalid api key")
	}
	return keyUUID, nil
}

func (s *RESTService) generateSignedURL(endpoint string, uuid string, exp time.Time) (string, error) {
	if s.env == nil {
		env, err := config.LoadOrCreateEnv(s.config.EnvPath)
		if err != nil {
			return "", err
		}
		s.env = env
	}

	// expiry time as unix timestamp
	expires := fmt.Sprintf("%d", exp.Unix())
	data := uuid + "|" + expires

	// hmac needs a hash function and a secret to create a signature
	mac := hmac.New(sha256.New, []byte(s.env.HMACSecret))
	// apply hmac
	mac.Write([]byte(data))
	signature := hex.EncodeToString(mac.Sum(nil))

	fullPath := path.Join("/", endpoint, uuid)

	return fmt.Sprintf("%s?expires=%s&signature=%s", fullPath, expires, signature), nil
}

func (s *RESTService) isValidSignedRequest(r *http.Request, uuid string) bool {
	if s.env == nil {
		return false
	}

	expiresStr := r.URL.Query().Get("expires")
	signature := r.URL.Query().Get("signature")
	if expiresStr == "" || signature == "" {
		return false
	}

	// base 10 -> int64
	expiresInt, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil || time.Now().Unix() > expiresInt {
		return false
	}

	data := uuid + "|" + expiresStr
	mac := hmac.New(sha256.New, []byte(s.env.HMACSecret))
	mac.Write([]byte(data))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// constant-time compare; checks all bytes for constant result
	if subtle.ConstantTimeCompare([]byte(signature), []byte(expectedSig)) != 1 {
		return false
	}

	return true
}

func writeJSONStatus(w http.ResponseWriter, statusCode int, statusMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `{"status": %q}`, statusMsg)
}

func writeJSONResponse(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
