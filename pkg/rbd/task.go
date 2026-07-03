package rbd

import (
	"context"
	"fmt"

	cephrados "github.com/ceph/go-ceph/rados"
	cephadmin "github.com/ceph/go-ceph/rbd/admin"
)

// TaskRefs identifies the RBD image associated with a background task.
type TaskRefs struct {
	Action        string `json:"action"`
	PoolName      string `json:"pool_name"`
	PoolNamespace string `json:"pool_namespace"`
	ImageName     string `json:"image_name"`
	ImageID       string `json:"image_id"`
}

// TaskResponse is the JSON response for ceph "rbd task" operations.
type TaskResponse struct {
	Sequence      int      `json:"sequence"`
	ID            string   `json:"id"`
	Message       string   `json:"message"`
	Refs          TaskRefs `json:"refs"`
	InProgress    bool     `json:"in_progress"`
	Progress      float64  `json:"progress"`
	RetryAttempts int      `json:"retry_attempts"`
	RetryTime     string   `json:"retry_time"`
	RetryMessage  string   `json:"retry_message"`
}

func taskResponseFromAdmin(tr cephadmin.TaskResponse) TaskResponse {
	return TaskResponse{
		Sequence:      tr.Sequence,
		ID:            tr.ID,
		Message:       tr.Message,
		Refs:          TaskRefs(tr.Refs),
		InProgress:    tr.InProgress,
		Progress:      tr.Progress,
		RetryAttempts: tr.RetryAttempts,
		RetryTime:     tr.RetryTime,
		RetryMessage:  tr.RetryMessage,
	}
}

func taskResponsesFromAdmin(trs []cephadmin.TaskResponse) []TaskResponse {
	out := make([]TaskResponse, len(trs))
	for i, tr := range trs {
		out[i] = taskResponseFromAdmin(tr)
	}
	return out
}

func adminImageSpec(spec ImageSpec) cephadmin.ImageSpec {
	return cephadmin.NewImageSpec(spec.Pool(), spec.Namespace(), spec.Image())
}

func taskAdmin(conn *cephrados.Conn) *cephadmin.TaskAdmin {
	return cephadmin.NewFromConn(conn).Task()
}

func validateTaskImageSpec(imageSpec ImageSpec) error {
	if !imageSpec.Valid() {
		return fmt.Errorf("invalid image spec: %s", imageSpec)
	}
	return nil
}

// RbdTaskAddFlatten submits a background flatten task for a cloned image.
// Equivalent to: rbd task add flatten <image-spec>
//
// The caller returns immediately; flatten runs in the Ceph manager.
// Requires a connected RADOS client with permission to run mgr commands.
func RbdTaskAddFlatten(ctx context.Context, conn *cephrados.Conn, imageSpec ImageSpec) (TaskResponse, error) {
	if err := validateTaskImageSpec(imageSpec); err != nil {
		return TaskResponse{}, err
	}

	tr, err := taskAdmin(conn).AddFlatten(adminImageSpec(imageSpec))
	if err != nil {
		return TaskResponse{}, fmt.Errorf("failed to add flatten task for image (%s): %w", imageSpec, err)
	}
	return taskResponseFromAdmin(tr), nil
}

// RbdTaskAddRemove submits a background remove task for an image.
// Equivalent to: rbd task add remove <image-spec>
func RbdTaskAddRemove(ctx context.Context, conn *cephrados.Conn, imageSpec ImageSpec) (TaskResponse, error) {
	if err := validateTaskImageSpec(imageSpec); err != nil {
		return TaskResponse{}, err
	}

	tr, err := taskAdmin(conn).AddRemove(adminImageSpec(imageSpec))
	if err != nil {
		return TaskResponse{}, fmt.Errorf("failed to add remove task for image (%s): %w", imageSpec, err)
	}
	return taskResponseFromAdmin(tr), nil
}

// RbdTaskAddTrashRemove submits a background trash-remove task.
// Equivalent to: rbd task add trash remove <image-id-spec>
func RbdTaskAddTrashRemove(ctx context.Context, conn *cephrados.Conn, imageIDSpec ImageSpec) (TaskResponse, error) {
	if imageIDSpec.clean() == "" {
		return TaskResponse{}, fmt.Errorf("invalid image id spec: %s", imageIDSpec)
	}

	tr, err := taskAdmin(conn).AddTrashRemove(cephadmin.NewRawImageSpec(string(imageIDSpec)))
	if err != nil {
		return TaskResponse{}, fmt.Errorf("failed to add trash remove task for image (%s): %w", imageIDSpec, err)
	}
	return taskResponseFromAdmin(tr), nil
}

// RbdTaskList lists pending or running asynchronous RBD tasks.
// Equivalent to: rbd task list
func RbdTaskList(ctx context.Context, conn *cephrados.Conn) ([]TaskResponse, error) {
	trs, err := taskAdmin(conn).List()
	if err != nil {
		return nil, fmt.Errorf("failed to list rbd tasks: %w", err)
	}
	return taskResponsesFromAdmin(trs), nil
}

// RbdTaskGet returns a task by ID.
// Equivalent to: rbd task list <task-id>
func RbdTaskGet(ctx context.Context, conn *cephrados.Conn, taskID string) (TaskResponse, error) {
	if taskID == "" {
		return TaskResponse{}, fmt.Errorf("task id is required")
	}

	tr, err := taskAdmin(conn).GetTaskByID(taskID)
	if err != nil {
		return TaskResponse{}, fmt.Errorf("failed to get rbd task (%s): %w", taskID, err)
	}
	return taskResponseFromAdmin(tr), nil
}

// RbdTaskCancel cancels a pending or running asynchronous RBD task.
// Equivalent to: rbd task cancel <task-id>
func RbdTaskCancel(ctx context.Context, conn *cephrados.Conn, taskID string) (TaskResponse, error) {
	if taskID == "" {
		return TaskResponse{}, fmt.Errorf("task id is required")
	}

	tr, err := taskAdmin(conn).Cancel(taskID)
	if err != nil {
		return TaskResponse{}, fmt.Errorf("failed to cancel rbd task (%s): %w", taskID, err)
	}
	return taskResponseFromAdmin(tr), nil
}
