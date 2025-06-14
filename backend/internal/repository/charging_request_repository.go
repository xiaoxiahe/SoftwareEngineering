package repository

import (
	"database/sql"
	"errors"
	"time"

	"backend/internal/model"

	"github.com/google/uuid"
)

// ChargingRequestRepository 充电请求仓库
type ChargingRequestRepository struct {
	db *sql.DB
}

// NewChargingRequestRepository 创建充电请求仓库
func NewChargingRequestRepository(db *sql.DB) *ChargingRequestRepository {
	return &ChargingRequestRepository{
		db: db,
	}
}

// Create 创建充电请求
func (r *ChargingRequestRepository) Create(request *model.ChargingRequest) (*model.ChargingRequest, error) {
	// 检查用户是否已有活跃充电请求
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) 
		FROM charging_requests 
		WHERE user_id = $1 AND status IN ('waiting', 'queued', 'charging')
	`, request.UserID).Scan(&count)

	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, errors.New("用户已有活跃的充电请求")
	}
	// 插入新的充电请求
	query := `
		INSERT INTO charging_requests 
		(id, user_id, charging_mode, requested_capacity, queue_number, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, charging_mode, requested_capacity, queue_number, status, 
		          pile_id, queue_position, estimated_wait_time, created_at, updated_at
	`
	now := time.Now().UTC()
	var newRequest model.ChargingRequest
	var pileID sql.NullString
	var queuePosition sql.NullInt64
	var estimatedWaitTime sql.NullInt64
	err = r.db.QueryRow(
		query,
		request.ID,
		request.UserID,
		request.ChargingMode,
		request.RequestedCapacity,
		request.QueueNumber,
		request.Status,
		now,
		now,
	).Scan(
		&newRequest.ID,
		&newRequest.UserID,
		&newRequest.ChargingMode,
		&newRequest.RequestedCapacity,
		&newRequest.QueueNumber,
		&newRequest.Status,
		&pileID,
		&queuePosition,
		&estimatedWaitTime,
		&newRequest.CreatedAt,
		&newRequest.UpdatedAt,
	)

	// 处理可能为NULL的字段
	if pileID.Valid {
		newRequest.PileID = pileID.String
	}
	if queuePosition.Valid {
		newRequest.QueuePosition = int(queuePosition.Int64)
	}
	if estimatedWaitTime.Valid {
		newRequest.EstimatedWaitTime = int(estimatedWaitTime.Int64)
	}

	if err != nil {
		return nil, err
	}

	return &newRequest, nil
}

// GetByID 通过ID获取充电请求
func (r *ChargingRequestRepository) GetByID(id uuid.UUID) (*model.ChargingRequest, error) {
	query := `
		SELECT id, user_id, charging_mode, requested_capacity, queue_number, status, 
		       pile_id, queue_position, estimated_wait_time, created_at, updated_at
		FROM charging_requests
		WHERE id = $1
	`

	var request model.ChargingRequest
	var pileID sql.NullString
	var queuePosition sql.NullInt64
	var estimatedWaitTime sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(
		&request.ID,
		&request.UserID,
		&request.ChargingMode,
		&request.RequestedCapacity,
		&request.QueueNumber,
		&request.Status,
		&pileID,
		&queuePosition,
		&estimatedWaitTime,
		&request.CreatedAt,
		&request.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("充电请求不存在")
		}
		return nil, err
	}

	if pileID.Valid {
		request.PileID = pileID.String
	}
	if queuePosition.Valid {
		request.QueuePosition = int(queuePosition.Int64)
	}
	if estimatedWaitTime.Valid {
		request.EstimatedWaitTime = int(estimatedWaitTime.Int64)
	}

	return &request, nil
}

// GetActiveRequestByUserID 通过用户ID获取活跃请求
func (r *ChargingRequestRepository) GetActiveRequestByUserID(userID uuid.UUID) (*model.ChargingRequest, error) {
	query := `
		SELECT id, user_id, charging_mode, requested_capacity, queue_number, status, 
		       pile_id, queue_position, estimated_wait_time, created_at, updated_at
		FROM charging_requests
		WHERE user_id = $1 AND status IN ('waiting', 'queued', 'charging')
		ORDER BY created_at DESC
		LIMIT 1
	`

	var request model.ChargingRequest
	var pileID sql.NullString
	var queuePosition sql.NullInt64
	var estimatedWaitTime sql.NullInt64

	err := r.db.QueryRow(query, userID).Scan(
		&request.ID,
		&request.UserID,
		&request.ChargingMode,
		&request.RequestedCapacity,
		&request.QueueNumber,
		&request.Status,
		&pileID,
		&queuePosition,
		&estimatedWaitTime,
		&request.CreatedAt,
		&request.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("用户没有活跃的充电请求")
		}
		return nil, err
	}

	if pileID.Valid {
		request.PileID = pileID.String
	}
	if queuePosition.Valid {
		request.QueuePosition = int(queuePosition.Int64)
	}
	if estimatedWaitTime.Valid {
		request.EstimatedWaitTime = int(estimatedWaitTime.Int64)
	}

	return &request, nil
}

// GetLatestRequestByUserID 通过用户ID获取最新请求
func (r *ChargingRequestRepository) GetLatestRequestByUserID(userID uuid.UUID) (*model.ChargingRequest, error) {
	query := `
		SELECT id, user_id, charging_mode, requested_capacity, queue_number, status, 
		       pile_id, queue_position, estimated_wait_time, created_at, updated_at
		FROM charging_requests
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var request model.ChargingRequest
	var pileID sql.NullString
	var queuePosition sql.NullInt64
	var estimatedWaitTime sql.NullInt64

	err := r.db.QueryRow(query, userID).Scan(
		&request.ID,
		&request.UserID,
		&request.ChargingMode,
		&request.RequestedCapacity,
		&request.QueueNumber,
		&request.Status,
		&pileID,
		&queuePosition,
		&estimatedWaitTime,
		&request.CreatedAt,
		&request.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("用户没有充电请求记录")
		}
		return nil, err
	}

	if pileID.Valid {
		request.PileID = pileID.String
	}
	if queuePosition.Valid {
		request.QueuePosition = int(queuePosition.Int64)
	}
	if estimatedWaitTime.Valid {
		request.EstimatedWaitTime = int(estimatedWaitTime.Int64)
	}

	return &request, nil
}

