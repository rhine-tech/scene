package v1

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
	events sync.Map
	count  atomic.Int64

	nextCleanupAt  atomic.Int64
	cleanupRunning atomic.Bool
}

type eventEntry struct {
	event   atomic.Pointer[asynctask.TaskEvent]
	deleted atomic.Bool
}

type eventSnapshot struct {
	taskID    string
	entry     *eventEntry
	updatedAt time.Time
}

func NewMemoryReporter(config ...Config) *MemoryReporter {
	cfg := Config{}
	if len(config) > 0 {
		cfg = config[0]
	}
	return &MemoryReporter{config: normalizeConfig(cfg)}
}

func (m *MemoryReporter) ImplName() scene.ImplName {
	return asynctask.Lens.ImplName("TaskReporter", "memory.v1")
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

	entry, inserted := m.eventEntry(taskID)
	for {
		if entry.deleted.Load() {
			entry, inserted = m.eventEntry(taskID)
			continue
		}
		current := entry.event.Load()
		if current != nil && event.UpdatedAt.Before(current.UpdatedAt) {
			m.cleanupIfNeeded(now)
			return nil
		}
		merged := mergeEvent(current, event)
		if !entry.event.CompareAndSwap(current, &merged) {
			continue
		}
		if entry.deleted.Load() {
			entry, inserted = m.eventEntry(taskID)
			continue
		}
		if inserted {
			m.cleanupAfterInsert(now, taskID, entry)
		} else {
			m.cleanupIfNeeded(now)
		}
		return nil
	}
}

func (m *MemoryReporter) Cleanup(_ context.Context) int64 {
	return m.cleanup(time.Now())
}

func (m *MemoryReporter) ListTasks(_ context.Context, query asynctask.TaskQuery) (*model.PaginationResult[asynctask.TaskEvent], error) {
	records := make([]asynctask.TaskEvent, 0, m.count.Load())
	filter := newTaskFilter(query)
	m.events.Range(func(key any, value any) bool {
		entry, ok := value.(*eventEntry)
		if !ok || entry.deleted.Load() {
			return true
		}
		current := entry.event.Load()
		if current == nil {
			return true
		}
		record := *current
		if filter.match(record) {
			records = append(records, record)
		}
		return true
	})

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
	raw, ok := m.events.Load(taskID)
	if !ok {
		return asynctask.TaskEvent{}, asynctask.ErrTaskNotFound.WithDetailStr(taskID)
	}
	entry, ok := raw.(*eventEntry)
	if !ok || entry.deleted.Load() {
		return asynctask.TaskEvent{}, asynctask.ErrTaskNotFound.WithDetailStr(taskID)
	}
	current := entry.event.Load()
	if current == nil {
		return asynctask.TaskEvent{}, asynctask.ErrTaskNotFound.WithDetailStr(taskID)
	}
	return *current, nil
}

func (m *MemoryReporter) GetQueueStats(_ context.Context, queue string) (asynctask.QueueStats, error) {
	stats := asynctask.QueueStats{
		Queue:    queue,
		ByStatus: make(map[asynctask.QueueTaskStatus]int64),
	}
	m.events.Range(func(key any, value any) bool {
		entry, ok := value.(*eventEntry)
		if !ok || entry.deleted.Load() {
			return true
		}
		current := entry.event.Load()
		if current == nil {
			return true
		}
		record := *current
		if queue != "" && record.Queue != queue {
			return true
		}
		stats.Total++
		stats.ByStatus[record.Status]++
		return true
	})
	return stats, nil
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
	m.events.Range(func(key any, value any) bool {
		taskID, ok := key.(string)
		if !ok {
			return true
		}
		entry, ok := value.(*eventEntry)
		if !ok || entry.deleted.Load() {
			return true
		}
		current := entry.event.Load()
		if current != nil && current.UpdatedAt.Before(cutoff) && m.deleteEntry(taskID, entry) {
			removed++
		}
		return true
	})
	return removed
}

func (m *MemoryReporter) cleanupAfterInsert(now time.Time, taskID string, entry *eventEntry) {
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
	if m.count.Load() > maxTasks && entry != nil {
		m.deleteEntry(taskID, entry)
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

	snapshots := make([]eventSnapshot, 0, m.count.Load())
	m.events.Range(func(key any, value any) bool {
		taskID, ok := key.(string)
		if !ok {
			return true
		}
		entry, ok := value.(*eventEntry)
		if !ok || entry.deleted.Load() {
			return true
		}
		current := entry.event.Load()
		if current == nil {
			return true
		}
		snapshots = append(snapshots, eventSnapshot{
			taskID:    taskID,
			entry:     entry,
			updatedAt: current.UpdatedAt,
		})
		return true
	})

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
		if m.deleteEntryIfUpdatedAt(snapshot.taskID, snapshot.entry, snapshot.updatedAt) {
			removed++
		}
	}
	return removed
}

func (m *MemoryReporter) eventEntry(taskID string) (*eventEntry, bool) {
	for {
		if raw, ok := m.events.Load(taskID); ok {
			entry := raw.(*eventEntry)
			if !entry.deleted.Load() {
				return entry, false
			}
			continue
		}
		entry := &eventEntry{}
		actual, loaded := m.events.LoadOrStore(taskID, entry)
		if !loaded {
			m.count.Add(1)
			return entry, true
		}
		actualEntry := actual.(*eventEntry)
		if !actualEntry.deleted.Load() {
			return actualEntry, false
		}
	}
}

func (m *MemoryReporter) deleteEntry(taskID string, entry *eventEntry) bool {
	if !entry.deleted.CompareAndSwap(false, true) {
		return false
	}
	if m.events.CompareAndDelete(taskID, entry) {
		m.count.Add(-1)
		return true
	}
	entry.deleted.Store(false)
	return false
}

func (m *MemoryReporter) deleteEntryIfUpdatedAt(taskID string, entry *eventEntry, updatedAt time.Time) bool {
	current := entry.event.Load()
	if current == nil || !current.UpdatedAt.Equal(updatedAt) {
		return false
	}
	return m.deleteEntry(taskID, entry)
}

func mergeEvent(current *asynctask.TaskEvent, event asynctask.TaskEvent) asynctask.TaskEvent {
	record := asynctask.TaskEvent{TaskID: event.TaskID}
	if current != nil {
		record = *current
	}
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
