// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package datasetexporter

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/scalyr/dataset-go/pkg/api/add_events"
	"github.com/scalyr/dataset-go/pkg/api/request"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/exportertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/testdata"
)

func TestCreateLogsExporter(t *testing.T) {
	ctx := context.Background()
	createSettings := exportertest.NewNopCreateSettings()
	tests := createExporterTests()

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			logs, err := createLogsExporter(ctx, createSettings, tt.config)

			if err == nil {
				assert.Nil(t, tt.expectedError, tt.name)
				assert.NotNil(t, logs, tt.name)
			} else {
				if tt.expectedError == nil {
					assert.Nil(t, err, tt.name)
				} else {
					assert.Equal(t, tt.expectedError.Error(), err.Error(), tt.name)
					assert.Nil(t, logs, tt.name)
				}
			}
		})
	}
}

func TestBuildBody(t *testing.T) {
	slice := pcommon.NewValueSlice()
	err := slice.FromRaw([]any{1, 2, 3})
	assert.NoError(t, err)

	bytes := pcommon.NewValueBytes()
	err = bytes.FromRaw([]byte{byte(65), byte(66), byte(67)})
	assert.NoError(t, err)
	tests := []struct {
		body      pcommon.Value
		valueType string
		message   string
	}{
		{
			body:      pcommon.NewValueEmpty(),
			valueType: "empty",
			message:   "",
		},
		{
			body:      pcommon.NewValueStr("foo"),
			valueType: "string",
			message:   "foo",
		},
		{
			body:      pcommon.NewValueBool(true),
			valueType: "bool",
			message:   "true",
		},
		{
			body:      pcommon.NewValueDouble(42.5),
			valueType: "double",
			message:   "42.5",
		},
		{
			body:      pcommon.NewValueInt(42),
			valueType: "int",
			message:   "42",
		},
		{
			body:      bytes,
			valueType: "bytes",
			message:   "QUJD",
		},
		{
			body:      slice,
			valueType: "simpleMap",
			message:   "[1,2,3]",
		},
	}

	settings := newDefaultLogsSettings()

	for _, tt := range tests {
		t.Run(tt.valueType, func(*testing.T) {
			attrs := make(map[string]interface{})
			msg := buildBody(settings, attrs, tt.body)

			assert.Equal(t, tt.message, msg, tt.valueType)
		})
	}
}

func TestBuildBodyMap(t *testing.T) {
	m := pcommon.NewValueMap()
	err := m.FromRaw(map[string]any{
		"scalar": "scalar-value",
		"map": map[string]any{
			"m1": "v1",
			"m2": "v2",
		},
		"array": []any{1, 2, 3},
	})
	if assert.NoError(t, err) {
		settings := newDefaultLogsSettings()
		settings.DecomposeComplexMessageField = true
		attrs := make(map[string]interface{})

		msg := buildBody(settings, attrs, m)
		expectedAttrs := make(map[string]interface{})
		expectedAttrs["body.map.scalar"] = "scalar-value"
		expectedAttrs["body.map.map.m1"] = "v1"
		expectedAttrs["body.map.map.m2"] = "v2"
		expectedAttrs["body.map.array.0"] = int64(1)
		expectedAttrs["body.map.array.1"] = int64(2)
		expectedAttrs["body.map.array.2"] = int64(3)

		expectedMsg := "{\"array\":[1,2,3],\"map\":{\"m1\":\"v1\",\"m2\":\"v2\"},\"scalar\":\"scalar-value\"}"

		assert.Equal(t, expectedMsg, msg)
		assert.Equal(t, expectedAttrs, attrs)
	}
}

var testLThread = &add_events.Thread{
	Id:   "TL",
	Name: "logs",
}

var testLLog = &add_events.Log{
	Id:    "LL",
	Attrs: map[string]interface{}{},
}

var testServerHost = "foo"

var testLEventRaw = &add_events.Event{
	Thread:     testLThread.Id,
	Log:        testLLog.Id,
	Sev:        3,
	Ts:         "1581452773000000789",
	ServerHost: testServerHost,
	Attrs: map[string]interface{}{
		"app":                      "server",
		"instance_num":             int64(1),
		"dropped_attributes_count": uint32(1),
		"message":                  "This is a log message",
		"span_id":                  "0102040800000000",
		"trace_id":                 "08040201000000000000000000000000",
	},
}

var testLEventRawWithScopeInfo = &add_events.Event{
	Thread:     testLThread.Id,
	Log:        testLLog.Id,
	Sev:        3,
	Ts:         "1581452773000000789",
	ServerHost: testServerHost,
	Attrs: map[string]interface{}{
		"app":                      "server",
		"instance_num":             int64(1),
		"dropped_attributes_count": uint32(1),
		"scope.name":               "test-scope",
		"message":                  "This is a log message",
		"span_id":                  "0102040800000000",
		"trace_id":                 "08040201000000000000000000000000",
	},
}

var testLEventReq = &add_events.Event{
	Thread: testLEventRaw.Thread,
	Log:    testLEventRaw.Log,
	Sev:    testLEventRaw.Sev,
	Ts:     testLEventRaw.Ts,
	Attrs: map[string]interface{}{
		add_events.AttrOrigServerHost: testServerHost,
		"app":                         "server",
		"instance_num":                float64(1),
		"dropped_attributes_count":    float64(1),
		"message":                     "This is a log message",
		"span_id":                     "0102040800000000",
		"trace_id":                    "08040201000000000000000000000000",
		"bundle_key":                  "d41d8cd98f00b204e9800998ecf8427e",
	},
}

