package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/failup-ventures/attio-cli/internal/api"
	"github.com/failup-ventures/attio-cli/internal/outfmt"
)

type MeetingsCmd struct {
	List       MeetingsListCmd       `cmd:"" help:"List meetings (Beta)"`
	Get        MeetingsGetCmd        `cmd:"" help:"Get meeting (Beta)"`
	Create     MeetingsCreateCmd     `cmd:"" help:"Find or create meeting (Alpha)"`
	Recordings MeetingsRecordingsCmd `cmd:"" help:"Manage call recordings (Alpha/Beta)"`
	Transcript MeetingsTranscriptCmd `cmd:"" help:"Get transcript segments (Beta)"`
}

type MeetingsListCmd struct {
	Limit          int    `name:"limit" help:"Page size" default:"50"`
	Cursor         string `name:"cursor" help:"Cursor"`
	Sort           string `name:"sort" help:"Sort order"`
	Participants   string `name:"participants" help:"Participants filter"`
	LinkedObject   string `name:"linked-object" help:"Linked object slug or UUID"`
	LinkedRecordID string `name:"linked-record-id" help:"Linked record UUID"`
	EndsFrom       string `name:"ends-from" help:"ISO timestamp lower bound for end"`
	StartsBefore   string `name:"starts-before" help:"ISO timestamp upper bound for start"`
	Timezone       string `name:"timezone" help:"IANA timezone"`
	All            bool   `name:"all" help:"Fetch all pages"`
	MaxPages       int    `name:"max-pages" help:"Maximum pages when --all is set" default:"100"`
}

func (c *MeetingsListCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}

	if c.All {
		startCursor := c.Cursor
		meetings, err := api.FetchAllCursor(ctx, c.MaxPages, func(cursor string) ([]map[string]any, string, error) {
			if cursor == "" {
				cursor = startCursor
			}
			return client.ListMeetings(ctx, c.Limit, cursor, c.Sort, c.Participants, c.LinkedObject, c.LinkedRecordID, c.EndsFrom, c.StartsBefore, c.Timezone)
		})
		if err != nil {
			return err
		}
		return writeMeetings(ctx, meetings)
	}

	meetings, nextCursor, err := client.ListMeetings(ctx, c.Limit, c.Cursor, c.Sort, c.Participants, c.LinkedObject, c.LinkedRecordID, c.EndsFrom, c.StartsBefore, c.Timezone)
	if err != nil {
		return err
	}
	if err := maybePreviewResults(ctx, len(meetings)); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{
			"data": meetings,
			"pagination": map[string]any{
				"next_cursor": nextCursor,
			},
		})
	}
	return writeMeetings(ctx, meetings)
}

type MeetingsGetCmd struct {
	MeetingID string `arg:"" name:"meeting-id" help:"Meeting UUID" required:""`
}

func (c *MeetingsGetCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	meeting, err := client.GetMeeting(ctx, c.MeetingID)
	if err != nil {
		return err
	}
	return writeSingleMeeting(ctx, meeting)
}

type MeetingsCreateCmd struct {
	Data string `name:"data" help:"Meeting payload JSON; supports '-' or @file.json" required:""`
}

func (c *MeetingsCreateCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := readJSONObjectInput(c.Data)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "meetings create", map[string]any{"data": data}); ok || err != nil {
		return err
	}
	meeting, err := client.FindOrCreateMeeting(ctx, data)
	if err != nil {
		return err
	}
	return writeSingleMeeting(ctx, meeting)
}

type MeetingsRecordingsCmd struct {
	List   MeetingsRecordingsListCmd   `cmd:"" help:"List call recordings (Beta)"`
	Get    MeetingsRecordingsGetCmd    `cmd:"" help:"Get call recording (Beta)"`
	Create MeetingsRecordingsCreateCmd `cmd:"" help:"Create call recording (Alpha)"`
	Delete MeetingsRecordingsDeleteCmd `cmd:"" help:"Delete call recording (Alpha)"`
}

type MeetingsRecordingsListCmd struct {
	MeetingID string `arg:"" name:"meeting-id" help:"Meeting UUID" required:""`
	Limit     int    `name:"limit" help:"Page size" default:"50"`
	Cursor    string `name:"cursor" help:"Cursor"`
	All       bool   `name:"all" help:"Fetch all pages"`
	MaxPages  int    `name:"max-pages" help:"Maximum pages when --all is set" default:"100"`
}

func (c *MeetingsRecordingsListCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}

	if c.All {
		startCursor := c.Cursor
		recordings, err := api.FetchAllCursor(ctx, c.MaxPages, func(cursor string) ([]map[string]any, string, error) {
			if cursor == "" {
				cursor = startCursor
			}
			return client.ListCallRecordings(ctx, c.MeetingID, c.Limit, cursor)
		})
		if err != nil {
			return err
		}
		return writeCallRecordings(ctx, recordings)
	}

	recordings, nextCursor, err := client.ListCallRecordings(ctx, c.MeetingID, c.Limit, c.Cursor)
	if err != nil {
		return err
	}
	if err := maybePreviewResults(ctx, len(recordings)); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{
			"data": recordings,
			"pagination": map[string]any{
				"next_cursor": nextCursor,
			},
		})
	}
	return writeCallRecordings(ctx, recordings)
}

