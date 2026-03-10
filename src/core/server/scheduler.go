package server

import (
	"certsync/src/core/cert"
	"certsync/src/core/meta"
	"log/slog"
	"sync"
	"time"
)

type Scheduler struct {
	Config      *meta.ServerConfig
	Storage     *cert.CertStorage
	CertManager *CertManager
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

func NewScheduler(config *meta.ServerConfig, storage *cert.CertStorage) *Scheduler {
	return &Scheduler{
		Config:      config,
		Storage:     storage,
		CertManager: NewCertManager(storage, config.DNS),
		stopChan:    make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	scheduleTime := s.parseCheckTime()
	initialDelay := s.calculateInitialDelay(scheduleTime)

	slog.Info("Scheduler: first check scheduled", "delay", initialDelay, "schedule", scheduleTime)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		timer := time.NewTimer(initialDelay)
		defer timer.Stop()

		select {
		case <-timer.C:
			s.CheckAllCerts()
		case <-s.stopChan:
			return
		}

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.CheckAllCerts()
			case <-s.stopChan:
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	close(s.stopChan)
	s.wg.Wait()
}

func (s *Scheduler) CheckAllCerts() {
	slog.Info("Scheduler: starting certificate check")

	for _, certConfig := range s.Config.Certs {
		s.checkAndRenewCert(certConfig)
	}

	slog.Info("Scheduler: certificate check completed")
}

// CheckAndRenewCertStatus 检查并尝试更新证书，返回操作状态
// 状态值: CERT_NONE(-2), CERT_RENEW_FAILED(-1), CERT_VALID(0), CERT_RENEW_SUCCESS(1)
func (s *Scheduler) CheckAndRenewCertStatus(certConfig *meta.CertConfig) int {
	alias := certConfig.Alias

	// Skip auto-renewal if disabled
	if !certConfig.AutoRenew {
		slog.Info("Skipping certificate (auto_renew disabled)", "alias", alias)
		return CERT_NONE
	}

	slog.Info("Checking certificate", "alias", alias)

	var localExpiry time.Time
	var hasLocalCert bool

	localExpiry, err := cert.GetLocalCertExpiry(s.Storage.GetFullchainPath(alias))
	if err == nil {
		hasLocalCert = true
		slog.Info("Certificate local expiry", "alias", alias, "expiry", localExpiry)
	} else {
		hasLocalCert = false
		slog.Info("No local cert found, checking remote", "alias", alias)
	}

	var remoteExpiry time.Time
	var hasRemoteCert bool

	remoteExpiry, err = cert.GetRemoteCertExpiry(certConfig.DomainCheck)
	if err == nil {
		hasRemoteCert = true
		slog.Info("Certificate remote expiry", "alias", alias, "expiry", remoteExpiry)
	} else {
		hasRemoteCert = false
		slog.Warn("Failed to check remote cert", "alias", alias, "error", err)
	}

	shouldRenew := false
	reason := ""

	if !hasLocalCert {
		if hasRemoteCert {
			if cert.NeedsRenewal(remoteExpiry, 10) {
				shouldRenew = true
				reason = "no local cert and remote cert expiring soon"
			}
		} else {
			shouldRenew = true
			reason = "no local cert available"
		}
	} else {
		if cert.NeedsRenewal(localExpiry, 10) {
			shouldRenew = true
			reason = "local cert expiring within 10 days"
		}
	}

	if shouldRenew {
		slog.Info("Renewing certificate", "alias", alias, "reason", reason)
		if err := s.CertManager.RenewCert(certConfig); err != nil {
			slog.Error("Certificate renewal failed", "alias", alias, "error", err)
			return CERT_RENEW_FAILED
		} else {
			slog.Info("Certificate renewed successfully", "alias", alias)
			return CERT_RENEW_SUCCESS
		}
	} else {
		slog.Info("Certificate no renewal needed", "alias", alias)
		return CERT_VALID
	}
}

func (s *Scheduler) checkAndRenewCert(certConfig *meta.CertConfig) {
	s.CheckAndRenewCertStatus(certConfig)
}

func (s *Scheduler) parseCheckTime() string {
	if s.Config.CheckTime == "" {
		return "03:00:00"
	}
	return s.Config.CheckTime
}

func (s *Scheduler) calculateInitialDelay(scheduleTime string) time.Duration {
	now := time.Now()
	next, err := time.ParseInLocation("15:04:05", scheduleTime, now.Location())
	if err != nil {
		slog.Error("Failed to parse schedule time, using default 03:00:00", "error", err)
		next, _ = time.ParseInLocation("15:04:05", "03:00:00", now.Location())
	}

	// Set the date to today
	next = time.Date(now.Year(), now.Month(), now.Day(), next.Hour(), next.Minute(), next.Second(), 0, now.Location())

	if next.Before(now) {
		next = next.Add(24 * time.Hour)
	}

	return time.Until(next)
}
