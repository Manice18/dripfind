package history

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/manice18/outfit_finder/backend/internal/models"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) CreateOutfit(ctx context.Context, sourceURL string) (*models.Outfit, error) {
	var o models.Outfit
	err := s.pool.QueryRow(ctx, `
		INSERT INTO outfits (source_url, status)
		VALUES ($1, $2)
		RETURNING id, image_path, source_url, style, gender, season, occasion, status, error_message, created_at, updated_at
	`, sourceURL, models.StatusPending).Scan(
		&o.ID, &o.ImagePath, &o.SourceURL, &o.Style, &o.Gender, &o.Season, &o.Occasion,
		&o.Status, &o.ErrorMessage, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (s *Store) MarkProcessing(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE outfits SET status=$2, updated_at=NOW() WHERE id=$1
	`, id, models.StatusProcessing)
	return err
}

func (s *Store) MarkFailed(ctx context.Context, id uuid.UUID, msg string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE outfits SET status=$2, error_message=$3, updated_at=NOW() WHERE id=$1
	`, id, models.StatusFailed, msg)
	return err
}

func (s *Store) SaveAnalysis(ctx context.Context, outfit *models.Outfit) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE outfits
		SET image_path=$2, style=$3, gender=$4, season=$5, occasion=$6,
		    status=$7, error_message='', updated_at=NOW()
		WHERE id=$1
	`, outfit.ID, outfit.ImagePath, outfit.Style, outfit.Gender, outfit.Season, outfit.Occasion, models.StatusCompleted)
	if err != nil {
		return err
	}

	for i := range outfit.Items {
		item := &outfit.Items[i]
		err = tx.QueryRow(ctx, `
			INSERT INTO clothing_items (outfit_id, category, color, material, fit, pattern, confidence, search_query)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			RETURNING id
		`, outfit.ID, item.Category, item.Color, item.Material, item.Fit, item.Pattern, item.Confidence, item.SearchQuery).Scan(&item.ID)
		if err != nil {
			return fmt.Errorf("insert item: %w", err)
		}

		for j := range item.Products {
			p := &item.Products[j]
			err = tx.QueryRow(ctx, `
				INSERT INTO products (item_id, title, brand, price, currency, website, url, image, match_score)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
				RETURNING id
			`, item.ID, p.Title, p.Brand, p.Price, p.Currency, p.Website, p.URL, p.Image, p.MatchScore).Scan(&p.ID)
			if err != nil {
				return fmt.Errorf("insert product: %w", err)
			}
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO history (outfit_id, source_url) VALUES ($1, $2)
	`, outfit.ID, outfit.SourceURL)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Store) GetOutfit(ctx context.Context, id uuid.UUID) (*models.Outfit, error) {
	var o models.Outfit
	err := s.pool.QueryRow(ctx, `
		SELECT id, image_path, source_url, style, gender, season, occasion, status, error_message, created_at, updated_at
		FROM outfits WHERE id=$1
	`, id).Scan(
		&o.ID, &o.ImagePath, &o.SourceURL, &o.Style, &o.Gender, &o.Season, &o.Occasion,
		&o.Status, &o.ErrorMessage, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	itemRows, err := s.pool.Query(ctx, `
		SELECT id, outfit_id, category, color, material, fit, pattern, confidence, search_query
		FROM clothing_items WHERE outfit_id=$1 ORDER BY confidence DESC
	`, id)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var item models.ClothingItem
		if err := itemRows.Scan(
			&item.ID, &item.OutfitID, &item.Category, &item.Color, &item.Material,
			&item.Fit, &item.Pattern, &item.Confidence, &item.SearchQuery,
		); err != nil {
			return nil, err
		}

		prodRows, err := s.pool.Query(ctx, `
			SELECT id, item_id, title, brand, price, currency, website, url, image, match_score
			FROM products WHERE item_id=$1 ORDER BY match_score DESC
		`, item.ID)
		if err != nil {
			return nil, err
		}
		for prodRows.Next() {
			var p models.Product
			if err := prodRows.Scan(
				&p.ID, &p.ItemID, &p.Title, &p.Brand, &p.Price, &p.Currency,
				&p.Website, &p.URL, &p.Image, &p.MatchScore,
			); err != nil {
				prodRows.Close()
				return nil, err
			}
			item.Products = append(item.Products, p)
		}
		prodRows.Close()
		o.Items = append(o.Items, item)
	}

	return &o, itemRows.Err()
}

func (s *Store) ListHistory(ctx context.Context, limit int) ([]models.HistoryEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	rows, err := s.pool.Query(ctx, `
		SELECT h.id, h.outfit_id, h.source_url, h.created_at,
		       o.style, o.gender, o.status, o.image_path
		FROM history h
		JOIN outfits o ON o.id = h.outfit_id
		ORDER BY h.created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.HistoryEntry
	for rows.Next() {
		var e models.HistoryEntry
		var imagePath string
		if err := rows.Scan(&e.ID, &e.OutfitID, &e.SourceURL, &e.CreatedAt, &e.Style, &e.Gender, &e.Status, &imagePath); err != nil {
			return nil, err
		}
		if imagePath != "" {
			e.ImageURL = "/images/" + imagePath
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) DeleteHistory(ctx context.Context, id uuid.UUID) error {
	ct, err := s.pool.Exec(ctx, `DELETE FROM history WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}
