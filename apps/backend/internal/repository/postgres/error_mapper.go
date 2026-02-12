package postgres

import (
	"database/sql"
	"errors"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/lib/pq"
)

func mapDBError(err error) error {
	if err == nil {
		return nil
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505": // unique_violation
			return domain.ErrConflict
		case "23503": // foreign_key_violation
			return domain.ErrInvalidInput
		case "23502": // not_null_violation
			return domain.ErrInvalidInput
		}
	}

	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}

	return err
}
