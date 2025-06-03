package repository

import (
	"database/sql"
	"time"

	"backend/internal/model"

	"github.com/google/uuid"
)

// QueueRepository 队列仓库
type QueueRepository struct {
	db *sql.DB
}

// NewQueueRepository 创建队列仓库
func NewQueueRepository(db *sql.DB) *QueueRepository {
	return &QueueRepository{
		db: db,
	}
}

// AddToQueue 添加到队列
func (r *QueueRepository) AddToQueue(item *model.QueueItem) error {
	query := `
		INSERT INTO queue_status (id, pile_id, position, request_id, user_id, queue_number, entered_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(
		query,
		uuid.New(), // 生成新的UUID作为记录ID
		item.PileID,
		item.Position,
		item.RequestID,
		item.UserID,
		item.QueueNumber,
		time.Now().UTC(),
	)

	return err
}

// RemoveFromQueue 从队列中删除
func (r *QueueRepository) RemoveFromQueue(requestID uuid.UUID) error {
	query := `DELETE FROM queue_status WHERE request_id = $1`
	_, err := r.db.Exec(query, requestID)
	return err
}

// UpdateQueuePosition 更新队列位置
func (r *QueueRepository) UpdateQueuePosition(pileID string, requestID uuid.UUID, position int) error {
	query := `
		UPDATE queue_status
		SET position = $1
		WHERE pile_id = $2 AND request_id = $3
	`

	_, err := r.db.Exec(query, position, pileID, requestID)
	return err
}

// SetStartCharging 设置开始充电
func (r *QueueRepository) SetStartCharging(requestID uuid.UUID) error {
	query := `
		UPDATE queue_status
		SET started_at = $1
		WHERE request_id = $2
	`

	_, err := r.db.Exec(query, time.Now().UTC(), requestID)
	return err
}

// GetQueueItemsByPile 获取充电桩的队列项
func (r *QueueRepository) GetQueueItemsByPile(pileID string) ([]*model.QueueItem, error) {
	query := `
		SELECT qs.pile_id, qs.position, qs.request_id, qs.user_id, qs.queue_number, 
		       qs.entered_at, cr.charging_mode, cr.requested_capacity
		FROM queue_status qs
		INNER JOIN charging_requests cr ON qs.request_id = cr.id
		WHERE qs.pile_id = $1
		ORDER BY qs.position ASC
	`

	rows, err := r.db.Query(query, pileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.QueueItem
	for rows.Next() {
		var item model.QueueItem
		err := rows.Scan(
			&item.PileID,
			&item.Position,
			&item.RequestID,
			&item.UserID,
			&item.QueueNumber,
			&item.EnterTime,
			&item.ChargingMode,
			&item.RequestedCapacity,
		)

		if err != nil {
			return nil, err
		}

		// 计算等待时间（当前时间减去进入队列的时间，单位秒）
		item.WaitTime = int(time.Since(item.EnterTime).Seconds())
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// GetQueueStatus 获取队列状态
func (r *QueueRepository) GetQueueStatus() (*model.QueueStatus, error) {
	// 获取快充队列
	fastQuery := `
		SELECT qs.user_id, qs.queue_number, qs.request_id, 
		       cr.charging_mode, cr.requested_capacity, qs.entered_at
		FROM queue_status qs
		INNER JOIN charging_requests cr ON qs.request_id = cr.id
		WHERE cr.charging_mode = 'fast' AND cr.status = 'waiting'
		ORDER BY qs.entered_at ASC
	`

	fastRows, err := r.db.Query(fastQuery)
	if err != nil {
		return nil, err
	}
	defer fastRows.Close()

	var fastQueue []model.QueueItem
	for fastRows.Next() {
		var item model.QueueItem
		err := fastRows.Scan(
			&item.UserID,
			&item.QueueNumber,
			&item.RequestID,
			&item.ChargingMode,
			&item.RequestedCapacity,
			&item.EnterTime,
		)

		if err != nil {
			return nil, err
		}

		// 计算等待时间（当前时间减去进入队列的时间，单位秒）
		item.WaitTime = int(time.Since(item.EnterTime).Seconds())
		item.Position = len(fastQueue) + 1
		fastQueue = append(fastQueue, item)
	}

	if err := fastRows.Err(); err != nil {
		return nil, err
	}

	// 获取慢充队列
	slowQuery := `
		SELECT qs.user_id, qs.queue_number, qs.request_id, 
		       cr.charging_mode, cr.requested_capacity, qs.entered_at
		FROM queue_status qs
		INNER JOIN charging_requests cr ON qs.request_id = cr.id
		WHERE cr.charging_mode = 'slow' AND cr.status = 'waiting'
		ORDER BY qs.entered_at ASC
	`

	slowRows, err := r.db.Query(slowQuery)
	if err != nil {
		return nil, err
	}
	defer slowRows.Close()

	var slowQueue []model.QueueItem
	for slowRows.Next() {
		var item model.QueueItem
		err := slowRows.Scan(
			&item.UserID,
			&item.QueueNumber,
			&item.RequestID,
			&item.ChargingMode,
			&item.RequestedCapacity,
			&item.EnterTime,
		)

		if err != nil {
			return nil, err
		}

		// 计算等待时间（当前时间减去进入队列的时间，单位秒）
		item.WaitTime = int(time.Since(item.EnterTime).Seconds())
		item.Position = len(slowQueue) + 1
		slowQueue = append(slowQueue, item)
	}

	if err := slowRows.Err(); err != nil {
		return nil, err
	}

	// 计算等候区可用车位数
	var totalCount int
	err = r.db.QueryRow(`SELECT COUNT(*) FROM charging_requests WHERE status = 'waiting'`).Scan(&totalCount)
	if err != nil {
		return nil, err
	}

	var maxWaitingSize int
	err = r.db.QueryRow(`SELECT config_value FROM system_config WHERE config_key = 'waiting_area_size'`).Scan(&maxWaitingSize)
	if err != nil {
		maxWaitingSize = 6 // 默认值
	}

	availableSlots := maxWaitingSize - totalCount
	if availableSlots < 0 {
		availableSlots = 0
	}

	return &model.QueueStatus{
		FastQueue:      fastQueue,
		SlowQueue:      slowQueue,
		AvailableSlots: availableSlots,
	}, nil
}

// GetUserPosition 获取用户在队列中的位置
func (r *QueueRepository) GetUserPosition(userID uuid.UUID) (*model.UserPosition, error) {
	// 首先检查用户是否在等待队列中
	waitingQuery := `
		SELECT cr.id, cr.queue_number, cr.charging_mode, cr.status
		FROM charging_requests cr
		WHERE cr.user_id = $1 AND cr.status = 'waiting'
		ORDER BY cr.created_at DESC
		LIMIT 1
	`

	var position model.UserPosition
	position.UserID = userID

	var requestID uuid.UUID
	err := r.db.QueryRow(waitingQuery, userID).Scan(
		&requestID,
		&position.QueueNumber,
		&position.ChargingMode,
		&position.Status,
	)

	if err == nil {
		// 用户在等待队列中，计算位置
		var count int
		var query string

		if position.ChargingMode == model.ChargingModeFast {
			query = `
				SELECT COUNT(*) 
				FROM charging_requests 
				WHERE charging_mode = 'fast' AND status = 'waiting' AND created_at < (
					SELECT created_at FROM charging_requests WHERE id = $1
				)
			`
		} else {
			query = `
				SELECT COUNT(*) 
				FROM charging_requests 
				WHERE charging_mode = 'slow' AND status = 'waiting' AND created_at < (
					SELECT created_at FROM charging_requests WHERE id = $1
				)
			`
		}

		err := r.db.QueryRow(query, requestID).Scan(&count)
		if err == nil {
			position.Position = count + 1
			return &position, nil
		}
	}

	// 检查用户是否在充电桩队列中
	queuedQuery := `
		SELECT cr.id, cr.queue_number, cr.charging_mode, cr.status, cr.pile_id, cr.queue_position
		FROM charging_requests cr
		WHERE cr.user_id = $1 AND cr.status IN ('queued', 'charging')
		ORDER BY cr.created_at DESC
		LIMIT 1
	`

	err = r.db.QueryRow(queuedQuery, userID).Scan(
		&requestID,
		&position.QueueNumber,
		&position.ChargingMode,
		&position.Status,
		&position.AssignedPile,
		&position.QueuePosition,
	)

	if err == nil {
		// 用户在充电桩队列中
		position.Position = 0 // 已经离开等待队列

		// 计算估计等待时间
		if position.Status == model.RequestStatusQueued && position.QueuePosition > 0 {
			// 获取前面的车辆还需要充电多久
			var waitTimeSeconds int
			query := `
				SELECT COALESCE(SUM(
					CASE 
						WHEN cr.charging_mode = 'fast' THEN cr.requested_capacity / 30.0 * 3600
						ELSE cr.requested_capacity / 7.0 * 3600
					END
				), 0) as wait_time
				FROM charging_requests cr
				WHERE cr.pile_id = $1 AND cr.queue_position < $2 AND cr.status IN ('queued', 'charging')
			`
			err := r.db.QueryRow(query, position.AssignedPile, position.QueuePosition).Scan(&waitTimeSeconds)
			if err == nil {
				position.WaitingTime = waitTimeSeconds
			}
		}

		return &position, nil
	}

	// 用户没有活跃的充电请求
	return nil, sql.ErrNoRows
}
