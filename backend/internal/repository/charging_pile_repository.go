package repository

import (
	"database/sql"
	"errors"
	"time"

	"backend/internal/model"
)

// ChargingPileRepository 充电桩仓库
type ChargingPileRepository struct {
	db *sql.DB
}

// NewChargingPileRepository 创建充电桩仓库
func NewChargingPileRepository(db *sql.DB) *ChargingPileRepository {
	return &ChargingPileRepository{
		db: db,
	}
}

// GetAll 获取所有充电桩
func (r *ChargingPileRepository) GetAll() ([]*model.ChargingPile, error) {
	query := `
		SELECT id, pile_type, power, status, queue_length, 
		       total_sessions, total_duration, total_energy, created_at, updated_at
		FROM charging_piles
		ORDER BY id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var piles []*model.ChargingPile
	for rows.Next() {
		var pile model.ChargingPile
		err := rows.Scan(
			&pile.ID,
			&pile.PileType,
			&pile.Power,
			&pile.Status,
			&pile.QueueLength,
			&pile.TotalSessions,
			&pile.TotalDuration,
			&pile.TotalEnergy,
			&pile.CreatedAt,
			&pile.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		piles = append(piles, &pile)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return piles, nil
}

// GetByID 根据ID获取充电桩
func (r *ChargingPileRepository) GetByID(id string) (*model.ChargingPile, error) {
	query := `
		SELECT id, pile_type, power, status, queue_length, 
		       total_sessions, total_duration, total_energy, created_at, updated_at
		FROM charging_piles
		WHERE id = $1
	`

	var pile model.ChargingPile
	err := r.db.QueryRow(query, id).Scan(
		&pile.ID,
		&pile.PileType,
		&pile.Power,
		&pile.Status,
		&pile.QueueLength,
		&pile.TotalSessions,
		&pile.TotalDuration,
		&pile.TotalEnergy,
		&pile.CreatedAt,
		&pile.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("充电桩不存在")
		}
		return nil, err
	}

	return &pile, nil
}

// GetByType 根据类型获取充电桩
func (r *ChargingPileRepository) GetByType(pileType model.PileType) ([]*model.ChargingPile, error) {
	query := `
		SELECT id, pile_type, power, status, queue_length, 
		       total_sessions, total_duration, total_energy, created_at, updated_at
		FROM charging_piles
		WHERE pile_type = $1
		ORDER BY id
	`

	rows, err := r.db.Query(query, pileType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var piles []*model.ChargingPile
	for rows.Next() {
		var pile model.ChargingPile
		err := rows.Scan(
			&pile.ID,
			&pile.PileType,
			&pile.Power,
			&pile.Status,
			&pile.QueueLength,
			&pile.TotalSessions,
			&pile.TotalDuration,
			&pile.TotalEnergy,
			&pile.CreatedAt,
			&pile.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		piles = append(piles, &pile)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return piles, nil
}

// UpdateStatus 更新充电桩状态
func (r *ChargingPileRepository) UpdateStatus(id string, status model.PileStatus) error {
	query := `
		UPDATE charging_piles
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(query, status, time.Now().UTC(), id)
	return err
}

// UpdateStats 更新充电桩统计信息
func (r *ChargingPileRepository) UpdateStats(id string, sessionCount int, duration, energy float64) error {
	query := `
		UPDATE charging_piles
		SET total_sessions = total_sessions + $1, 
		    total_duration = total_duration + $2,
		    total_energy = total_energy + $3,
		    updated_at = $4
		WHERE id = $5
	`

	_, err := r.db.Exec(query, sessionCount, duration, energy, time.Now().UTC(), id)
	return err
}

// UpdateQueueLength 更新充电桩队列长度
func (r *ChargingPileRepository) UpdateQueueLength(id string, queueLength int) error {
	query := `
		UPDATE charging_piles
		SET queue_length = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(query, queueLength, time.Now().UTC(), id)
	return err
}

// GetAvailablePiles 获取可用的充电桩
func (r *ChargingPileRepository) GetAvailablePiles(pileType model.PileType, maxQueueLength int) ([]*model.ChargingPile, error) {
	query := `
		SELECT id, pile_type, power, status, queue_length, 
		       total_sessions, total_duration, total_energy, created_at, updated_at
		FROM charging_piles
		WHERE pile_type = $1 AND status != 'fault' AND status != 'maintenance' AND status != 'offline' AND queue_length < $2
		ORDER BY queue_length, id
	`

	rows, err := r.db.Query(query, pileType, maxQueueLength)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var piles []*model.ChargingPile
	for rows.Next() {
		var pile model.ChargingPile
		err := rows.Scan(
			&pile.ID,
			&pile.PileType,
			&pile.Power,
			&pile.Status,
			&pile.QueueLength,
			&pile.TotalSessions,
			&pile.TotalDuration,
			&pile.TotalEnergy,
			&pile.CreatedAt,
			&pile.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		piles = append(piles, &pile)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return piles, nil
}

// GetNormalPiles 获取正常状态的充电桩
func (r *ChargingPileRepository) GetNormalPiles(pileType model.PileType) ([]*model.ChargingPile, error) {
	query := `
		SELECT id, pile_type, power, status, queue_length, 
		       total_sessions, total_duration, total_energy, created_at, updated_at
		FROM charging_piles
		WHERE pile_type = $1 AND status != 'fault' AND status != 'maintenance' AND status != 'offline'
		ORDER BY queue_length, id
	`

	rows, err := r.db.Query(query, pileType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var piles []*model.ChargingPile
	for rows.Next() {
		var pile model.ChargingPile
		err := rows.Scan(
			&pile.ID,
			&pile.PileType,
			&pile.Power,
			&pile.Status,
			&pile.QueueLength,
			&pile.TotalSessions,
			&pile.TotalDuration,
			&pile.TotalEnergy,
			&pile.CreatedAt,
			&pile.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		piles = append(piles, &pile)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return piles, nil
}

// Create 创建新充电桩
func (r *ChargingPileRepository) Create(pile *model.ChargingPile) error {
	query := `
		INSERT INTO charging_piles (id, pile_type, power, status, queue_length, total_sessions, total_duration, total_energy, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO NOTHING
	`

	_, err := r.db.Exec(
		query,
		pile.ID,
		pile.PileType,
		pile.Power,
		pile.Status,
		pile.QueueLength,
		pile.TotalSessions,
		pile.TotalDuration,
		pile.TotalEnergy,
		pile.CreatedAt,
		pile.UpdatedAt,
	)
	return err
}