type MeetingsRecordingsGetCmd struct {
	MeetingID       string `arg:"" name:"meeting-id" help:"Meeting UUID" required:""`
	CallRecordingID string `arg:"" name:"recording-id" help:"Call recording UUID" required:""`
}

func (c *MeetingsRecordingsGetCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	recording, err := client.GetCallRecording(ctx, c.MeetingID, c.CallRecordingID)
	if err != nil {
		return err
	}
	return writeSingleCallRecording(ctx, recording)
}

type MeetingsRecordingsCreateCmd struct {
	MeetingID string `arg:"" name:"meeting-id" help:"Meeting UUID" required:""`
	Data      string `name:"data" help:"Call recording payload JSON; supports '-' or @file.json" required:""`
}

func (c *MeetingsRecordingsCreateCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := readJSONObjectInput(c.Data)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "meetings recordings create", map[string]any{"meeting_id": c.MeetingID, "data": data}); ok || err != nil {
		return err
	}
	recording, err := client.CreateCallRecording(ctx, c.MeetingID, data)
	if err != nil {
		return err
	}
	return writeSingleCallRecording(ctx, recording)
}

type MeetingsRecordingsDeleteCmd struct {
	MeetingID       string `arg:"" name:"meeting-id" help:"Meeting UUID" required:""`
	CallRecordingID string `arg:"" name:"recording-id" help:"Call recording UUID" required:""`
}

func (c *MeetingsRecordingsDeleteCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "meetings recordings delete", map[string]any{"meeting_id": c.MeetingID, "recording_id": c.CallRecordingID}); ok || err != nil {
		return err
	}
	if err := client.DeleteCallRecording(ctx, c.MeetingID, c.CallRecordingID); err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"deleted": true, "recording_id": c.CallRecordingID})
	}
	_, _ = os.Stdout.WriteString("Deleted call recording " + c.CallRecordingID + "\n")
	return nil
}

type MeetingsTranscriptCmd struct {
	MeetingID       string `arg:"" name:"meeting-id" help:"Meeting UUID" required:""`
	CallRecordingID string `arg:"" name:"recording-id" help:"Call recording UUID" required:""`
	Cursor          string `name:"cursor" help:"Cursor"`
	All             bool   `name:"all" help:"Fetch all pages"`
	MaxPages        int    `name:"max-pages" help:"Maximum pages when --all is set" default:"100"`
}

func (c *MeetingsTranscriptCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}

	if c.All {
		startCursor := c.Cursor
		segments, err := api.FetchAllCursor(ctx, c.MaxPages, func(cursor string) ([]map[string]any, string, error) {
			if cursor == "" {
				cursor = startCursor
			}
			return client.GetTranscript(ctx, c.MeetingID, c.CallRecordingID, cursor)
		})
		if err != nil {
			return err
		}
		return writeTranscriptSegments(ctx, segments)
	}

	segments, nextCursor, err := client.GetTranscript(ctx, c.MeetingID, c.CallRecordingID, c.Cursor)
	if err != nil {
		return err
	}
	if err := maybePreviewResults(ctx, len(segments)); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{
			"data": segments,
			"pagination": map[string]any{
				"next_cursor": nextCursor,
			},
		})
	}
	return writeTranscriptSegments(ctx, segments)
}

func writeSingleMeeting(ctx context.Context, meeting map[string]any) error {
	if ok, err := maybeWriteIDOnly(ctx, meeting); ok || err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": meeting})
	}
	return writeMeetings(ctx, []map[string]any{meeting})
}

func writeMeetings(ctx context.Context, meetings []map[string]any) error {
	if err := maybePreviewResults(ctx, len(meetings)); err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": meetings})
	}
	w, done := tableWriter(ctx)
	defer done()
	_, _ = fmt.Fprintln(w, "ID\tTITLE\tSTART\tEND\tPARTICIPANTS")
	for _, meeting := range meetings {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			idString(meeting["id"]),
			mapString(meeting, "title"),
			mapString(meeting, "start_at"),
			mapString(meeting, "end_at"),
			meetingParticipantsSummary(meeting),
		)
	}
	return nil
}

func writeSingleCallRecording(ctx context.Context, recording map[string]any) error {
	if ok, err := maybeWriteIDOnly(ctx, recording); ok || err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": recording})
	}
	return writeCallRecordings(ctx, []map[string]any{recording})
}

func writeCallRecordings(ctx context.Context, recordings []map[string]any) error {
	if err := maybePreviewResults(ctx, len(recordings)); err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": recordings})
	}
	w, done := tableWriter(ctx)
	defer done()
	_, _ = fmt.Fprintln(w, "ID\tCREATED_AT\tURL")
	for _, recording := range recordings {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n",
			idString(recording["id"]),
			mapString(recording, "created_at"),
			mapString(recording, "url"),
		)
	}
	return nil
}

func writeTranscriptSegments(ctx context.Context, segments []map[string]any) error {
	if err := maybePreviewResults(ctx, len(segments)); err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": segments})
	}
	w, done := tableWriter(ctx)
	defer done()
	_, _ = fmt.Fprintln(w, "START\tEND\tSPEAKER\tTEXT")
	for _, seg := range segments {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			mapString(seg, "start_at"),
			mapString(seg, "end_at"),
			mapString(seg, "speaker_name"),
			mapString(seg, "text"),
		)
	}
	return nil
}