func TestBuildEventFromLog(t *testing.T) {
	tests := []struct {
		name        string
		settings    LogsSettings
		complexBody bool
		expected    add_events.EventAttrs
	}{
		{
			name:        "DefaultSimpleBody",
			settings:    newDefaultLogsSettings(),
			complexBody: false,
			expected: add_events.EventAttrs{
				"message": "This is a log message",

				"string":                     "stringA",
				"map.map_empty":              nil,
				"map.map_string":             "map_stringA",
				"map.map_map.map_map_string": "map_map_stringA",
				"slice.0":                    "slice_stringA",

				"scope.attributes.string":                     "stringS",
				"scope.attributes.map.map_empty":              nil,
				"scope.attributes.map.map_string":             "map_stringS",
				"scope.attributes.map.map_map.map_map_string": "map_map_stringS",
				"scope.attributes.slice.0":                    "slice_stringS",
			},
		},
		{
			name:        "DefaultComplexBody",
			settings:    newDefaultLogsSettings(),
			complexBody: true,
			expected: add_events.EventAttrs{
				"message": "{\"empty_map\":{},\"empty_slice\":[],\"map\":{\"map_empty\":null,\"map_map\":{\"map_map_string\":\"map_map_stringM\"},\"map_string\":\"map_stringM\"},\"slice\":[\"slice_stringM\"],\"string\":\"stringM\"}",

				"string":                     "stringA",
				"map.map_empty":              nil,
				"map.map_string":             "map_stringA",
				"map.map_map.map_map_string": "map_map_stringA",
				"slice.0":                    "slice_stringA",

				"scope.attributes.string":                     "stringS",
				"scope.attributes.map.map_empty":              nil,
				"scope.attributes.map.map_string":             "map_stringS",
				"scope.attributes.map.map_map.map_map_string": "map_map_stringS",
				"scope.attributes.slice.0":                    "slice_stringS",
			},
		},
		{
			name:        "FullSimpleBody",
			complexBody: false,
			settings: LogsSettings{
				ExportResourceInfo:             true,
				ExportResourcePrefix:           "__R__",
				ExportScopeInfo:                true,
				ExportScopePrefix:              "__S__",
				ExportSeparator:                ".SEP.",
				ExportDistinguishingSuffix:     ".SUF.",
				DecomposeComplexMessageField:   true,
				DecomposedComplexMessagePrefix: "__M__",
			},
			expected: add_events.EventAttrs{
				"message": "This is a log message",

				"string":                             "stringA",
				"map.SEP.map_empty":                  nil,
				"map.SEP.map_string":                 "map_stringA",
				"map.SEP.map_map.SEP.map_map_string": "map_map_stringA",
				"slice.SEP.0":                        "slice_stringA",

				"__S__string":                             "stringS",
				"__S__map.SEP.map_empty":                  nil,
				"__S__map.SEP.map_string":                 "map_stringS",
				"__S__map.SEP.map_map.SEP.map_map_string": "map_map_stringS",
				"__S__slice.SEP.0":                        "slice_stringS",

				"__R__string":                             "stringR",
				"__R__map.SEP.map_empty":                  nil,
				"__R__map.SEP.map_string":                 "map_stringR",
				"__R__map.SEP.map_map.SEP.map_map_string": "map_map_stringR",
				"__R__slice.SEP.0":                        "slice_stringR",

				"__R__resource-attr": "resource-attr-val-1",
			},
		},
		{
			name:        "FullComplexBody",
			complexBody: true,
			settings: LogsSettings{
				ExportResourceInfo:             true,
				ExportResourcePrefix:           "__R__",
				ExportScopeInfo:                true,
				ExportScopePrefix:              "__S__",
				ExportSeparator:                ".SEP.",
				ExportDistinguishingSuffix:     ".SUF.",
				DecomposeComplexMessageField:   true,
				DecomposedComplexMessagePrefix: "__M__",
			},
			expected: add_events.EventAttrs{
				"message":                 "{\"empty_map\":{},\"empty_slice\":[],\"map\":{\"map_empty\":null,\"map_map\":{\"map_map_string\":\"map_map_stringM\"},\"map_string\":\"map_stringM\"},\"slice\":[\"slice_stringM\"],\"string\":\"stringM\"}",
				"__M__string":             "stringM",
				"__M__map.SEP.map_empty":  nil,
				"__M__map.SEP.map_string": "map_stringM",
				"__M__map.SEP.map_map.SEP.map_map_string": "map_map_stringM",
				"__M__slice.SEP.0":                        "slice_stringM",

				"string":                             "stringA",
				"map.SEP.map_empty":                  nil,
				"map.SEP.map_string":                 "map_stringA",
				"map.SEP.map_map.SEP.map_map_string": "map_map_stringA",
				"slice.SEP.0":                        "slice_stringA",

				"__S__string":                             "stringS",
				"__S__map.SEP.map_empty":                  nil,
				"__S__map.SEP.map_string":                 "map_stringS",
				"__S__map.SEP.map_map.SEP.map_map_string": "map_map_stringS",
				"__S__slice.SEP.0":                        "slice_stringS",

				"__R__string":                             "stringR",
				"__R__map.SEP.map_empty":                  nil,
				"__R__map.SEP.map_string":                 "map_stringR",
				"__R__map.SEP.map_map.SEP.map_map_string": "map_map_stringR",
				"__R__slice.SEP.0":                        "slice_stringR",

				"__R__resource-attr": "resource-attr-val-1",
			},
		},
		{
			name:        "Minimal",
			complexBody: false,
			settings: LogsSettings{
				ExportResourceInfo:           false,
				ExportResourcePrefix:         "__R__",
				ExportScopeInfo:              false,
				ExportScopePrefix:            "__S__",
				ExportSeparator:              ".SEP.",
				ExportDistinguishingSuffix:   ".SUP.",
				DecomposeComplexMessageField: false,
			},
			expected: add_events.EventAttrs{
				"message": "This is a log message",

				"string":                             "stringA",
				"map.SEP.map_empty":                  nil,
				"map.SEP.map_string":                 "map_stringA",
				"map.SEP.map_map.SEP.map_map_string": "map_map_stringA",
				"slice.SEP.0":                        "slice_stringA",
			},
		},
		{
			name:        "EmptyPrefixesAndNonEmptySuffixSimpleBody",
			complexBody: false,
			settings: LogsSettings{
				ExportResourceInfo:           true,
				ExportResourcePrefix:         "",
				ExportScopeInfo:              true,
				ExportScopePrefix:            "",
				ExportSeparator:              ".SEP.",
				ExportDistinguishingSuffix:   ".SUF.",
				DecomposeComplexMessageField: true,
			},
			expected: add_events.EventAttrs{
				"message": "This is a log message",

				"string":                             "stringR",
				"map.SEP.map_empty":                  nil,
				"map.SEP.map_string":                 "map_stringR",
				"map.SEP.map_map.SEP.map_map_string": "map_map_stringR",
				"slice.SEP.0":                        "slice_stringR",

				"string.SUF.":                             "stringS",
				"map.SEP.map_empty.SUF.":                  nil,
				"map.SEP.map_string.SUF.":                 "map_stringS",
				"map.SEP.map_map.SEP.map_map_string.SUF.": "map_map_stringS",
				"slice.SEP.0.SUF.":                        "slice_stringS",

				"string.SUF..SUF.":                             "stringA",
				"map.SEP.map_empty.SUF..SUF.":                  nil,
				"map.SEP.map_string.SUF..SUF.":                 "map_stringA",
				"map.SEP.map_map.SEP.map_map_string.SUF..SUF.": "map_map_stringA",
				"slice.SEP.0.SUF..SUF.":                        "slice_stringA",

				"resource-attr": "resource-attr-val-1",
			},
		},
		{
			name:        "EmptyPrefixesAndNonEmptySuffixComplexBody",
			complexBody: true,
			settings: LogsSettings{
				ExportResourceInfo:             true,
				ExportResourcePrefix:           "",
				ExportScopeInfo:                true,
				ExportScopePrefix:              "",
				ExportSeparator:                ".SEP.",
				ExportDistinguishingSuffix:     ".SUF.",
				DecomposeComplexMessageField:   true,
				DecomposedComplexMessagePrefix: "",
			},
			expected: add_events.EventAttrs{
				"message": "{\"empty_map\":{},\"empty_slice\":[],\"map\":{\"map_empty\":null,\"map_map\":{\"map_map_string\":\"map_map_stringM\"},\"map_string\":\"map_stringM\"},\"slice\":[\"slice_stringM\"],\"string\":\"stringM\"}",

				"string":                             "stringM",
				"map.SEP.map_empty":                  nil,
				"map.SEP.map_string":                 "map_stringM",
				"map.SEP.map_map.SEP.map_map_string": "map_map_stringM",
				"slice.SEP.0":                        "slice_stringM",

				"string.SUF.":                             "stringR",
				"map.SEP.map_empty.SUF.":                  nil,
				"map.SEP.map_string.SUF.":                 "map_stringR",
				"map.SEP.map_map.SEP.map_map_string.SUF.": "map_map_stringR",
				"slice.SEP.0.SUF.":                        "slice_stringR",

				"string.SUF..SUF.":                             "stringS",
				"map.SEP.map_empty.SUF..SUF.":                  nil,
				"map.SEP.map_string.SUF..SUF.":                 "map_stringS",
				"map.SEP.map_map.SEP.map_map_string.SUF..SUF.": "map_map_stringS",
				"slice.SEP.0.SUF..SUF.":                        "slice_stringS",

				"string.SUF..SUF..SUF.":                             "stringA",
				"map.SEP.map_empty.SUF..SUF..SUF.":                  nil,
				"map.SEP.map_string.SUF..SUF..SUF.":                 "map_stringA",
				"map.SEP.map_map.SEP.map_map_string.SUF..SUF..SUF.": "map_map_stringA",
				"slice.SEP.0.SUF..SUF..SUF.":                        "slice_stringA",

				"resource-attr": "resource-attr-val-1",
			},
		},
		{
			name:        "EmptyPrefixesAndEmptySuffixSimpleBody",
			complexBody: false,
			settings: LogsSettings{
				ExportResourceInfo:             true,
				ExportResourcePrefix:           "",
				ExportScopeInfo:                true,
				ExportScopePrefix:              "",
				ExportSeparator:                ".SEP.",
				ExportDistinguishingSuffix:     "",
				DecomposeComplexMessageField:   true,
				DecomposedComplexMessagePrefix: "",
			},
			expected: add_events.EventAttrs{
				"message": "This is a log message",

				"string":                             "stringA",
				"map.SEP.map_empty":                  nil,
				"map.SEP.map_string":                 "map_stringA",
				"map.SEP.map_map.SEP.map_map_string": "map_map_stringA",
				"slice.SEP.0":                        "slice_stringA",

				"resource-attr": "resource-attr-val-1",
			},
		},
		{
			name:        "EmptyPrefixesAndEmptySuffixSimpleBody",
			complexBody: true,
			settings: LogsSettings{
				ExportResourceInfo:             true,
				ExportResourcePrefix:           "",
				ExportScopeInfo:                true,
				ExportScopePrefix:              "",
				ExportSeparator:                ".SEP.",
				ExportDistinguishingSuffix:     "",
				DecomposeComplexMessageField:   true,
				DecomposedComplexMessagePrefix: "",
			},
			expected: add_events.EventAttrs{
				"message": "{\"empty_map\":{},\"empty_slice\":[],\"map\":{\"map_empty\":null,\"map_map\":{\"map_map_string\":\"map_map_stringM\"},\"map_string\":\"map_stringM\"},\"slice\":[\"slice_stringM\"],\"string\":\"stringM\"}",

				"string":                             "stringA",
				"map.SEP.map_empty":                  nil,
				"map.SEP.map_string":                 "map_stringA",
				"map.SEP.map_map.SEP.map_map_string": "map_map_stringA",
				"slice.SEP.0":                        "slice_stringA",

				"resource-attr": "resource-attr-val-1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lr := testdata.GenerateLogsOneEmptyLogRecord()
			ld := lr.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
			resource := lr.ResourceLogs().At(0).Resource()
			scope := lr.ResourceLogs().At(0).ScopeLogs().At(0).Scope()

			fillAttributes(resource.Attributes(), false, "R")
			fillAttributes(scope.Attributes(), false, "S")
			fillAttributes(ld.Attributes(), false, "A")
			ld.SetTimestamp(testdata.TestLogTimestamp)

			if tt.complexBody {
				m := ld.Body().SetEmptyMap()
				fillAttributes(m, false, "M")
			} else {
				ld.Body().SetStr("This is a log message")
			}

			expected := &add_events.EventBundle{
				Event: &add_events.Event{
					Thread:     testLThread.Id,
					Log:        testLLog.Id,
					Sev:        3,
					Ts:         "1581452773000000789",
					ServerHost: testServerHost,
					Attrs:      tt.expected,
				},
				Thread: testLThread,
				Log:    testLLog,
			}

			was := buildEventFromLog(
				ld,
				resource,
				scope,
				testServerHost,
				tt.settings,
			)

			assert.Equal(t, expected, was)
		})
	}

}

