package infrastructure

import (
	"fmt"
	"math"
	"sync"
	"time"

	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
)

// CameraQueue はカメラごとのキュー状態を管理します。
type CameraQueue struct {
	CameraID        string
	PTZQueue        []*protov1.Task
	CinematicQueue  []*protov1.Task
	ExecutingTask   *protov1.Task
	LastPollingAtMs int64
	Interrupt       bool
}

// PTZRepo はPTZサービスのキュー管理を行うリポジトリです。
type PTZRepo struct {
	mu           sync.RWMutex
	cameraQueues map[string]*CameraQueue
}

// NewPTZRepo は新しいPTZRepoを作成します。
func NewPTZRepo() *PTZRepo {
	return &PTZRepo{
		mu:           sync.RWMutex{},
		cameraQueues: make(map[string]*CameraQueue),
	}
}

// EnqueuePTZCommand はPTZ命令をキューに追加します。
// PTZ命令（Layer 1）が届いた場合、シネマティック枠のキューを即座に全削除し、
// 実行中のシネマティック動作を中断させます。
func (r *PTZRepo) EnqueuePTZCommand(
	cameraID string,
	command *protov1.PTZCommand,
) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	queue := r.getOrCreateCameraQueue(cameraID)

	taskID := fmt.Sprintf("ptz-task-%d", time.Now().UnixNano())
	task := &protov1.Task{
		TaskId:           taskID,
		Layer:            protov1.CommandLayer_COMMAND_LAYER_PTZ,
		Status:           protov1.TaskStatus_TASK_STATUS_PENDING,
		PtzCommand:       command,
		CinematicCommand: nil,
		Interrupt:        false,
		CreatedAtMs:      time.Now().UnixMilli(),
	}

	// シネマティック枠を全クリア
	queue.CinematicQueue = make([]*protov1.Task, 0)

	// 実行中のタスクがシネマティックの場合、中断フラグを設定
	if queue.ExecutingTask != nil &&
		queue.ExecutingTask.GetLayer() == protov1.CommandLayer_COMMAND_LAYER_CINEMATIC {
		queue.Interrupt = true
	}

	// PTZキューに追加
	queue.PTZQueue = append(queue.PTZQueue, task)

	return taskID, true
}

// EnqueueCinematicCommand はシネマティック命令をキューに追加します。
// シネマティック枠（Layer 2）はPTZ枠が空の時のみ実行されます。
func (r *PTZRepo) EnqueueCinematicCommand(
	cameraID string,
	command *protov1.CinematographyInstruction,
) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	queue := r.getOrCreateCameraQueue(cameraID)

	taskID := fmt.Sprintf("cine-task-%d", time.Now().UnixNano())
	task := &protov1.Task{
		TaskId:           taskID,
		Layer:            protov1.CommandLayer_COMMAND_LAYER_CINEMATIC,
		Status:           protov1.TaskStatus_TASK_STATUS_PENDING,
		PtzCommand:       nil,
		CinematicCommand: command,
		Interrupt:        false,
		CreatedAtMs:      time.Now().UnixMilli(),
	}

	// シネマティックキューに追加
	queue.CinematicQueue = append(queue.CinematicQueue, task)

	return taskID, true
}

// ProcessPolling はFDからのポーリングを処理し、次の命令を返します。
// completedTaskIdが設定されている場合、該当タスクをデキューします。
func (r *PTZRepo) ProcessPolling(
	cameraID string,
	completedTaskID string,
	executingTaskID string,
	currentPTZ *protov1.PTZParameters,
	deviceStatus protov1.DeviceStatus,
	cameraStatus protov1.CameraStatus,
) (*protov1.Task, *protov1.Task, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	queue := r.getOrCreateCameraQueue(cameraID)
	queue.LastPollingAtMs = time.Now().UnixMilli()

	// 完了タスクの処理
	if completedTaskID != "" {
		r.dequeueCompletedPTZTask(queue, completedTaskID)
		r.dequeueCompletedCinematicTask(queue, completedTaskID)
		r.clearExecutingTaskIfCompleted(queue, completedTaskID)
	}

	// 中断フラグを取得してリセット
	interrupt := queue.Interrupt
	queue.Interrupt = false

	// 次のタスクを取得
	currentCommand, nextCommand := r.getNextTasks(queue)

	// 現在の実行タスクを更新
	if currentCommand != nil {
		queue.ExecutingTask = currentCommand
		currentCommand.Status = protov1.TaskStatus_TASK_STATUS_EXECUTING
	}

	return currentCommand, nextCommand, interrupt
}

