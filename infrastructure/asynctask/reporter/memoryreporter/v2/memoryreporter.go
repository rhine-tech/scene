package v2

import (
	"context"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/asynctask"
	"github.com/rhine-tech/scene/model"
)

const (
	defaultLimit           = 20
	maxLimit               = 200
	shardCount             = 64
	defaultEventTTL        = 24 * time.Hour
	defaultMaxTasks        = 10000
	defaultCleanupInterval = time.Minute
)

type Config struct {
	// EventTTL controls how long the latest event snapshot is retained.
	// A negative value disables TTL cleanup.
	EventTTL time.Duration

	// MaxTasks limits the number of task snapshots retained in memory.
	// A negative value disables count-based cleanup.
	MaxTasks int

	// CleanupInterval controls how often ReportTaskEvent triggers opportunistic cleanup.
	// A negative value disables opportunistic cleanup.
	CleanupInterval time.Duration
}

type MemoryReporter struct {
	config Config
	shards [shardCount]memoryShard
	count  atomic.Int64

	nextCleanupAt  atomic.Int64
	cleanupRunning atomic.Bool
}

type memoryShard struct {
	mu     sync.RWMutex
	events map[string]asynctask.TaskEvent
}

type eventSnapshot struct {
	taskID    string
	updatedAt time.Time
}

func NewMemoryReporter(config ...Config) *MemoryReporter {
	cfg := Config{}
	if len(config) > 0 {
		cfg = config[0]
	}

	reporter := &MemoryReporter{config: normalizeConfig(cfg)}
	for i := range reporter.shards {
		reporter.shards[i].events = make(map[string]asynctask.TaskEvent)
	}
	return reporter
}

func (m *MemoryReporter) ImplName() scene.ImplName {
	return asynctask.Lens.ImplName("TaskReporter", "memory.v2")
}

func (m *MemoryReporter) ReportTaskEvent(_ context.Context, event asynctask.TaskEvent) error {
	taskID := strings.TrimSpace(event.TaskID)
	if taskID == "" {
		return asynctask.ErrInvalidQueueTask
	}
	event.TaskID = taskID

	now := time.Now()
	if event.UpdatedAt.IsZero() {
		event.UpdatedAt = now
	}

	shard := m.shard(taskID)
	shard.mu.Lock()
	current, exists := shard.events[taskID]
	if exists && event.UpdatedAt.Before(current.UpdatedAt) {
		shard.mu.Unlock()
		m.cleanupIfNeeded(now)
		return nil
	}
	if !exists {
		current = asynctask.TaskEvent{TaskID: taskID}
		m.count.Add(1)
	}
	shard.events[taskID] = mergeEvent(current, event)
	shard.mu.Unlock()

	if !exists {
		m.cleanupAfterInsert(now, taskID)
	} else {
		m.cleanupIfNeeded(now)
	}
	return nil
}

func (m *MemoryReporter) Cleanup(_ context.Context) int64 {
	return m.cleanup(time.Now())
}

func (m *MemoryReporter) ListTasks(_ context.Context, query asynctask.TaskQuery) (*model.PaginationResult[asynctask.TaskEvent], error) {
	filter := newTaskFilter(query)
	records := m.snapshot(filter.match)

	sort.Slice(records, func(i, j int) bool {
		if records[i].UpdatedAt.Equal(records[j].UpdatedAt) {
			return records[i].TaskID < records[j].TaskID
		}
		return records[i].UpdatedAt.After(records[j].UpdatedAt)
	})

	offset, limit := normalizePagination(query.Offset, query.Limit, len(records))
	end := offset + limit
	if end > len(records) {
		end = len(records)
	}

	result := records[offset:end]
	return &model.PaginationResult[asynctask.TaskEvent]{
		Total:   int64(len(records)),
		Offset:  int64(offset),
		Results: result,
		Count:   int64(len(result)),
	}, nil
}

func (m *MemoryReporter) GetTask(_ context.Context, taskID string) (asynctask.TaskEvent, error) {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return asynctask.TaskEvent{}, asynctask.ErrTaskNotFound.WithDetailStr(taskID)
	}

	shard := m.shard(taskID)
	shard.mu.RLock()
	record, ok := shard.events[taskID]
	shard.mu.RUnlock()
	if !ok {
		return asynctask.TaskEvent{}, asynctask.ErrTaskNotFound.WithDetailStr(taskID)
	}
	return record, nil
}