func TestBuildEventFromLogExportResources(t *testing.T) {
	lr := testdata.GenerateLogsOneLogRecord()
	ld := lr.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)

	defaultAttrs := testLEventRaw.Attrs
	defaultAttrs["resource-attr"] = "resource-attr-val-1"

	expected := &add_events.EventBundle{
		Event: &add_events.Event{
			Thread:     testLEventRaw.Thread,
			Log:        testLEventRaw.Log,
			Sev:        testLEventRaw.Sev,
			Ts:         testLEventRaw.Ts,
			Attrs:      defaultAttrs,
			ServerHost: testLEventRaw.ServerHost,
		},
		Thread: testLThread,
		Log:    testLLog,
	}
	was := buildEventFromLog(
		ld,
		lr.ResourceLogs().At(0).Resource(),
		lr.ResourceLogs().At(0).ScopeLogs().At(0).Scope(),
		testServerHost,
		LogsSettings{
			ExportResourceInfo: true,
			ExportScopeInfo:    true,
		},
	)

	assert.Equal(t, expected, was)
}

func TestBuildEventFromLogExportScopeInfo(t *testing.T) {
	lr := testdata.GenerateLogsOneLogRecord()
	ld := lr.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)

	scope := pcommon.NewInstrumentationScope()
	scope.SetName("test-scope")
	scope.SetDroppedAttributesCount(11)

	expected := &add_events.EventBundle{
		Event: &add_events.Event{
			Thread:     testLEventRawWithScopeInfo.Thread,
			Log:        testLEventRawWithScopeInfo.Log,
			Sev:        testLEventRawWithScopeInfo.Sev,
			Ts:         testLEventRawWithScopeInfo.Ts,
			Attrs:      testLEventRawWithScopeInfo.Attrs,
			ServerHost: testLEventRaw.ServerHost,
		},
		Thread: testLThread,
		Log:    testLLog,
	}
	was := buildEventFromLog(
		ld,
		lr.ResourceLogs().At(0).Resource(),
		scope,
		testServerHost,
		LogsSettings{
			ExportResourceInfo: false,
			ExportScopeInfo:    true,
		},
	)

	assert.Equal(t, expected, was)
}
func TestBuildEventFromLogEventWithoutTimestampWithObservedTimestampUseObservedTimestamp(t *testing.T) {
	// When LogRecord doesn't have timestamp set, but it has ObservedTimestamp set,
	// ObservedTimestamp should be used
	lr := testdata.GenerateLogsOneLogRecord()
	ld := lr.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)

	ld.SetTimestamp(pcommon.NewTimestampFromTime(time.Unix(0, 0)))
	ld.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Unix(1686235113, 0)))

	testLEventRaw.Ts = "1686235113000000000"
	// 2023-06-08 14:38:33 +0000 UTC
	testLEventRaw.Attrs["sca:observedTime"] = "1686235113000000000"
	delete(testLEventRaw.Attrs, "timestamp")
	delete(testLEventRaw.Attrs, "resource-attr")

	expected := &add_events.EventBundle{
		Event:  testLEventRaw,
		Thread: testLThread,
		Log:    testLLog,
	}
	was := buildEventFromLog(
		ld,
		lr.ResourceLogs().At(0).Resource(),
		lr.ResourceLogs().At(0).ScopeLogs().At(0).Scope(),
		testServerHost,
		newDefaultLogsSettings(),
	)

	assert.Equal(t, expected, was)
}

