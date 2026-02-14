package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/failup-ventures/attio-cli/internal/outfmt"
)

type TasksCmd struct {
	List   TasksListCmd   `cmd:"" help:"List tasks"`
	Create TasksCreateCmd `cmd:"" help:"Create task"`
	Get    TasksGetCmd    `cmd:"" help:"Get task"`
	Update TasksUpdateCmd `cmd:"" help:"Update task"`
	Delete TasksDeleteCmd `cmd:"" help:"Delete task"`
}

type TasksListCmd struct {
	Limit        int    `name:"limit" help:"Page size" default:"20"`
	Offset       int    `name:"offset" help:"Offset" default:"0"`
	Sort         string `name:"sort" help:"Sort key"`
	LinkedObject string `name:"linked-object" help:"Linked object slug or UUID"`
	LinkedRecord string `name:"linked-record" help:"Linked record UUID"`
	Assignee     string `name:"assignee" help:"Assignee workspace member UUID"`
	IsCompleted  string `name:"is-completed" help:"Filter completed tasks (true|false)"`
}

func (c *TasksListCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}

	isCompleted, err := parseOptionalBoolFlag(c.IsCompleted, "--is-completed")
	if err != nil {
		return err
	}

	tasks, err := client.ListTasks(ctx, c.Limit, c.Offset, c.Sort, c.LinkedObject, c.LinkedRecord, c.Assignee, isCompleted)
	if err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		if err := maybePreviewResults(ctx, len(tasks)); err != nil {
			return err
		}
		return writeOffsetPaginatedJSON(ctx, tasks, c.Limit, c.Offset)
	}
	return writeTasks(ctx, tasks)
}

type TasksCreateCmd struct {
	Data          string `name:"data" help:"Optional task payload JSON; supports '-' or @file.json"`
	Content       string `name:"content" help:"Task content (required)"`
	DeadlineAt    string `name:"deadline" help:"Task deadline timestamp (ISO 8601) or 'null'"`
	Assignees     string `name:"assignees" help:"Comma-separated assignee IDs or emails"`
	LinkedRecords string `name:"linked-records" help:"Linked records JSON array"`
	IsCompleted   string `name:"is-completed" help:"Task completion state (true|false)"`
}

func (c *TasksCreateCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := c.payload()
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "tasks create", map[string]any{"data": data}); ok || err != nil {
		return err
	}
	task, err := client.CreateTask(ctx, data)
	if err != nil {
		return err
	}
	return writeSingleTask(ctx, task)
}

type TasksGetCmd struct {
	TaskID string `arg:"" name:"task-id" help:"Task UUID" required:""`
}

func (c *TasksGetCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	task, err := client.GetTask(ctx, c.TaskID)
	if err != nil {
		return err
	}
	return writeSingleTask(ctx, task)
}

type TasksUpdateCmd struct {
	TaskID string `arg:"" name:"task-id" help:"Task UUID" required:""`
	Data   string `name:"data" help:"Optional task payload JSON; supports '-' or @file.json"`

	DeadlineAt    string `name:"deadline" help:"Task deadline timestamp (ISO 8601) or 'null'"`
	Assignees     string `name:"assignees" help:"Comma-separated assignee IDs or emails"`
	LinkedRecords string `name:"linked-records" help:"Linked records JSON array"`
	IsCompleted   string `name:"is-completed" help:"Task completion state (true|false)"`
}

func (c *TasksUpdateCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := c.payload()
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "tasks update", map[string]any{"task_id": c.TaskID, "data": data}); ok || err != nil {
		return err
	}
	task, err := client.UpdateTask(ctx, c.TaskID, data)
	if err != nil {
		return err
	}
	return writeSingleTask(ctx, task)
}

type TasksDeleteCmd struct {
	TaskID string `arg:"" name:"task-id" help:"Task UUID" required:""`
}

func (c *TasksDeleteCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "tasks delete", map[string]any{"task_id": c.TaskID}); ok || err != nil {
		return err
	}
	if err := client.DeleteTask(ctx, c.TaskID); err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"deleted": true, "task_id": c.TaskID})
	}
	_, _ = os.Stdout.WriteString("Deleted task " + c.TaskID + "\n")
	return nil
}

