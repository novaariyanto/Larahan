package services

import (
	"fmt"
	"strings"

	"github.com/larahan/larahan/backend/apache"
	"github.com/larahan/larahan/backend/models"
	"github.com/larahan/larahan/backend/mysql"
	"github.com/larahan/larahan/backend/php"
	"github.com/larahan/larahan/backend/phpmyadmin"
)

// DashboardService exposes environment summary to the frontend.
type DashboardService struct {
	apache     *apache.Manager
	php        *php.Manager
	mysql      *mysql.Manager
	phpmyadmin *phpmyadmin.Manager
}

// NewDashboardService wires component managers into the dashboard facade.
func NewDashboardService(apacheMgr *apache.Manager, phpMgr *php.Manager, mysqlMgr *mysql.Manager, pmaMgr *phpmyadmin.Manager) *DashboardService {
	return &DashboardService{
		apache:     apacheMgr,
		php:        phpMgr,
		mysql:      mysqlMgr,
		phpmyadmin: pmaMgr,
	}
}

// GetSummary returns current Apache, MySQL, PHP, and phpMyAdmin overview.
func (s *DashboardService) GetSummary() models.DashboardSummary {
	phpInfo := s.php.Info()
	return models.DashboardSummary{
		Apache:     s.apache.Info(),
		MySQL:      s.mysql.Info(),
		ActivePHP:  s.php.Active(),
		PHP:        phpInfo,
		PhpMyAdmin: s.phpmyadmin.Info(),
	}
}

// StartAll starts Apache and MySQL.
func (s *DashboardService) StartAll() models.Result {
	var errs []string
	if err := s.apache.Start(); err != nil {
		errs = append(errs, "Apache: "+err.Error())
	}
	if err := s.mysql.Start(); err != nil {
		errs = append(errs, "MySQL: "+err.Error())
	}
	if len(errs) > 0 {
		return models.FailResult(strings.Join(errs, "; "))
	}
	return models.OKResult("Apache & MySQL started")
}

// StopAll stops Apache and MySQL.
func (s *DashboardService) StopAll() models.Result {
	var errs []string
	if err := s.apache.Stop(); err != nil {
		errs = append(errs, "Apache: "+err.Error())
	}
	if err := s.mysql.Stop(); err != nil {
		errs = append(errs, "MySQL: "+err.Error())
	}
	if len(errs) > 0 {
		return models.FailResult(strings.Join(errs, "; "))
	}
	return models.OKResult("Apache & MySQL stopped")
}

// RestartAll restarts Apache and MySQL.
func (s *DashboardService) RestartAll() models.Result {
	var errs []string
	if err := s.apache.Restart(); err != nil {
		errs = append(errs, "Apache: "+err.Error())
	}
	if err := s.mysql.Restart(); err != nil {
		errs = append(errs, "MySQL: "+err.Error())
	}
	if len(errs) > 0 {
		return models.FailResult(fmt.Sprintf("%s", strings.Join(errs, "; ")))
	}
	return models.OKResult("Apache & MySQL restarted")
}
