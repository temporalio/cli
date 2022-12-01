package common

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gogo/protobuf/proto"
	"github.com/temporalio/cli/headers"
	tColor "github.com/temporalio/tctl-kit/pkg/color"
	"github.com/urfave/cli/v2"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	"go.temporal.io/server/common/codec"
	"go.temporal.io/server/common/payloads"
)

// ColorEvent takes an event and return string with color
// Event with color mapping rules:
//
//	Failed - red
//	Timeout - yellow
//	Canceled - magenta
//	Completed - green
//	Started - blue
//	Others - default (white/black)
func ColorEvent(e *historypb.HistoryEvent) string {
	var data string
	switch e.GetEventType() {
	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED:
		data = color.BlueString(e.EventType.String())

	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED:
		data = color.GreenString(e.EventType.String())

	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED:
		data = color.RedString(e.EventType.String())

	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_TIMED_OUT:
		data = color.YellowString(e.EventType.String())

	case enumspb.EVENT_TYPE_WORKFLOW_TASK_SCHEDULED:
		data = e.EventType.String()

	case enumspb.EVENT_TYPE_WORKFLOW_TASK_STARTED:
		data = e.EventType.String()

	case enumspb.EVENT_TYPE_WORKFLOW_TASK_COMPLETED:
		data = e.EventType.String()

	case enumspb.EVENT_TYPE_WORKFLOW_TASK_TIMED_OUT:
		data = color.YellowString(e.EventType.String())

	case enumspb.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED:
		data = e.EventType.String()

	case enumspb.EVENT_TYPE_ACTIVITY_TASK_STARTED:
		data = e.EventType.String()

	case enumspb.EVENT_TYPE_ACTIVITY_TASK_COMPLETED:
		data = e.EventType.String()

	case enumspb.EVENT_TYPE_ACTIVITY_TASK_FAILED:
		data = color.RedString(e.EventType.String())

	case enumspb.EVENT_TYPE_ACTIVITY_TASK_TIMED_OUT:
		data = color.YellowString(e.EventType.String())

	case enumspb.EVENT_TYPE_ACTIVITY_TASK_CANCEL_REQUESTED:
		data = e.EventType.String()

	case enumspb.EVENT_TYPE_ACTIVITY_TASK_CANCELED:
		data = e.EventType.String()

	case enumspb.EVENT_TYPE_TIMER_STARTED:
		data = e.EventType.String()

	case enumspb.EVENT_TYPE_TIMER_FIRED:
		data = e.EventType.String()

	case enumspb.EVENT_TYPE_TIMER_CANCELED:
		data = color.MagentaString(e.EventType.String())

	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_CANCEL_REQUESTED:
		data = e.EventType.String()

	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_CANCELED:
		data = color.MagentaString(e.EventType.String())

	case enumspb.EVENT_TYPE_REQUEST_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_INITIATED:
		data = e.EventType.String()

	case enumspb.EVENT_TYPE_REQUEST_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED:
		data = color.RedString(e.EventType.String())

	case enumspb.EVENT_TYPE_EXTERNAL_WORKFLOW_EXECUTION_CANCEL_REQUESTED:
		data = e.EventType.String()

	case enumspb.EVENT_TYPE_MARKER_RECORDED:
		data = e.EventType.String()

	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_SIGNALED:
		data = e.EventType.String()

	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_TERMINATED:
		data = e.EventType.String()

	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_CONTINUED_AS_NEW:
		data = e.EventType.String()

	case enumspb.EVENT_TYPE_START_CHILD_WORKFLOW_EXECUTION_INITIATED:
		data = e.EventType.String()

	case enumspb.EVENT_TYPE_START_CHILD_WORKFLOW_EXECUTION_FAILED:
		data = color.RedString(e.EventType.String())

	case enumspb.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_STARTED:
		data = color.BlueString(e.EventType.String())

	case enumspb.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_COMPLETED:
		data = color.GreenString(e.EventType.String())

	case enumspb.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_FAILED:
		data = color.RedString(e.EventType.String())

	case enumspb.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_CANCELED:
		data = color.MagentaString(e.EventType.String())

	case enumspb.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_TIMED_OUT:
		data = color.YellowString(e.EventType.String())

	case enumspb.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_TERMINATED:
		data = e.EventType.String()

	case enumspb.EVENT_TYPE_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_INITIATED:
		data = e.EventType.String()

	case enumspb.EVENT_TYPE_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED:
		data = color.RedString(e.EventType.String())

	case enumspb.EVENT_TYPE_EXTERNAL_WORKFLOW_EXECUTION_SIGNALED:
		data = e.EventType.String()

	case enumspb.EVENT_TYPE_UPSERT_WORKFLOW_SEARCH_ATTRIBUTES:
		data = e.EventType.String()

	default:
		data = e.EventType.String()
	}
	return data
}