func TestBuildEventFromLogEventWithoutTimestampWithOutObservedTimestampUseCurrentTimestamp(t *testing.T) {
	// When LogRecord doesn't have timestamp and ObservedTimestamp set, current timestamp
	// should be used
	// We mock current time to ensure stability across runs

	now = func() time.Time { return time.Unix(123456789, 0) }
	currentTime := now()
	assert.Equal(t, currentTime, time.Unix(123456789, 0))
	assert.Equal(t, strconv.FormatInt(currentTime.UnixNano(), 10), "123456789000000000")

	lr := testdata.GenerateLogsOneLogRecord()
	ld := lr.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)

	ld.SetTimestamp(pcommon.NewTimestampFromTime(time.Unix(0, 0)))
	ld.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Unix(0, 0)))

	testLEventRaw.Ts = strconv.FormatInt(currentTime.UnixNano(), 10)
	delete(testLEventRaw.Attrs, "timestamp")
	delete(testLEventRaw.Attrs, "sca:observedTime")
	delete(testLEventRaw.Attrs, "resource-attr")

	expected := &add_events.EventBundle{
		Event:  testLEventRaw,
		Thread: testLThread,
		Log:    testLLog,
	}
	was := buildEventFromLog(
		ld,
		lr.ResourceLogs().At(0).Resource(),
		lr.ResourceLogs().At(0).ScopeLogs().At(0).Scope(),
		testServerHost,
		newDefaultLogsSettings(),
	)

	assert.Equal(t, expected, was)
}