func (m *MemoryReporter) GetQueueStats(_ context.Context, queue string) (asynctask.QueueStats, error) {
	stats := asynctask.QueueStats{
		Queue:    queue,
		ByStatus: make(map[asynctask.QueueTaskStatus]int64),
	}

	for i := range m.shards {
		shard := &m.shards[i]
		shard.mu.RLock()
		for _, record := range shard.events {
			if queue != "" && record.Queue != queue {
				continue
			}
			stats.Total++
			stats.ByStatus[record.Status]++
		}
		shard.mu.RUnlock()
	}
	return stats, nil
}

func (m *MemoryReporter) snapshot(match func(asynctask.TaskEvent) bool) []asynctask.TaskEvent {
	records := make([]asynctask.TaskEvent, 0, m.count.Load())
	for i := range m.shards {
		shard := &m.shards[i]
		shard.mu.RLock()
		for _, record := range shard.events {
			if match(record) {
				records = append(records, record)
			}
		}
		shard.mu.RUnlock()
	}
	return records
}

func (m *MemoryReporter) shard(taskID string) *memoryShard {
	return &m.shards[int(fnv32a(taskID)&uint32(shardCount-1))]
}

func fnv32a(s string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

func normalizeConfig(config Config) Config {
	if config.EventTTL == 0 {
		config.EventTTL = defaultEventTTL
	}
	if config.MaxTasks == 0 {
		config.MaxTasks = defaultMaxTasks
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = defaultCleanupInterval
	}
	return config
}

func (m *MemoryReporter) cleanupIfNeeded(now time.Time) {
	if m.config.CleanupInterval < 0 {
		return
	}
	next := m.nextCleanupAt.Load()
	nowNano := now.UnixNano()
	if next > nowNano {
		return
	}
	if !m.nextCleanupAt.CompareAndSwap(next, now.Add(m.config.CleanupInterval).UnixNano()) {
		return
	}
	m.cleanup(now)
}

func (m *MemoryReporter) cleanup(now time.Time) int64 {
	if !m.cleanupRunning.CompareAndSwap(false, true) {
		return 0
	}
	defer m.cleanupRunning.Store(false)

	removed := m.cleanupExpired(now)
	removed += m.cleanupOverflow()
	return removed
}

func (m *MemoryReporter) cleanupExpired(now time.Time) int64 {
	if m.config.EventTTL < 0 {
		return 0
	}
	cutoff := now.Add(-m.config.EventTTL)
	var removed int64
	for i := range m.shards {
		shard := &m.shards[i]
		shard.mu.Lock()
		for taskID, record := range shard.events {
			if record.UpdatedAt.Before(cutoff) {
				delete(shard.events, taskID)
				removed++
			}
		}
		shard.mu.Unlock()
	}
	if removed > 0 {
		m.count.Add(-removed)
	}
	return removed
}

func (m *MemoryReporter) cleanupAfterInsert(now time.Time, taskID string) {
	if m.config.MaxTasks < 0 {
		m.cleanupIfNeeded(now)
		return
	}
	maxTasks := int64(m.config.MaxTasks)
	if maxTasks <= 0 || m.count.Load() <= maxTasks {
		m.cleanupIfNeeded(now)
		return
	}
	m.cleanup(now)
	if m.count.Load() > maxTasks {
		m.deleteTask(taskID)
	}
}

func (m *MemoryReporter) cleanupOverflow() int64 {
	if m.config.MaxTasks < 0 {
		return 0
	}
	maxTasks := int64(m.config.MaxTasks)
	if maxTasks <= 0 || m.count.Load() <= maxTasks {
		return 0
	}

	snapshots := m.snapshotEvents()
	sort.Slice(snapshots, func(i, j int) bool {
		if snapshots[i].updatedAt.Equal(snapshots[j].updatedAt) {
			return snapshots[i].taskID < snapshots[j].taskID
		}
		return snapshots[i].updatedAt.Before(snapshots[j].updatedAt)
	})

	var removed int64
	for _, snapshot := range snapshots {
		if m.count.Load() <= maxTasks {
			break
		}
		if m.deleteTaskIfUpdatedAt(snapshot.taskID, snapshot.updatedAt) {
			removed++
		}
	}
	return removed
}

func (m *MemoryReporter) snapshotEvents() []eventSnapshot {
	snapshots := make([]eventSnapshot, 0, m.count.Load())
	for i := range m.shards {
		shard := &m.shards[i]
		shard.mu.RLock()
		for taskID, record := range shard.events {
			snapshots = append(snapshots, eventSnapshot{
				taskID:    taskID,
				updatedAt: record.UpdatedAt,
			})
		}
		shard.mu.RUnlock()
	}
	return snapshots
}

func (m *MemoryReporter) deleteTask(taskID string) bool {
	shard := m.shard(taskID)
	shard.mu.Lock()
	_, exists := shard.events[taskID]
	if exists {
		delete(shard.events, taskID)
	}
	shard.mu.Unlock()
	if exists {
		m.count.Add(-1)
	}
	return exists
}

func (m *MemoryReporter) deleteTaskIfUpdatedAt(taskID string, updatedAt time.Time) bool {
	shard := m.shard(taskID)
	shard.mu.Lock()
	record, exists := shard.events[taskID]
	if exists && record.UpdatedAt.Equal(updatedAt) {
		delete(shard.events, taskID)
	} else {
		exists = false
	}
	shard.mu.Unlock()
	if exists {
		m.count.Add(-1)
	}
	return exists
}

func mergeEvent(record asynctask.TaskEvent, event asynctask.TaskEvent) asynctask.TaskEvent {
	if event.Queue != "" {
		record.Queue = event.Queue
	}
	if event.TaskType != "" {
		record.TaskType = event.TaskType
	}
	if event.Key != "" {
		record.Key = event.Key
	}
	if event.Status != "" {
		record.Status = event.Status
	}
	if event.Attempt != 0 || record.Attempt == 0 {
		record.Attempt = event.Attempt
	}
	if event.Progress != 0 || record.Progress == 0 {
		record.Progress = event.Progress
	}
	if event.Message != "" {
		record.Message = event.Message
	}
	record.Error = event.Error
	record.UpdatedAt = event.UpdatedAt
	return record
}

type taskFilter struct {
	query   asynctask.TaskQuery
	keyword string
	status  map[asynctask.QueueTaskStatus]struct{}
}

func newTaskFilter(query asynctask.TaskQuery) taskFilter {
	filter := taskFilter{
		query:   query,
		keyword: strings.ToLower(strings.TrimSpace(query.Keyword)),
	}
	if len(query.Status) > 0 {
		filter.status = make(map[asynctask.QueueTaskStatus]struct{}, len(query.Status))
		for _, status := range query.Status {
			filter.status[status] = struct{}{}
		}
	}
	return filter
}

func (f taskFilter) match(record asynctask.TaskEvent) bool {
	query := f.query
	if query.TaskID != "" && record.TaskID != query.TaskID {
		return false
	}
	if query.Queue != "" && record.Queue != query.Queue {
		return false
	}
	if query.TaskType != "" && record.TaskType != query.TaskType {
		return false
	}
	if query.Key != "" && record.Key != query.Key {
		return false
	}
	if len(f.status) > 0 {
		if _, ok := f.status[record.Status]; !ok {
			return false
		}
	}
	if f.keyword != "" && !containsKeyword(record, f.keyword) {
		return false
	}
	if !query.UpdatedFrom.IsZero() && record.UpdatedAt.Before(query.UpdatedFrom) {
		return false
	}
	if !query.UpdatedTo.IsZero() && record.UpdatedAt.After(query.UpdatedTo) {
		return false
	}
	return true
}

func containsKeyword(record asynctask.TaskEvent, keyword string) bool {
	return strings.Contains(strings.ToLower(record.TaskID), keyword) ||
		strings.Contains(strings.ToLower(record.Queue), keyword) ||
		strings.Contains(strings.ToLower(record.TaskType), keyword) ||
		strings.Contains(strings.ToLower(record.Key), keyword) ||
		strings.Contains(strings.ToLower(record.Message), keyword) ||
		strings.Contains(strings.ToLower(record.Error), keyword)
}

func normalizePagination(offset, limit int64, total int) (int, int) {
	if offset < 0 {
		offset = 0
	}
	total64 := int64(total)
	if offset > total64 {
		offset = total64
	}
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return int(offset), int(limit)
}