// UpdateRequestStatus 更新请求状态
func (r *ChargingRequestRepository) UpdateRequestStatus(id uuid.UUID, status model.RequestStatus) error {
	query := `
		UPDATE charging_requests
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(query, status, time.Now().UTC(), id)
	return err
}

// AssignToPile 将请求分配给充电桩
func (r *ChargingRequestRepository) AssignToPile(id uuid.UUID, pileID string, queuePosition int, waitTime int, status model.RequestStatus) error {
	query := `
		UPDATE charging_requests
		SET pile_id = $1, queue_position = $2, estimated_wait_time = $3, status = $4, updated_at = $5
		WHERE id = $6
	`

	// 如果pileID为空字符串，使用NULL值
	var pileIDParam interface{}
	if pileID == "" {
		pileIDParam = nil
	} else {
		pileIDParam = pileID
	}

	_, err := r.db.Exec(query, pileIDParam, queuePosition, waitTime, status, time.Now().UTC(), id)
	return err
}

// UpdateRequest 更新充电请求
func (r *ChargingRequestRepository) UpdateRequest(request *model.ChargingRequest) error {
	query := `
		UPDATE charging_requests
		SET charging_mode = $1, requested_capacity = $2, queue_number = $3, 
		    pile_id = $4, queue_position = $5, status = $6, estimated_wait_time = $7, 
		    updated_at = $8
		WHERE id = $9
	`

	var pileID any = nil
	if request.PileID != "" {
		pileID = request.PileID
	}

	_, err := r.db.Exec(
		query,
		request.ChargingMode,
		request.RequestedCapacity,
		request.QueueNumber,
		pileID,
		request.QueuePosition,
		request.Status,
		request.EstimatedWaitTime,
		time.Now().UTC(),
		request.ID,
	)

	return err
}

// GetWaitingRequestsByMode 获取特定模式的等待请求
func (r *ChargingRequestRepository) GetWaitingRequestsByMode(mode model.ChargingMode) ([]*model.ChargingRequest, error) {
	query := `
		SELECT id, user_id, charging_mode, requested_capacity, queue_number, status, 
		       pile_id, queue_position, estimated_wait_time, created_at, updated_at
		FROM charging_requests
		WHERE charging_mode = $1 AND status = 'waiting'
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, mode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*model.ChargingRequest
	for rows.Next() {
		var request model.ChargingRequest
		var pileID sql.NullString
		var queuePosition sql.NullInt64
		var estimatedWaitTime sql.NullInt64

		err := rows.Scan(
			&request.ID,
			&request.UserID,
			&request.ChargingMode,
			&request.RequestedCapacity,
			&request.QueueNumber,
			&request.Status,
			&pileID,
			&queuePosition,
			&estimatedWaitTime,
			&request.CreatedAt,
			&request.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if pileID.Valid {
			request.PileID = pileID.String
		}
		if queuePosition.Valid {
			request.QueuePosition = int(queuePosition.Int64)
		}
		if estimatedWaitTime.Valid {
			request.EstimatedWaitTime = int(estimatedWaitTime.Int64)
		}

		requests = append(requests, &request)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

// CountWaitingRequests 计算等待请求的数量
func (r *ChargingRequestRepository) CountWaitingRequests() (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM charging_requests WHERE status = 'waiting'`).Scan(&count)
	return count, err
}

// GetQueuedRequestsByPile 获取指定充电桩的排队请求（按优先级和时间排序）
func (r *ChargingRequestRepository) GetQueuedRequestsByPile(pileID string) ([]*model.ChargingRequest, error) {
	query := `
		SELECT id, user_id, charging_mode, requested_capacity, queue_number, status, 
		       pile_id, queue_position, estimated_wait_time, created_at, updated_at
		FROM charging_requests
		WHERE pile_id = $1 AND status = 'queued'
		ORDER BY queue_position ASC, created_at ASC
	`

	rows, err := r.db.Query(query, pileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*model.ChargingRequest
	for rows.Next() {
		var request model.ChargingRequest
		var pID sql.NullString
		var queuePosition sql.NullInt64
		var estimatedWaitTime sql.NullInt64

		err := rows.Scan(
			&request.ID,
			&request.UserID,
			&request.ChargingMode,
			&request.RequestedCapacity,
			&request.QueueNumber,
			&request.Status,
			&pID,
			&queuePosition,
			&estimatedWaitTime,
			&request.CreatedAt,
			&request.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if pID.Valid {
			request.PileID = pID.String
		}
		if queuePosition.Valid {
			request.QueuePosition = int(queuePosition.Int64)
		}
		if estimatedWaitTime.Valid {
			request.EstimatedWaitTime = int(estimatedWaitTime.Int64)
		}

		requests = append(requests, &request)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

// GetRequestsByPile 获取特定充电桩的请求
func (r *ChargingRequestRepository) GetRequestsByPile(pileID string) ([]*model.ChargingRequest, error) {
	query := `
		SELECT id, user_id, charging_mode, requested_capacity, queue_number, status, 
		       pile_id, queue_position, estimated_wait_time, created_at, updated_at
		FROM charging_requests
		WHERE pile_id = $1 AND status IN ('queued', 'charging')
		ORDER BY queue_position ASC
	`

	rows, err := r.db.Query(query, pileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*model.ChargingRequest
	for rows.Next() {
		var request model.ChargingRequest
		var pID sql.NullString
		var queuePosition sql.NullInt64
		var estimatedWaitTime sql.NullInt64

		err := rows.Scan(
			&request.ID,
			&request.UserID,
			&request.ChargingMode,
			&request.RequestedCapacity,
			&request.QueueNumber,
			&request.Status,
			&pID,
			&queuePosition,
			&estimatedWaitTime,
			&request.CreatedAt,
			&request.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if pID.Valid {
			request.PileID = pID.String
		}
		if queuePosition.Valid {
			request.QueuePosition = int(queuePosition.Int64)
		}
		if estimatedWaitTime.Valid {
			request.EstimatedWaitTime = int(estimatedWaitTime.Int64)
		}

		requests = append(requests, &request)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

// GetUserRequests 获取用户的充电请求历史
func (r *ChargingRequestRepository) GetUserRequests(userID uuid.UUID, page, pageSize int) ([]*model.ChargingRequest, int, error) {
	// 获取总数
	var total int
	err := r.db.QueryRow("SELECT COUNT(*) FROM charging_requests WHERE user_id = $1", userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	// 分页查询
	offset := (page - 1) * pageSize
	query := `
		SELECT id, user_id, charging_mode, requested_capacity, queue_number, status, 
		       pile_id, queue_position, estimated_wait_time, created_at, updated_at
		FROM charging_requests
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var requests []*model.ChargingRequest
	for rows.Next() {
		var request model.ChargingRequest
		var pileID sql.NullString
		var queuePosition sql.NullInt64
		var estimatedWaitTime sql.NullInt64

		err := rows.Scan(
			&request.ID,
			&request.UserID,
			&request.ChargingMode,
			&request.RequestedCapacity,
			&request.QueueNumber,
			&request.Status,
			&pileID,
			&queuePosition,
			&estimatedWaitTime,
			&request.CreatedAt,
			&request.UpdatedAt,
		)

		if err != nil {
			return nil, 0, err
		}

		if pileID.Valid {
			request.PileID = pileID.String
		}
		if queuePosition.Valid {
			request.QueuePosition = int(queuePosition.Int64)
		}
		if estimatedWaitTime.Valid {
			request.EstimatedWaitTime = int(estimatedWaitTime.Int64)
		}

		requests = append(requests, &request)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}