func GetCurrentUserFromEnv() string {
	for _, n := range envKeysForUserName {
		if len(os.Getenv(n)) > 0 {
			return os.Getenv(n)
		}
	}
	return "unknown"
}

func PrettyPrintJSONObject(o interface{}) {
	var b []byte
	var err error
	if pb, ok := o.(proto.Message); ok {
		encoder := codec.NewJSONPBIndentEncoder("  ")
		b, err = encoder.Encode(pb)
	} else {
		b, err = json.MarshalIndent(o, "", "  ")
	}

	if err != nil {
		fmt.Printf("Error when try to print pretty: %v", err)
		fmt.Println(o)
	}
	_, _ = os.Stdout.Write(b)
	fmt.Println()
}

func RequiredFlag(c *cli.Context, optionName string) (string, error) {
	value := c.String(optionName)
	if len(value) == 0 {
		return "", fmt.Errorf("option %v is required: ", tColor.Yellow(c, "--%v", optionName))
	}
	return value, nil
}

func FormatTime(t time.Time, onlyTime bool) string {
	var result string
	if onlyTime {
		result = t.Format(defaultTimeFormat)
	} else {
		result = t.Format(defaultDateTimeFormat)
	}
	return result
}

func ParseTime(timeStr string, defaultValue time.Time, now time.Time) (time.Time, error) {
	if len(timeStr) == 0 {
		return defaultValue, nil
	}

	// try to parse
	parsedTime, err := time.Parse(defaultDateTimeFormat, timeStr)
	if err == nil {
		return parsedTime, nil
	}

	// treat as raw unix time
	resultValue, err := strconv.ParseInt(timeStr, 10, 64)
	if err == nil {
		return time.Unix(0, resultValue).UTC(), nil
	}

	// treat as time range format
	parsedTime, err = parseTimeRange(timeStr, now)
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse time '%s', use UTC format '2006-01-02T15:04:05', "+
			"time range or raw UnixNano directly. See help for more details: %s", timeStr, err)
	}
	return parsedTime, nil
}

// parseTimeRange parses a given time duration string (in format X<time-duration>) and
// returns parsed timestamp given that duration in the past from current time.
// All valid values must contain a number followed by a time-duration, from the following list (long form/short form):
// - second/s
// - minute/m
// - hour/h
// - day/d
// - week/w
// - month/M
// - year/y
// For example, possible input values, and their result:
// - "3d" or "3day" --> three days --> time.Now().UTC().Add(-3 * 24 * time.Hour)
// - "2m" or "2minute" --> two minutes --> time.Now().UTC().Add(-2 * time.Minute)
// - "1w" or "1week" --> one week --> time.Now().UTC().Add(-7 * 24 * time.Hour)
// - "30s" or "30second" --> thirty seconds --> time.Now().UTC().Add(-30 * time.Second)
// Note: Duration strings are case-sensitive, and should be used as mentioned above only.
// Limitation: Value of numerical multiplier, X should be in b/w 0 - 1e6 (1 million), boundary values excluded i.e.
// 0 < X < 1e6. Also, the maximum time in the past can be 1 January 1970 00:00:00 UTC (epoch time),
// so giving "1000y" will result in epoch time.
func parseTimeRange(timeRange string, now time.Time) (time.Time, error) {
	match, err := regexp.MatchString(defaultDateTimeRangeShortRE, timeRange)
	if !match { // fallback on to check if it's of longer notation
		_, err = regexp.MatchString(defaultDateTimeRangeLongRE, timeRange)
	}
	if err != nil {
		return time.Time{}, err
	}

	re, _ := regexp.Compile(defaultDateTimeRangeNum)
	idx := re.FindStringSubmatchIndex(timeRange)
	if idx == nil {
		return time.Time{}, fmt.Errorf("cannot parse timeRange %s", timeRange)
	}

	num, err := strconv.Atoi(timeRange[idx[0]:idx[1]])
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse timeRange %s", timeRange)
	}
	if num >= 1e6 {
		return time.Time{}, fmt.Errorf("invalid time-duation multiplier %d, allowed range is 0 < multiplier < 1000000", num)
	}

	dur, err := parseTimeDuration(timeRange[idx[1]:])
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse timeRange %s", timeRange)
	}

	res := now.Add(time.Duration(-num) * dur) // using server's local timezone
	epochTime := time.Unix(0, 0).UTC()
	if res.Before(epochTime) {
		res = epochTime
	}
	return res, nil
}

