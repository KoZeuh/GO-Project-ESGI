package repository

import "errors"

// Erreurs sentinelles renvoyées par les repositories et interprétées par la
// couche service pour choisir le bon code HTTP dans les handlers.
var (
	ErrNotFound     = errors.New("ressource introuvable")
	ErrDuplicate    = errors.New("ressource déjà existante")
	ErrInsufficient = errors.New("quantité insuffisante en stock")
)
