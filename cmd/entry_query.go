package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/spf13/cobra"
)

const calendarDateLayout = "2006-01-02"

type entryQueryOptions struct {
	number  int
	all     bool
	date    string
	since   string
	until   string
	reverse bool
	long    bool
	json    bool
}

type logEntryJSON struct {
	CommitID     string   `json:"commit_id"`
	Date         string   `json:"date"`
	Message      string   `json:"message"`
	MessageBody  string   `json:"message_body,omitempty"`
	Files        []string `json:"files"`
	IsCommitted  bool     `json:"is_committed"`
	NotionSynced bool     `json:"notion_synced"`
}

func addEntryQueryFlags(command *cobra.Command, options *entryQueryOptions) {
	flags := command.Flags()
	flags.IntVarP(&options.number, "number", "n", 10, "Number of entries to show")
	flags.BoolVar(&options.all, "all", false, "Show all matching entries")
	flags.StringVar(&options.date, "date", "", "Show entries recorded on a date (YYYY-MM-DD)")
	flags.StringVar(&options.since, "since", "", "Show entries on or after a date (YYYY-MM-DD)")
	flags.StringVar(&options.until, "until", "", "Show entries on or before a date (YYYY-MM-DD)")
	flags.BoolVarP(&options.reverse, "reverse", "r", false, "Show oldest entries first")
	flags.BoolVarP(&options.long, "long", "l", false, "Show entry bodies and full metadata")
	flags.BoolVar(&options.json, "json", false, "Print entries as JSON")
}

func runEntryQuery(command *cobra.Command, search string, options entryQueryOptions) error {
	query, err := buildEntryQuery(command, search, options)
	if err != nil {
		return err
	}

	_, manager, err := loadManager()
	if err != nil {
		return err
	}
	entries, err := manager.QueryEntries(query)
	if err != nil {
		return err
	}

	if options.json {
		return writeEntriesJSON(command.OutOrStdout(), entries)
	}
	if len(entries) == 0 {
		if strings.TrimSpace(search) == "" {
			fmt.Fprintln(command.OutOrStdout(), "No entries found")
		} else {
			fmt.Fprintf(command.OutOrStdout(), "No entries matched %q\n", strings.TrimSpace(search))
		}
		return nil
	}
	if options.long {
		return writeEntriesLong(command.OutOrStdout(), entries)
	}
	return writeEntriesTable(command.OutOrStdout(), entries)
}

func buildEntryQuery(
	command *cobra.Command,
	search string,
	options entryQueryOptions,
) (til.EntryQuery, error) {
	if options.all && command.Flags().Changed("number") {
		return til.EntryQuery{}, errors.New("--all and --number cannot be used together")
	}
	if !options.all && options.number <= 0 {
		return til.EntryQuery{}, errors.New("--number must be greater than zero")
	}
	if options.long && options.json {
		return til.EntryQuery{}, errors.New("--long and --json cannot be used together")
	}
	if options.date != "" && (options.since != "" || options.until != "") {
		return til.EntryQuery{}, errors.New("--date cannot be combined with --since or --until")
	}
	if command.Name() == "slog" && strings.TrimSpace(search) == "" {
		return til.EntryQuery{}, errors.New("search query cannot be empty")
	}

	query := til.EntryQuery{
		Limit:       options.number,
		Search:      strings.TrimSpace(search),
		OldestFirst: options.reverse,
	}
	if options.all {
		query.Limit = 0
	}

	if options.date != "" {
		start, err := parseCalendarDate("--date", options.date)
		if err != nil {
			return til.EntryQuery{}, err
		}
		query.OnDate = start.Format(calendarDateLayout)
		return query, nil
	}
	if options.since != "" {
		since, err := parseCalendarDate("--since", options.since)
		if err != nil {
			return til.EntryQuery{}, err
		}
		query.SinceDate = since.Format(calendarDateLayout)
	}
	if options.until != "" {
		until, err := parseCalendarDate("--until", options.until)
		if err != nil {
			return til.EntryQuery{}, err
		}
		before := until.AddDate(0, 0, 1)
		query.BeforeDate = before.Format(calendarDateLayout)
	}
	if query.SinceDate != "" && query.BeforeDate != "" && query.SinceDate >= query.BeforeDate {
		return til.EntryQuery{}, errors.New("--since must not be later than --until")
	}
	return query, nil
}

func parseCalendarDate(flagName, value string) (time.Time, error) {
	date, err := time.ParseInLocation(calendarDateLayout, value, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf(
			"%s must use YYYY-MM-DD (received %q)",
			flagName,
			value,
		)
	}
	return date, nil
}

func writeEntriesTable(output io.Writer, entries []til.Entry) error {
	table := tabwriter.NewWriter(output, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(table, "COMMIT\tDATE\tMESSAGE\tFILES"); err != nil {
		return err
	}
	for _, entry := range entries {
		if _, err := fmt.Fprintf(
			table,
			"%s\t%s\t%s\t%s\n",
			entry.CommitID,
			entry.Date.Format(calendarDateLayout),
			singleLine(entry.Message),
			formatFiles(entry.Files),
		); err != nil {
			return err
		}
	}
	return table.Flush()
}

func writeEntriesLong(output io.Writer, entries []til.Entry) error {
	for index, entry := range entries {
		if index > 0 {
			if _, err := fmt.Fprintln(output); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(output, "commit %s\n", entry.CommitID); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(output, "Date:   %s\n", entry.Date.Format(time.RFC3339)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(output, "Files:  %s\n", formatFiles(entry.Files)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(output, "\n    %s\n", entry.Message); err != nil {
			return err
		}
		if entry.MessageBody != "" {
			for _, line := range strings.Split(entry.MessageBody, "\n") {
				if _, err := fmt.Fprintf(output, "    %s\n", line); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func writeEntriesJSON(output io.Writer, entries []til.Entry) error {
	result := make([]logEntryJSON, len(entries))
	for i, entry := range entries {
		result[i] = logEntryJSON{
			CommitID:     entry.CommitID,
			Date:         entry.Date.Format(time.RFC3339Nano),
			Message:      entry.Message,
			MessageBody:  entry.MessageBody,
			Files:        append([]string(nil), entry.Files...),
			IsCommitted:  entry.IsCommitted,
			NotionSynced: entry.NotionSynced,
		}
		if result[i].Files == nil {
			result[i].Files = []string{}
		}
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("encode entries as JSON: %w", err)
	}
	return nil
}

func singleLine(value string) string {
	replacer := strings.NewReplacer("\t", " ", "\r", " ", "\n", " ")
	return replacer.Replace(value)
}

func formatFiles(files []string) string {
	if len(files) == 0 {
		return "-"
	}
	names := make([]string, len(files))
	for i, fileName := range files {
		names[i] = filepath.Base(fileName)
	}
	return strings.Join(names, ", ")
}