func writeSingleTask(ctx context.Context, task map[string]any) error {
	if ok, err := maybeWriteIDOnly(ctx, task); ok || err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": task})
	}
	return writeTasks(ctx, []map[string]any{task})
}

func writeTasks(ctx context.Context, tasks []map[string]any) error {
	if err := maybePreviewResults(ctx, len(tasks)); err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": tasks})
	}
	w, done := tableWriter(ctx)
	defer done()
	_, _ = fmt.Fprintln(w, "ID\tCONTENT\tSTATUS\tDEADLINE\tASSIGNEE")
	for _, task := range tasks {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			idString(task["id"]),
			mapString(task, "content"),
			taskStatusSummary(task),
			mapString(task, "deadline_at"),
			taskAssigneeSummary(task),
		)
	}
	return nil
}

func (c *TasksCreateCmd) payload() (map[string]any, error) {
	data := map[string]any{}
	if strings.TrimSpace(c.Data) != "" {
		parsed, err := readJSONObjectInput(c.Data)
		if err != nil {
			return nil, err
		}
		data = parsed
	}

	if c.Content != "" {
		data["content"] = c.Content
	}
	if c.DeadlineAt != "" {
		if strings.EqualFold(c.DeadlineAt, "null") {
			data["deadline_at"] = nil
		} else {
			data["deadline_at"] = c.DeadlineAt
		}
	}
	if c.IsCompleted != "" {
		val, err := parseOptionalBoolFlag(c.IsCompleted, "--is-completed")
		if err != nil {
			return nil, err
		}
		if val != nil {
			data["is_completed"] = *val
		}
	}
	if c.Assignees != "" {
		assignees, err := parseTaskAssigneesFlag(c.Assignees)
		if err != nil {
			return nil, err
		}
		data["assignees"] = assignees
	}
	if c.LinkedRecords != "" {
		records, err := parseJSONArrayFlag(c.LinkedRecords, "--linked-records")
		if err != nil {
			return nil, err
		}
		data["linked_records"] = records
	}

	if !hasMapKey(data, "format") {
		data["format"] = "plaintext"
	}
	if !hasMapKey(data, "content") {
		return nil, newUsageError(errors.New("tasks create requires --content (or content in --data)"))
	}
	if !hasMapKey(data, "deadline_at") {
		data["deadline_at"] = nil
	}
	if !hasMapKey(data, "is_completed") {
		data["is_completed"] = false
	}
	if !hasMapKey(data, "linked_records") {
		data["linked_records"] = []any{}
	}
	if !hasMapKey(data, "assignees") {
		data["assignees"] = []any{}
	}

	return data, nil
}

func (c *TasksUpdateCmd) payload() (map[string]any, error) {
	data := map[string]any{}
	if strings.TrimSpace(c.Data) != "" {
		parsed, err := readJSONObjectInput(c.Data)
		if err != nil {
			return nil, err
		}
		data = parsed
	}

	if c.DeadlineAt != "" {
		if strings.EqualFold(c.DeadlineAt, "null") {
			data["deadline_at"] = nil
		} else {
			data["deadline_at"] = c.DeadlineAt
		}
	}
	if c.IsCompleted != "" {
		val, err := parseOptionalBoolFlag(c.IsCompleted, "--is-completed")
		if err != nil {
			return nil, err
		}
		if val != nil {
			data["is_completed"] = *val
		}
	}
	if c.Assignees != "" {
		assignees, err := parseTaskAssigneesFlag(c.Assignees)
		if err != nil {
			return nil, err
		}
		data["assignees"] = assignees
	}
	if c.LinkedRecords != "" {
		records, err := parseJSONArrayFlag(c.LinkedRecords, "--linked-records")
		if err != nil {
			return nil, err
		}
		data["linked_records"] = records
	}

	if len(data) == 0 {
		return nil, newUsageError(errors.New("tasks update requires at least one update field"))
	}
	if hasMapKey(data, "content") {
		return nil, newUsageError(errors.New("tasks update does not support content updates"))
	}

	return data, nil
}
