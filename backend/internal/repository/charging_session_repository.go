package repository

import (
	"database/sql"
	"errors"
	"time"

	"backend/internal/model"

	"github.com/google/uuid"
)

// ChargingSessionRepository 充电会话仓库
type ChargingSessionRepository struct {
	db *sql.DB
}

// NewChargingSessionRepository 创建充电会话仓库
func NewChargingSessionRepository(db *sql.DB) *ChargingSessionRepository {
	return &ChargingSessionRepository{
		db: db,
	}
}

// Create 创建充电会话
func (r *ChargingSessionRepository) Create(session *model.ChargingSession) (*model.ChargingSession, error) {
	query := `
		INSERT INTO charging_sessions 
		(id, request_id, user_id, pile_id, queue_number, requested_capacity, actual_capacity,
		 start_time, status, duration, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, request_id, user_id, pile_id, queue_number, requested_capacity, 
				  actual_capacity, start_time, duration, status, created_at
	`
	now := time.Now().UTC()
	var newSession model.ChargingSession
	err := r.db.QueryRow(
		query,
		session.ID,
		session.RequestID,
		session.UserID,
		session.PileID,
		session.QueueNumber,
		session.RequestedCapacity,
		session.ActualCapacity,
		session.StartTime,
		session.Status,
		session.Duration,
		now,
	).Scan(
		&newSession.ID,
		&newSession.RequestID,
		&newSession.UserID,
		&newSession.PileID,
		&newSession.QueueNumber,
		&newSession.RequestedCapacity,
		&newSession.ActualCapacity,
		&newSession.StartTime,
		&newSession.Duration,
		&newSession.Status,
		&newSession.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &newSession, nil
}

// GetByID 通过ID获取充电会话
func (r *ChargingSessionRepository) GetByID(id uuid.UUID) (*model.ChargingSession, error) {
	query := `
		SELECT id, request_id, user_id, pile_id, queue_number, requested_capacity, 
			   actual_capacity, start_time, end_time, status, duration, created_at
		FROM charging_sessions
		WHERE id = $1
	`

	var session model.ChargingSession
	var endTime sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&session.ID,
		&session.RequestID,
		&session.UserID,
		&session.PileID,
		&session.QueueNumber,
		&session.RequestedCapacity,
		&session.ActualCapacity,
		&session.StartTime,
		&endTime,
		&session.Status,
		&session.Duration,
		&session.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("充电会话不存在")
		}
		return nil, err
	}

	if endTime.Valid {
		session.EndTime = &endTime.Time
	}

	return &session, nil
}

