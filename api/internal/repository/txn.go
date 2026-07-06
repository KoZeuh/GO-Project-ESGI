package repository

import "database/sql"

// dbtx est le sous-ensemble de *sql.DB utilisé par les repositories. Il est
// aussi implémenté par *sql.Tx, ce qui permet à un service d'exécuter
// plusieurs opérations de repositories différents dans une même transaction
// via la méthode WithTx de chacun (voir MovementService.Create).
type dbtx interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}
