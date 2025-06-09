package repository

import (
	"database/sql"
	"fmt"
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

// RemoveFromQueueAndDecrementPile 使用事务从队列中删除请求并将充电桩队列长度减1
func (r *QueueRepository) RemoveFromQueueAndDecrementPile(requestID uuid.UUID, pileID string) error {
	// 开始事务
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 1. 从队列中删除请求
	deleteQuery := `DELETE FROM queue_status WHERE request_id = $1`
	result, err := tx.Exec(deleteQuery, requestID)
	if err != nil {
		return fmt.Errorf("从队列中删除请求失败: %w", err)
	}

	// 检查是否真的删除了记录
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除的行数失败: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("未找到要删除的队列记录，requestID: %s", requestID.String())
	}

	// 2. 将充电桩队列长度减1（确保不小于0）
	updateQuery := `
		UPDATE charging_piles 
		SET queue_length = GREATEST(queue_length - 1, 0), updated_at = $1
		WHERE id = $2
	`
	result, err = tx.Exec(updateQuery, time.Now().UTC(), pileID)
	if err != nil {
		return fmt.Errorf("更新充电桩队列长度失败: %w", err)
	}

	// 检查是否真的更新了充电桩
	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取更新的行数失败: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("未找到要更新的充电桩，pileID: %s", pileID)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
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