// parseTimeDuration parses the given time duration in either short or long convention
// and returns the time.Duration
// Valid values (long notation/short notation):
// - second/s
// - minute/m
// - hour/h
// - day/d
// - week/w
// - month/M
// - year/y
// NOTE: the input "duration" is case-sensitive
func parseTimeDuration(duration string) (dur time.Duration, err error) {
	switch duration {
	case "s", "second":
		dur = time.Second
	case "m", "minute":
		dur = time.Minute
	case "h", "hour":
		dur = time.Hour
	case "d", "day":
		dur = day
	case "w", "week":
		dur = week
	case "M", "month":
		dur = month
	case "y", "year":
		dur = year
	default:
		err = fmt.Errorf("unknown time duration %s", duration)
	}
	return
}

func GetCliIdentity() string {
	hostName, err := os.Hostname()
	if err != nil {
		hostName = "Unknown"
	}
	userName := "unknown"
	if u, err := user.Current(); err == nil {
		userName = u.Username
	}
	return fmt.Sprintf("tctl:%s@%s", userName, hostName)
}

func NewContext(c *cli.Context) (context.Context, context.CancelFunc) {
	return NewContextWithTimeout(c, defaultContextTimeout)
}

func NewContextForLongPoll(c *cli.Context) (context.Context, context.CancelFunc) {
	return NewContextWithTimeout(c, defaultContextTimeoutForLongPoll)
}

func NewIndefiniteContext(c *cli.Context) (context.Context, context.CancelFunc) {
	if c.IsSet(FlagContextTimeout) {
		timeout := time.Duration(c.Int(FlagContextTimeout)) * time.Second
		return NewContextWithTimeoutAndCLIHeaders(timeout)
	}

	return NewContextWithCLIHeaders()
}

func NewContextWithTimeout(c *cli.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if c.IsSet(FlagContextTimeout) {
		timeout = time.Duration(c.Int(FlagContextTimeout)) * time.Second
	}

	return NewContextWithTimeoutAndCLIHeaders(timeout)
}

// NewContextWithCLIHeaders creates context with version headers for CLI.
func NewContextWithCLIHeaders() (context.Context, context.CancelFunc) {
	return context.WithCancel(headers.SetCLIVersions(context.Background()))
}

// NewContextWithTimeoutAndCLIHeaders creates context with timeout and version headers for CLI.
func NewContextWithTimeoutAndCLIHeaders(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(headers.SetCLIVersions(context.Background()), timeout)
}

func UnmarshalInputsFromCLI(c *cli.Context) ([]interface{}, error) {
	jsonsRaw, err := readJSONInputs(c)
	if err != nil {
		return nil, err
	}

	var result []interface{}
	for _, jsonRaw := range jsonsRaw {
		if jsonRaw == nil {
			result = append(result, nil)
		} else {
			var j interface{}
			if err := json.Unmarshal(jsonRaw, &j); err != nil {
				return nil, fmt.Errorf("input is not valid JSON: %w", err)
			}
			result = append(result, j)
		}

	}

	return result, nil
}

