package verification

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/goadesign/goa"

	"github.com/fabric8-services/fabric8-webhook/util"
)

const (
	meta = "https://api.github.com/meta"
)

type service struct {
	// IP address ranges specifying incoming webhooks
	// and service hookIPs that originates from on GitHub.com
	hookIPs []*net.IPNet
	// lock for writing to hooks
	lock    sync.RWMutex
	Service *goa.Service
	ticker  *time.Ticker
}

// Service defines verification
type Service interface {
	Verify(req *http.Request) (bool, error)
}

// New returns a verification service instance
func New(gs *goa.Service, duration time.Duration) (Service, error) {
	s := &service{
		Service: gs,
		ticker:  time.NewTicker(duration),
	}
	if err := s.setHookIPs(); err != nil {
		return nil, err
	}
	return s, nil
}

// Verify verifies whether request came
// from approved source
func (s *service) Verify(req *http.Request) (bool, error) {
	ip := strings.Split(req.Header.Get("X-Forwarded-For"), ",")
	if len(ip[0]) == 0 {
		ip = strings.Split(req.RemoteAddr, ":")
	}
	s.Service.LogInfo("IP", "ip", ip)

	if !s.isGithubIP(ip[0]) {
		if err := s.setHookIPs(); err != nil {
			s.Service.LogError("Error while setting up"+
				" hookips", "err", err)
			return false, nil
		}
		return s.isGithubIP(ip[0]), nil
	}
	return true, nil
}

func (s *service) setHookIPs() error {
	res, err := util.NetClient.Get(meta)
	if err != nil {
		s.Service.LogError("Error while making request to github:",
			"err", err)
		return err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		s.Service.LogError("Error while reading response body"+
			" from github:", "err", err)
		return err
	}

	ips := struct {
		Hooks []string `json:"hooks"`
	}{}
	if err := json.Unmarshal(body, &ips); err != nil {
		s.Service.LogError("Error while making request to github:",
			"err", err)
		return err
	}
	var ipnets []*net.IPNet
	for _, hook := range ips.Hooks {
		_, ipnet, err := net.ParseCIDR(hook)
		if err != nil {
			s.Service.LogError("Error parsing ipnet from github",
				"err", err)
			return err
		}
		ipnets = append(ipnets, ipnet)
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.hookIPs = ipnets
	return nil
}

func (s *service) isGithubIP(i string) bool {
	ip := net.ParseIP(i)
	s.lock.RLock()
	defer s.lock.RUnlock()
	for _, ipnet := range s.hookIPs {
		if ipnet.Contains(ip) {
			return true
		}
	}
	return false
}
