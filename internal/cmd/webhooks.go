package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/failup-ventures/attio-cli/internal/outfmt"
)

type WebhooksCmd struct {
	List   WebhooksListCmd   `cmd:"" help:"List webhooks"`
	Create WebhooksCreateCmd `cmd:"" help:"Create webhook"`
	Get    WebhooksGetCmd    `cmd:"" help:"Get webhook"`
	Update WebhooksUpdateCmd `cmd:"" help:"Update webhook"`
	Delete WebhooksDeleteCmd `cmd:"" help:"Delete webhook"`
}

type WebhooksListCmd struct {
	Limit  int `name:"limit" help:"Page size" default:"20"`
	Offset int `name:"offset" help:"Offset" default:"0"`
}

func (c *WebhooksListCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	webhooks, err := client.ListWebhooks(ctx, c.Limit, c.Offset)
	if err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		if err := maybePreviewResults(ctx, len(webhooks)); err != nil {
			return err
		}
		return writeOffsetPaginatedJSON(ctx, webhooks, c.Limit, c.Offset)
	}
	return writeWebhooks(ctx, webhooks)
}

type WebhooksCreateCmd struct {
	Data          string `name:"data" help:"Optional webhook payload JSON; supports '-' or @file.json"`
	TargetURL     string `name:"target-url" help:"Webhook destination URL (https://...)"`
	Subscriptions string `name:"subscriptions" help:"Subscriptions JSON array"`
}

func (c *WebhooksCreateCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := c.payload()
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "webhooks create", map[string]any{"data": data}); ok || err != nil {
		return err
	}
	webhook, err := client.CreateWebhook(ctx, data)
	if err != nil {
		return err
	}
	return writeSingleWebhook(ctx, webhook)
}

type WebhooksGetCmd struct {
	WebhookID string `arg:"" name:"webhook-id" help:"Webhook UUID" required:""`
}

func (c *WebhooksGetCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	webhook, err := client.GetWebhook(ctx, c.WebhookID)
	if err != nil {
		return err
	}
	return writeSingleWebhook(ctx, webhook)
}

type WebhooksUpdateCmd struct {
	WebhookID string `arg:"" name:"webhook-id" help:"Webhook UUID" required:""`
	Data      string `name:"data" help:"Optional webhook payload JSON; supports '-' or @file.json"`

	TargetURL     string `name:"target-url" help:"Webhook destination URL (https://...)"`
	Subscriptions string `name:"subscriptions" help:"Subscriptions JSON array"`
}

func (c *WebhooksUpdateCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := c.payload()
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "webhooks update", map[string]any{"webhook_id": c.WebhookID, "data": data}); ok || err != nil {
		return err
	}
	webhook, err := client.UpdateWebhook(ctx, c.WebhookID, data)
	if err != nil {
		return err
	}
	return writeSingleWebhook(ctx, webhook)
}

type WebhooksDeleteCmd struct {
	WebhookID string `arg:"" name:"webhook-id" help:"Webhook UUID" required:""`
}

func (c *WebhooksDeleteCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "webhooks delete", map[string]any{"webhook_id": c.WebhookID}); ok || err != nil {
		return err
	}
	if err := client.DeleteWebhook(ctx, c.WebhookID); err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"deleted": true, "webhook_id": c.WebhookID})
	}
	_, _ = os.Stdout.WriteString("Deleted webhook " + c.WebhookID + "\n")
	return nil
}

func writeSingleWebhook(ctx context.Context, webhook map[string]any) error {
	if ok, err := maybeWriteIDOnly(ctx, webhook); ok || err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": webhook})
	}
	return writeWebhooks(ctx, []map[string]any{webhook})
}

func writeWebhooks(ctx context.Context, webhooks []map[string]any) error {
	if err := maybePreviewResults(ctx, len(webhooks)); err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": webhooks})
	}
	w, done := tableWriter(ctx)
	defer done()
	_, _ = fmt.Fprintln(w, "ID\tTARGET_URL\tSUBSCRIPTIONS\tSTATUS\tCREATED_AT")
	for _, webhook := range webhooks {
		subs, _ := webhook["subscriptions"].([]any)
		_, _ = fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
			idString(webhook["id"]),
			mapString(webhook, "target_url"),
			len(subs),
			mapString(webhook, "status"),
			mapString(webhook, "created_at"),
		)
	}
	return nil
}

func (c *WebhooksCreateCmd) payload() (map[string]any, error) {
	data := map[string]any{}
	if strings.TrimSpace(c.Data) != "" {
		parsed, err := readJSONObjectInput(c.Data)
		if err != nil {
			return nil, err
		}
		data = parsed
	}

	if c.TargetURL != "" {
		data["target_url"] = c.TargetURL
	}
	if c.Subscriptions != "" {
		items, err := parseJSONArrayFlag(c.Subscriptions, "--subscriptions")
		if err != nil {
			return nil, err
		}
		data["subscriptions"] = items
	}

	for _, key := range []string{"target_url", "subscriptions"} {
		if !hasMapKey(data, key) {
			return nil, newUsageError(errors.New("webhooks create requires --target-url and --subscriptions (or equivalent --data payload)"))
		}
	}
	return data, nil
}

func (c *WebhooksUpdateCmd) payload() (map[string]any, error) {
	data := map[string]any{}
	if strings.TrimSpace(c.Data) != "" {
		parsed, err := readJSONObjectInput(c.Data)
		if err != nil {
			return nil, err
		}
		data = parsed
	}

	if c.TargetURL != "" {
		data["target_url"] = c.TargetURL
	}
	if c.Subscriptions != "" {
		items, err := parseJSONArrayFlag(c.Subscriptions, "--subscriptions")
		if err != nil {
			return nil, err
		}
		data["subscriptions"] = items
	}
	if len(data) == 0 {
		return nil, newUsageError(errors.New("webhooks update requires at least one update field"))
	}
	return data, nil
}