// process and validate input provided through cmd or file
func ProcessJSONInput(c *cli.Context) (*commonpb.Payloads, error) {
	jsons, err := UnmarshalInputsFromCLI(c)
	if err != nil {
		return nil, err
	}
	p, err := payloads.Encode(jsons...)
	if err != nil {
		return nil, fmt.Errorf("unable to encode input: %w", err)
	}

	return p, nil
}

// read multiple inputs presented in json format
func readJSONInputs(c *cli.Context) ([][]byte, error) {
	if c.IsSet(FlagInput) {
		inputsG := c.Generic(FlagInput)

		var inputs *cli.StringSlice
		var ok bool
		if inputs, ok = inputsG.(*cli.StringSlice); !ok {
			// input could be provided as StringFlag instead of StringSliceFlag
			ss := cli.StringSlice{}
			ss.Set(fmt.Sprintf("%v", inputsG))
			inputs = &ss
		}

		var inputsRaw [][]byte
		for _, i := range inputs.Value() {
			if strings.EqualFold(i, "null") {
				inputsRaw = append(inputsRaw, []byte(nil))
			} else {
				inputsRaw = append(inputsRaw, []byte(i))
			}
		}

		return inputsRaw, nil
	} else if c.IsSet(FlagInputFile) {
		inputFile := c.String(FlagInputFile)
		// This method is purely used to parse input from the CLI. The input comes from a trusted user
		// #nosec
		data, err := os.ReadFile(inputFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read input file: %w", err)
		}
		return [][]byte{data}, nil
	}
	return nil, nil
}

func Truncate(str string) string {
	if len(str) > maxOutputStringLength {
		return str[:maxOutputStringLength]
	}
	return str
}

func StringToEnum(search string, candidates map[string]int32) (int32, error) {
	if search == "" {
		return 0, nil
	}

	var candidateNames []string
	for key, value := range candidates {
		if strings.EqualFold(key, search) {
			return value, nil
		}
		candidateNames = append(candidateNames, key)
	}

	return 0, fmt.Errorf("could not find corresponding candidate for %s. Possible candidates: %q", search, candidateNames)
}

func AllowedEnumValues(names map[int32]string) []string {
	result := make([]string, len(names)-1)
	for i := 0; i < len(result); i++ {
		result[i] = names[int32(i+1)]
	}
	return result
}

func PromptYes(msg string, autoConfirm bool) bool {
	return Prompt(msg, autoConfirm, "yes", "y")
}

// Prompt user to confirm/deny action. Supports empty expectedInputs.
func Prompt(msg string, autoConfirm bool, expectedInputs ...string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(msg, " ")
	var text string
	if autoConfirm {
		if len(expectedInputs) > 0 {
			text = expectedInputs[0]
		}
		fmt.Print(text)
	} else {
		text, _ = reader.ReadString('\n')
	}
	fmt.Println()

	if len(expectedInputs) == 0 {
		return true
	}

	textLower := strings.ToLower(strings.TrimSpace(text))
	for _, expectedInput := range expectedInputs {
		if expectedInput == textLower {
			return true
		}
	}
	return false
}

func EnsureNonNil[T any, P ~*T](ptr *P) {
	if *ptr == nil {
		*ptr = new(T)
	}
}

// SplitKeyValuePairs parses key=value pairs
func SplitKeyValuePairs(kvs []string) (map[string]string, error) {
	pairs := make(map[string]string, len(kvs))
	for _, v := range kvs {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("unable to split key=value pair: %v", v)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, errors.New("empty key is not allowed")
		}

		pairs[key] = value
	}

	return pairs, nil
}

func AddBeforeHandler(cmd *cli.Command, h func(*cli.Context) error) {
	// only populate leaf commands
	if len(cmd.Subcommands) == 0 {
		_before := cmd.Before
		cmd.Before = func(c *cli.Context) error {
			if _before != nil {
				err := _before(c)
				if err != nil {
					return err
				}
			}

			return h(c)
		}
	}

	for _, subcmd := range cmd.Subcommands {
		AddBeforeHandler(subcmd, h)
	}
}
