package repository

import (
	"database/sql"
	"fmt"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/models"
)

// AlertRepository encapsule l'accès à la table "alert_settings" (ligne unique).
type AlertRepository struct {
	db *sql.DB
}

func NewAlertRepository(db *sql.DB) *AlertRepository {
	return &AlertRepository{db: db}
}

// GetSettings retourne le seuil d'alerte par défaut courant.
func (r *AlertRepository) GetSettings() (*models.AlertSettings, error) {
	row := r.db.QueryRow(`SELECT id, default_threshold, updated_at FROM alert_settings WHERE id = 1`)
	var s models.AlertSettings
	if err := row.Scan(&s.ID, &s.DefaultThreshold, &s.UpdatedAt); err != nil {
		return nil, fmt.Errorf("lecture réglages d'alerte : %w", err)
	}
	return &s, nil
}

// UpdateSettings modifie le seuil d'alerte par défaut.
func (r *AlertRepository) UpdateSettings(defaultThreshold int) (*models.AlertSettings, error) {
	_, err := r.db.Exec(
		`UPDATE alert_settings SET default_threshold = ?, updated_at = CURRENT_TIMESTAMP WHERE id = 1`,
		defaultThreshold,
	)
	if err != nil {
		return nil, fmt.Errorf("mise à jour réglages d'alerte : %w", err)
	}
	return r.GetSettings()
}