// GetByRequestID 通过请求ID获取充电会话
func (r *ChargingSessionRepository) GetByRequestID(requestID uuid.UUID) (*model.ChargingSession, error) {
	query := `
		SELECT id, request_id, user_id, pile_id, queue_number, requested_capacity, 
			   actual_capacity, start_time, end_time, status, duration, created_at
		FROM charging_sessions
		WHERE request_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var session model.ChargingSession
	var endTime sql.NullTime

	err := r.db.QueryRow(query, requestID).Scan(
		&session.ID,
		&session.RequestID,
		&session.UserID,
		&session.PileID,
		&session.QueueNumber,
		&session.RequestedCapacity,
		&session.ActualCapacity,
		&session.StartTime,
		&endTime,
		&session.Status,
		&session.Duration,
		&session.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("充电会话不存在")
		}
		return nil, err
	}

	if endTime.Valid {
		session.EndTime = &endTime.Time
	}

	return &session, nil
}

// GetActiveSessionByPileID 通过充电桩ID获取活跃会话
func (r *ChargingSessionRepository) GetActiveSessionByPileID(pileID string) (*model.ChargingSession, error) {
	query := `
		SELECT id, request_id, user_id, pile_id, queue_number, requested_capacity, 
			   actual_capacity, start_time, end_time, status, duration, created_at
		FROM charging_sessions
		WHERE pile_id = $1 AND status = 'active'
		ORDER BY start_time DESC
		LIMIT 1
	`

	var session model.ChargingSession
	var endTime sql.NullTime

	err := r.db.QueryRow(query, pileID).Scan(
		&session.ID,
		&session.RequestID,
		&session.UserID,
		&session.PileID,
		&session.QueueNumber,
		&session.RequestedCapacity,
		&session.ActualCapacity,
		&session.StartTime,
		&endTime,
		&session.Status,
		&session.Duration,
		&session.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("没有活跃的充电会话")
		}
		return nil, err
	}

	if endTime.Valid {
		session.EndTime = &endTime.Time
	}

	return &session, nil
}

// Update 更新充电会话
func (r *ChargingSessionRepository) Update(session *model.ChargingSession) error {
	query := `
		UPDATE charging_sessions
		SET actual_capacity = $1, end_time = $2, status = $3, duration = $4
		WHERE id = $5
	`

	var endTime any = nil
	if session.EndTime != nil {
		endTime = *session.EndTime
	}

	_, err := r.db.Exec(
		query,
		session.ActualCapacity,
		endTime,
		session.Status,
		session.Duration,
		session.ID,
	)

	return err
}

// GetUserSessions 获取用户的充电会话历史
func (r *ChargingSessionRepository) GetUserSessions(userID uuid.UUID, page, pageSize int) ([]*model.ChargingSession, int, error) {
	// 获取总数
	var total int
	err := r.db.QueryRow("SELECT COUNT(*) FROM charging_sessions WHERE user_id = $1", userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := `
		SELECT id, request_id, user_id, pile_id, queue_number, requested_capacity, 
			   actual_capacity, start_time, end_time, status, duration, created_at
		FROM charging_sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var sessions []*model.ChargingSession
	for rows.Next() {
		var session model.ChargingSession
		var endTime sql.NullTime

		err := rows.Scan(
			&session.ID,
			&session.RequestID,
			&session.UserID,
			&session.PileID,
			&session.QueueNumber,
			&session.RequestedCapacity,
			&session.ActualCapacity,
			&session.StartTime,
			&endTime,
			&session.Status,
			&session.Duration,
			&session.CreatedAt,
		)

		if err != nil {
			return nil, 0, err
		}

		if endTime.Valid {
			session.EndTime = &endTime.Time
		}

		sessions = append(sessions, &session)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}

// CountCompletedSessions 计算特定时期内完成的会话数
func (r *ChargingSessionRepository) CountCompletedSessions(startTime, endTime time.Time, pileID *string) (int, error) {
	var query string
	var args []any

	if pileID != nil {
		query = `
			SELECT COUNT(*)
			FROM charging_sessions
			WHERE status IN ('completed', 'interrupted')
				AND start_time >= $1 AND (end_time <= $2 OR end_time IS NULL)
				AND pile_id = $3
		`
		args = []any{startTime, endTime, *pileID}
	} else {
		query = `
			SELECT COUNT(*)
			FROM charging_sessions
			WHERE status IN ('completed', 'interrupted')
				AND start_time >= $1 AND (end_time <= $2 OR end_time IS NULL)
		`
		args = []any{startTime, endTime}
	}

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	return count, err
}

// GetSessionsInPeriod 获取特定时期内的会话
func (r *ChargingSessionRepository) GetSessionsInPeriod(startTime, endTime time.Time, pileID *string) ([]*model.ChargingSession, error) {
	var query string
	var args []any

	if pileID != nil {
		query = `
			SELECT id, request_id, user_id, pile_id, queue_number, requested_capacity, 
				   actual_capacity, start_time, end_time, status, duration, created_at
			FROM charging_sessions
			WHERE start_time >= $1 AND (end_time <= $2 OR end_time IS NULL)
				AND pile_id = $3
			ORDER BY start_time ASC
		`
		args = []any{startTime, endTime, *pileID}
	} else {
		query = `
			SELECT id, request_id, user_id, pile_id, queue_number, requested_capacity, 
				   actual_capacity, start_time, end_time, status, duration, created_at
			FROM charging_sessions
			WHERE start_time >= $1 AND (end_time <= $2 OR end_time IS NULL)
			ORDER BY start_time ASC
		`
		args = []any{startTime, endTime}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*model.ChargingSession
	for rows.Next() {
		var session model.ChargingSession
		var endTime sql.NullTime

		err := rows.Scan(
			&session.ID,
			&session.RequestID,
			&session.UserID,
			&session.PileID,
			&session.QueueNumber,
			&session.RequestedCapacity,
			&session.ActualCapacity,
			&session.StartTime,
			&endTime,
			&session.Status,
			&session.Duration,
			&session.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		if endTime.Valid {
			session.EndTime = &endTime.Time
		}

		sessions = append(sessions, &session)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

// GetSessionsByPileID 获取指定充电桩在特定时期内的会话
func (r *ChargingSessionRepository) GetSessionsByPileID(pileID string, startTime, endTime time.Time) ([]*model.ChargingSession, error) {
	query := `
		SELECT id, request_id, user_id, pile_id, queue_number, requested_capacity, 
			   actual_capacity, start_time, end_time, status, duration, created_at
		FROM charging_sessions
		WHERE pile_id = $1 AND start_time >= $2 AND (end_time <= $3 OR end_time IS NULL)
		ORDER BY start_time ASC
	`

	rows, err := r.db.Query(query, pileID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*model.ChargingSession
	for rows.Next() {
		var session model.ChargingSession
		var endTime sql.NullTime

		err := rows.Scan(
			&session.ID,
			&session.RequestID,
			&session.UserID,
			&session.PileID,
			&session.QueueNumber,
			&session.RequestedCapacity,
			&session.ActualCapacity,
			&session.StartTime,
			&endTime,
			&session.Status,
			&session.Duration,
			&session.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		if endTime.Valid {
			session.EndTime = &endTime.Time
		}

		sessions = append(sessions, &session)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}