func extract(req *http.Request) (add_events.AddEventsRequest, error) {
	data, _ := io.ReadAll(req.Body)
	b := bytes.NewBuffer(data)
	reader, _ := gzip.NewReader(b)

	var resB bytes.Buffer
	_, _ = resB.ReadFrom(reader)

	cer := &add_events.AddEventsRequest{}
	err := json.Unmarshal(resB.Bytes(), cer)
	return *cer, err
}

func TestConsumeLogsShouldSucceed(t *testing.T) {
	createSettings := exportertest.NewNopCreateSettings()

	attempt := atomic.Uint64{}
	wasSuccessful := atomic.Bool{}
	addRequest := add_events.AddEventsRequest{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		attempt.Add(1)
		cer, err := extract(req)
		addRequest = cer

		assert.NoError(t, err, "Error reading request: %v", err)

		wasSuccessful.Store(true)
		payload, err := json.Marshal(map[string]interface{}{
			"status":       "success",
			"bytesCharged": 42,
		})
		assert.NoError(t, err)
		l, err := w.Write(payload)
		assert.Greater(t, l, 1)
		assert.NoError(t, err)
	}))
	defer server.Close()

	config := &Config{
		DatasetURL: server.URL,
		APIKey:     "key-lib",
		BufferSettings: BufferSettings{
			MaxLifetime:          500 * time.Millisecond,
			GroupBy:              []string{"attributes.container_id"},
			RetryInitialInterval: time.Second,
			RetryMaxInterval:     time.Minute,
			RetryMaxElapsedTime:  time.Hour,
			RetryShutdownTimeout: time.Minute,
		},
		LogsSettings: LogsSettings{
			ExportResourceInfo:   true,
			ExportResourcePrefix: "R#",
			ExportScopeInfo:      true,
			ExportScopePrefix:    "S#",
			ExportSeparator:      "#",
		},
		TracesSettings: TracesSettings{
			ExportSeparator: "?",
		},
		ServerHostSettings: ServerHostSettings{
			ServerHost: testServerHost,
		},
		RetrySettings:   exporterhelper.NewDefaultRetrySettings(),
		QueueSettings:   exporterhelper.NewDefaultQueueSettings(),
		TimeoutSettings: exporterhelper.NewDefaultTimeoutSettings(),
	}

	lr1 := testdata.GenerateLogsOneLogRecord()
	lr2 := testdata.GenerateLogsOneLogRecord()
	// set attribute for the hostname, it should beat value from resource
	lr2.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).Attributes().PutStr(add_events.AttrServerHost, "serverHostFromAttribute")
	lr2.ResourceLogs().At(0).Resource().Attributes().PutStr(add_events.AttrServerHost, "serverHostFromResource")
	// set attribute serverHost and host.name attributes in the resource, serverHost will win
	lr3 := testdata.GenerateLogsOneLogRecord()
	lr3.ResourceLogs().At(0).Resource().Attributes().PutStr(add_events.AttrServerHost, "serverHostFromResourceServer")
	lr3.ResourceLogs().At(0).Resource().Attributes().PutStr("host.name", "serverHostFromResourceHost")
	// set attribute host.name in the resource attribute
	lr4 := testdata.GenerateLogsOneLogRecord()
	lr4.ResourceLogs().At(0).Resource().Attributes().PutStr("host.name", "serverHostFromResourceHost")
	// set all possible values
	lr5 := testdata.GenerateLogsOneLogRecord()
	fillAttributes(lr5.ResourceLogs().At(0).Resource().Attributes(), true, "R")
	fillAttributes(lr5.ResourceLogs().At(0).ScopeLogs().At(0).Scope().Attributes(), true, "S")
	fillAttributes(lr5.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).Attributes(), true, "A")

	ld := plog.NewLogs()
	ld.ResourceLogs().AppendEmpty()
	ld.ResourceLogs().AppendEmpty()
	ld.ResourceLogs().AppendEmpty()
	ld.ResourceLogs().AppendEmpty()
	ld.ResourceLogs().AppendEmpty()

	lr1.ResourceLogs().At(0).CopyTo(ld.ResourceLogs().At(0))
	lr2.ResourceLogs().At(0).CopyTo(ld.ResourceLogs().At(1))
	lr3.ResourceLogs().At(0).CopyTo(ld.ResourceLogs().At(2))
	lr4.ResourceLogs().At(0).CopyTo(ld.ResourceLogs().At(3))
	lr5.ResourceLogs().At(0).CopyTo(ld.ResourceLogs().At(4))

	logs, err := createLogsExporter(context.Background(), createSettings, config)
	if assert.NoError(t, err) {
		err = logs.Start(context.Background(), componenttest.NewNopHost())
		assert.NoError(t, err)

		assert.NotNil(t, logs)
		err = logs.ConsumeLogs(context.Background(), ld)
		assert.Nil(t, err)
		time.Sleep(time.Second)
		err = logs.Shutdown(context.Background())
		assert.Nil(t, err)
	}

	assert.True(t, wasSuccessful.Load())
	assert.Equal(t,
		add_events.AddEventsRequest{
			AuthParams: request.AuthParams{
				Token: "key-lib",
			},
			AddEventsRequestParams: add_events.AddEventsRequestParams{
				Session:     addRequest.Session,
				SessionInfo: addRequest.SessionInfo,
				Events: []*add_events.Event{
					{
						Thread: testLEventReq.Thread,
						Log:    testLEventReq.Log,
						Sev:    testLEventReq.Sev,
						Ts:     testLEventReq.Ts,
						Attrs: map[string]interface{}{
							add_events.AttrOrigServerHost: testServerHost,
							"app":                         "server",
							"instance_num":                float64(1),
							"dropped_attributes_count":    float64(1),
							"message":                     "This is a log message",
							"span_id":                     "0102040800000000",
							"trace_id":                    "08040201000000000000000000000000",
							"bundle_key":                  "d41d8cd98f00b204e9800998ecf8427e",

							"R#resource-attr": "resource-attr-val-1",
						},
					},
					{
						Thread: testLEventReq.Thread,
						Log:    testLEventReq.Log,
						Sev:    testLEventReq.Sev,
						Ts:     testLEventReq.Ts,
						Attrs: map[string]interface{}{
							add_events.AttrOrigServerHost: "serverHostFromAttribute",
							"app":                         "server",
							"instance_num":                float64(1),
							"dropped_attributes_count":    float64(1),
							"message":                     "This is a log message",
							"span_id":                     "0102040800000000",
							"trace_id":                    "08040201000000000000000000000000",
							"bundle_key":                  "d41d8cd98f00b204e9800998ecf8427e",

							"R#resource-attr": "resource-attr-val-1",
							"R#serverHost":    "serverHostFromResource",
						},
					},
					{
						Thread: testLEventReq.Thread,
						Log:    testLEventReq.Log,
						Sev:    testLEventReq.Sev,
						Ts:     testLEventReq.Ts,
						Attrs: map[string]interface{}{
							add_events.AttrOrigServerHost: "serverHostFromResourceServer",
							"app":                         "server",
							"instance_num":                float64(1),
							"dropped_attributes_count":    float64(1),
							"message":                     "This is a log message",
							"span_id":                     "0102040800000000",
							"trace_id":                    "08040201000000000000000000000000",
							"bundle_key":                  "d41d8cd98f00b204e9800998ecf8427e",

							"R#resource-attr": "resource-attr-val-1",
							"R#host.name":     "serverHostFromResourceHost",
							"R#serverHost":    "serverHostFromResourceServer",
						},
					},
					{
						Thread: testLEventReq.Thread,
						Log:    testLEventReq.Log,
						Sev:    testLEventReq.Sev,
						Ts:     testLEventReq.Ts,
						Attrs: map[string]interface{}{
							add_events.AttrOrigServerHost: "serverHostFromResourceHost",
							"app":                         "server",
							"instance_num":                float64(1),
							"dropped_attributes_count":    float64(1),
							"message":                     "This is a log message",
							"span_id":                     "0102040800000000",
							"trace_id":                    "08040201000000000000000000000000",
							"bundle_key":                  "d41d8cd98f00b204e9800998ecf8427e",

							"R#resource-attr": "resource-attr-val-1",
							"R#host.name":     "serverHostFromResourceHost",
						},
					},
					{
						Thread: testLEventReq.Thread,
						Log:    testLEventReq.Log,
						Sev:    testLEventReq.Sev,
						Ts:     testLEventReq.Ts,
						Attrs: map[string]interface{}{
							add_events.AttrOrigServerHost: testServerHost,
							"app":                         "server",
							"instance_num":                float64(1),
							"dropped_attributes_count":    float64(1),
							"message":                     "This is a log message",
							"span_id":                     "0102040800000000",
							"trace_id":                    "08040201000000000000000000000000",
							"bundle_key":                  "d41d8cd98f00b204e9800998ecf8427e",

							"string":                     "stringA",
							"double":                     2.0,
							"bool":                       true,
							"empty":                      nil,
							"int":                        float64(3),
							"map#map_empty":              nil,
							"map#map_string":             "map_stringA",
							"map#map_map#map_map_string": "map_map_stringA",
							"slice#0":                    "slice_stringA",

							"S#string":                     "stringS",
							"S#double":                     2.0,
							"S#bool":                       true,
							"S#empty":                      nil,
							"S#int":                        float64(3),
							"S#map#map_empty":              nil,
							"S#map#map_string":             "map_stringS",
							"S#map#map_map#map_map_string": "map_map_stringS",
							"S#slice#0":                    "slice_stringS",

							"R#string":                     "stringR",
							"R#double":                     2.0,
							"R#bool":                       true,
							"R#empty":                      nil,
							"R#int":                        float64(3),
							"R#map#map_empty":              nil,
							"R#map#map_string":             "map_stringR",
							"R#map#map_map#map_map_string": "map_map_stringR",
							"R#slice#0":                    "slice_stringR",

							"R#resource-attr": "resource-attr-val-1",
						},
					},
				},
				Threads: []*add_events.Thread{testLThread},
				Logs:    []*add_events.Log{testLLog},
			},
		},
		addRequest,
	)
}