// GetQueueStatus はカメラのキュー状態を取得します。
func (r *PTZRepo) GetQueueStatus(cameraID string) *protov1.CameraQueueStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	queue, ok := r.cameraQueues[cameraID]
	if !ok {
		return &protov1.CameraQueueStatus{
			CameraId:           cameraID,
			PtzQueueSize:       0,
			CinematicQueueSize: 0,
			ExecutingTask:      nil,
			LastPollingAtMs:    0,
		}
	}

	return &protov1.CameraQueueStatus{
		CameraId:           cameraID,
		PtzQueueSize:       safeIntToUint32(len(queue.PTZQueue)),
		CinematicQueueSize: safeIntToUint32(len(queue.CinematicQueue)),
		ExecutingTask:      queue.ExecutingTask,
		LastPollingAtMs:    queue.LastPollingAtMs,
	}
}

// GetAllQueueStatuses は全カメラのキュー状態を取得します。
func (r *PTZRepo) GetAllQueueStatuses() []*protov1.CameraQueueStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	statuses := make([]*protov1.CameraQueueStatus, 0, len(r.cameraQueues))

	for _, queue := range r.cameraQueues {
		statuses = append(statuses, &protov1.CameraQueueStatus{
			CameraId:           queue.CameraID,
			PtzQueueSize:       safeIntToUint32(len(queue.PTZQueue)),
			CinematicQueueSize: safeIntToUint32(len(queue.CinematicQueue)),
			ExecutingTask:      queue.ExecutingTask,
			LastPollingAtMs:    queue.LastPollingAtMs,
		})
	}

	return statuses
}

// getOrCreateCameraQueue はカメラキューを取得または作成します。
func (r *PTZRepo) getOrCreateCameraQueue(cameraID string) *CameraQueue {
	if queue, ok := r.cameraQueues[cameraID]; ok {
		return queue
	}

	queue := &CameraQueue{
		CameraID:        cameraID,
		PTZQueue:        make([]*protov1.Task, 0),
		CinematicQueue:  make([]*protov1.Task, 0),
		ExecutingTask:   nil,
		LastPollingAtMs: 0,
		Interrupt:       false,
	}
	r.cameraQueues[cameraID] = queue

	return queue
}

// dequeueCompletedPTZTask はPTZキューから完了したタスクを削除します。
func (r *PTZRepo) dequeueCompletedPTZTask(queue *CameraQueue, taskID string) {
	for i, task := range queue.PTZQueue {
		if task.GetTaskId() == taskID {
			task.Status = protov1.TaskStatus_TASK_STATUS_COMPLETED

			queue.PTZQueue = append(queue.PTZQueue[:i], queue.PTZQueue[i+1:]...)

			return
		}
	}
}

// dequeueCompletedCinematicTask はシネマティックキューから完了したタスクを削除します。
func (r *PTZRepo) dequeueCompletedCinematicTask(queue *CameraQueue, taskID string) {
	for i, task := range queue.CinematicQueue {
		if task.GetTaskId() == taskID {
			task.Status = protov1.TaskStatus_TASK_STATUS_COMPLETED

			queue.CinematicQueue = append(queue.CinematicQueue[:i], queue.CinematicQueue[i+1:]...)

			return
		}
	}
}

// clearExecutingTaskIfCompleted は完了したタスクが実行中の場合、クリアします。
func (r *PTZRepo) clearExecutingTaskIfCompleted(queue *CameraQueue, taskID string) {
	if queue.ExecutingTask != nil && queue.ExecutingTask.GetTaskId() == taskID {
		queue.ExecutingTask = nil
	}
}

// getNextTasks は次に実行すべきタスク（最大2件）を取得します。
// PTZ枠が優先され、PTZ枠が空の場合のみシネマティック枠を実行します。
func (r *PTZRepo) getNextTasks(queue *CameraQueue) (*protov1.Task, *protov1.Task) {
	var tasks []*protov1.Task

	// PTZキューを優先
	if len(queue.PTZQueue) > 0 {
		tasks = queue.PTZQueue
	} else if len(queue.CinematicQueue) > 0 {
		tasks = queue.CinematicQueue
	}

	if len(tasks) == 0 {
		return nil, nil
	}

	var currentCommand *protov1.Task

	var nextCommand *protov1.Task

	currentCommand = tasks[0]

	if len(tasks) > 1 {
		nextCommand = tasks[1]
	}

	return currentCommand, nextCommand
}

// safeIntToUint32 はintをuint32に安全に変換します。
func safeIntToUint32(num int) uint32 {
	if num < 0 {
		return 0
	}

	if num > math.MaxUint32 {
		return math.MaxUint32
	}

	return uint32(num)
}