func makeLogRecordWithSeverityNumberAndSeverityText(sevNum int, sevText string) plog.LogRecord {
	lr := testdata.GenerateLogsOneLogRecord()
	ld := lr.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)

	ld.SetSeverityNumber(plog.SeverityNumber(sevNum))
	ld.SetSeverityText(sevText)

	return ld
}

func TestOtelSeverityToDataSetSeverityWithSeverityNumberNoSeverityTextInvalidValues(t *testing.T) {
	// Invalid values get mapped to info (3 - INFO)
	ld := makeLogRecordWithSeverityNumberAndSeverityText(0, "")
	assert.Equal(t, defaultDataSetSeverityLevel, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(-1, "")
	assert.Equal(t, defaultDataSetSeverityLevel, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(25, "")
	assert.Equal(t, defaultDataSetSeverityLevel, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(100, "")
	assert.Equal(t, defaultDataSetSeverityLevel, mapOtelSeverityToDataSetSeverity(ld))

}

func TestOtelSeverityToDataSetSeverityWithSeverityNumberNoSeverityTextDataSetTraceLogLevel(t *testing.T) {
	// trace
	ld := makeLogRecordWithSeverityNumberAndSeverityText(1, "")
	assert.Equal(t, dataSetLogLevelTrace, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(2, "")
	assert.Equal(t, dataSetLogLevelTrace, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(3, "")
	assert.Equal(t, dataSetLogLevelTrace, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(4, "")
	assert.Equal(t, dataSetLogLevelTrace, mapOtelSeverityToDataSetSeverity(ld))
}

func TestOtelSeverityToDataSetSeverityWithSeverityNumberNoSeverityTextDataSetDebugLogLevel(t *testing.T) {
	// debug
	ld := makeLogRecordWithSeverityNumberAndSeverityText(5, "")
	assert.Equal(t, dataSetLogLevelDebug, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(6, "")
	assert.Equal(t, dataSetLogLevelDebug, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(7, "")
	assert.Equal(t, dataSetLogLevelDebug, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(8, "")
	assert.Equal(t, dataSetLogLevelDebug, mapOtelSeverityToDataSetSeverity(ld))
}

func TestOtelSeverityToDataSetSeverityWithSeverityNumberNoSeverityTextDataSetInfoLogLevel(t *testing.T) {
	// info
	ld := makeLogRecordWithSeverityNumberAndSeverityText(9, "")
	assert.Equal(t, dataSetLogLevelInfo, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(10, "")
	assert.Equal(t, dataSetLogLevelInfo, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(11, "")
	assert.Equal(t, dataSetLogLevelInfo, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(12, "")
	assert.Equal(t, dataSetLogLevelInfo, mapOtelSeverityToDataSetSeverity(ld))
}

func TestOtelSeverityToDataSetSeverityWithSeverityNumberNoSeverityTextDataSetWarnLogLevel(t *testing.T) {
	// warn
	ld := makeLogRecordWithSeverityNumberAndSeverityText(13, "")
	assert.Equal(t, dataSetLogLevelWarn, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(14, "")
	assert.Equal(t, dataSetLogLevelWarn, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(15, "")
	assert.Equal(t, dataSetLogLevelWarn, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(16, "")
	assert.Equal(t, dataSetLogLevelWarn, mapOtelSeverityToDataSetSeverity(ld))
}

func TestOtelSeverityToDataSetSeverityWithSeverityNumberNoSeverityTextDataSetErrorLogLevel(t *testing.T) {
	// error
	ld := makeLogRecordWithSeverityNumberAndSeverityText(17, "")
	assert.Equal(t, dataSetLogLevelError, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(18, "")
	assert.Equal(t, dataSetLogLevelError, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(19, "")
	assert.Equal(t, dataSetLogLevelError, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(20, "")
	assert.Equal(t, dataSetLogLevelError, mapOtelSeverityToDataSetSeverity(ld))
}

func TestOtelSeverityToDataSetSeverityWithSeverityNumberNoSeverityTextDataSetFatalLogLevel(t *testing.T) {
	// fatal
	ld := makeLogRecordWithSeverityNumberAndSeverityText(21, "")
	assert.Equal(t, dataSetLogLevelFatal, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(22, "")
	ld.SetSeverityNumber(22)
	assert.Equal(t, dataSetLogLevelFatal, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(23, "")
	ld.SetSeverityNumber(22)
	assert.Equal(t, dataSetLogLevelFatal, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(24, "")
	assert.Equal(t, dataSetLogLevelFatal, mapOtelSeverityToDataSetSeverity(ld))
}

func TestOtelSeverityToDataSetSeverityWithSeverityTextNoSeverityNumberInvalidValues(t *testing.T) {
	// invalid values
	ld := makeLogRecordWithSeverityNumberAndSeverityText(0, "a")
	assert.Equal(t, defaultDataSetSeverityLevel, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(0, "infoinfo")
	assert.Equal(t, defaultDataSetSeverityLevel, mapOtelSeverityToDataSetSeverity(ld))
}

func TestOtelSeverityToDataSetSeverityWithSeverityTextNoSeverityNumberDataSetTraceLogLevel(t *testing.T) {
	// trace
	ld := makeLogRecordWithSeverityNumberAndSeverityText(0, "trace")
	assert.Equal(t, dataSetLogLevelTrace, mapOtelSeverityToDataSetSeverity(ld))
}

func TestOtelSeverityToDataSetSeverityWithSeverityTextNoSeverityNumberDataSetDebugLogLevel(t *testing.T) {
	// debug
	ld := makeLogRecordWithSeverityNumberAndSeverityText(0, "debug")
	assert.Equal(t, dataSetLogLevelDebug, mapOtelSeverityToDataSetSeverity(ld))
}

func TestOtelSeverityToDataSetSeverityWithSeverityTextNoSeverityNumberDataSetInfoLogLevel(t *testing.T) {
	// info
	ld := makeLogRecordWithSeverityNumberAndSeverityText(0, "info")
	assert.Equal(t, dataSetLogLevelInfo, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(0, "informational")
	ld.SetSeverityText("informational")
}

func TestOtelSeverityToDataSetSeverityWithSeverityTextNoSeverityNumberDataSetInfoWarnLevel(t *testing.T) {
	// warn
	ld := makeLogRecordWithSeverityNumberAndSeverityText(0, "warn")
	assert.Equal(t, dataSetLogLevelWarn, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(0, "warning")
	assert.Equal(t, dataSetLogLevelWarn, mapOtelSeverityToDataSetSeverity(ld))
}

func TestOtelSeverityToDataSetSeverityWithSeverityTextNoSeverityNumberDataSetInfoErrorLevel(t *testing.T) {
	// error
	ld := makeLogRecordWithSeverityNumberAndSeverityText(0, "error")
	assert.Equal(t, dataSetLogLevelError, mapOtelSeverityToDataSetSeverity(ld))
}

func TestOtelSeverityToDataSetSeverityWithSeverityTextNoSeverityNumberDataSetInfoFatalLevel(t *testing.T) {
	// fatal
	ld := makeLogRecordWithSeverityNumberAndSeverityText(0, "fatal")
	assert.Equal(t, dataSetLogLevelFatal, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(0, "fatal")
	assert.Equal(t, dataSetLogLevelFatal, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(0, "emergency")
	assert.Equal(t, dataSetLogLevelFatal, mapOtelSeverityToDataSetSeverity(ld))
}

func TestOtelSeverityToDataSetSeverityWithSeverityNumberAndSeverityTextSeverityNumberHasPriority(t *testing.T) {
	// If provided, SeverityNumber has priority over SeverityText
	ld := makeLogRecordWithSeverityNumberAndSeverityText(3, "debug")
	assert.Equal(t, dataSetLogLevelTrace, mapOtelSeverityToDataSetSeverity(ld))

	ld = makeLogRecordWithSeverityNumberAndSeverityText(22, "info")
	assert.Equal(t, dataSetLogLevelFatal, mapOtelSeverityToDataSetSeverity(ld))
}
