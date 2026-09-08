package models_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/IsaqueB/notification-server/internal/models"
	"github.com/stretchr/testify/require"
)

func Test_TopicUnmarshalJSON(t *testing.T) {
	testCases := []struct {
		purpose     string
		entry       []byte
		expected    models.Topic
		expectedErr error
	}{
		{
			purpose:     "Valid ALL topic",
			entry:       []byte(`"all"`),
			expected:    models.ALL,
			expectedErr: nil,
		},
		{
			purpose:     "Valid PRIVATE topic",
			entry:       []byte(`"private"`),
			expected:    models.PRIVATE,
			expectedErr: nil,
		},
		{
			purpose:     "Valid INVOICE_ISSUED topic",
			entry:       []byte(`"invoice_issued"`),
			expected:    models.INVOICE_ISSUED,
			expectedErr: nil,
		},
		{
			purpose:     "Valid EMPTY topic",
			entry:       []byte(`""`),
			expected:    models.EMPTY,
			expectedErr: models.ErrInvalidTopic, // Even though there is an empty topic registered. It is not a valid topic
		},
		{
			purpose:     "Invalid topic",
			entry:       []byte(`"invalid_topic"`),
			expected:    models.EMPTY,
			expectedErr: models.ErrInvalidTopic,
		},
		{
			purpose:     "Invalid JSON - number instead of string",
			entry:       []byte(`123`),
			expected:    models.EMPTY,
			expectedErr: fmt.Errorf("json: cannot unmarshal number into Go value of type string"),
		},
		{
			purpose:     "Invalid JSON - object instead of string",
			entry:       []byte(`{"topic":"all"}`),
			expected:    models.EMPTY,
			expectedErr: fmt.Errorf("json: cannot unmarshal object into Go value of type string"),
		},
		{
			purpose:     "Invalid JSON - null value",
			entry:       []byte(`null`),
			expected:    models.EMPTY,
			expectedErr: models.ErrInvalidTopic,
		},
		{
			purpose:     "Invalid JSON - empty bytes",
			entry:       []byte(``),
			expected:    models.EMPTY,
			expectedErr: fmt.Errorf("unexpected end of JSON input"),
		},
		{
			purpose:     "Valid topic with spaces",
			entry:       []byte(`" all "`),
			expected:    models.EMPTY,
			expectedErr: models.ErrInvalidTopic,
		},
		{
			purpose:     "Valid topic with uppercase",
			entry:       []byte(`"ALL"`),
			expected:    models.EMPTY,
			expectedErr: models.ErrInvalidTopic,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.purpose, func(t *testing.T) {
			var actual models.Topic

			err := json.Unmarshal(tt.entry, &actual)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, actual)
			}
		})
	}
}

func Test_TopicMarshalJSON(t *testing.T) {
	testCases := []struct {
		purpose     string
		entry       models.Topic
		expected    []byte
		expectedErr error
	}{
		{
			purpose:     "Marshal ALL topic",
			entry:       models.ALL,
			expected:    []byte(`"all"`),
			expectedErr: nil,
		},
		{
			purpose:     "Marshal PRIVATE topic",
			entry:       models.PRIVATE,
			expected:    []byte(`"private"`),
			expectedErr: nil,
		},
		{
			purpose:     "Marshal INVOICE_ISSUED topic",
			entry:       models.INVOICE_ISSUED,
			expected:    []byte(`"invoice_issued"`),
			expectedErr: nil,
		},
		{
			purpose:     "Marshal EMPTY topic",
			entry:       models.EMPTY,
			expected:    []byte(`""`),
			expectedErr: nil,
		},
		{
			purpose:     "Marshal invalid topic (not in list)",
			entry:       models.Topic("invalid"),
			expected:    []byte(`"invalid"`),
			expectedErr: nil, // Marshal não valida
		},
	}

	for _, tt := range testCases {
		t.Run(tt.purpose, func(t *testing.T) {
			actual, err := json.Marshal(tt.entry)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, string(tt.expected), string(actual))
			}
		})
	}
}

func Test_TopicRoundTrip(t *testing.T) {
	testCases := []struct {
		purpose string
		topic   models.Topic
	}{
		{
			purpose: "Round trip ALL",
			topic:   models.ALL,
		},
		{
			purpose: "Round trip PRIVATE",
			topic:   models.PRIVATE,
		},
		{
			purpose: "Round trip INVOICE_ISSUED",
			topic:   models.INVOICE_ISSUED,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.purpose, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.topic)
			require.NoError(t, err)

			// Unmarshal
			var actual models.Topic
			err = json.Unmarshal(data, &actual)
			require.NoError(t, err)

			// Verificar
			require.Equal(t, tt.topic, actual)
		})
	}
}

func Test_TopicInStruct(t *testing.T) {
	type TestStruct struct {
		Topic models.Topic `json:"topic"`
		Title string       `json:"title"`
	}

	testCases := []struct {
		purpose     string
		json        []byte
		expected    TestStruct
		expectedErr error
	}{
		{
			purpose: "Valid struct with topic",
			json:    []byte(`{"topic":"all","title":"Test"}`),
			expected: TestStruct{
				Topic: models.ALL,
				Title: "Test",
			},
			expectedErr: nil,
		},
		{
			purpose:     "Invalid struct with invalid topic",
			json:        []byte(`{"topic":"invalid","title":"Test"}`),
			expected:    TestStruct{},
			expectedErr: models.ErrInvalidTopic,
		},
		{
			purpose: "Struct without topic",
			json:    []byte(`{"title":"Test"}`),
			expected: TestStruct{
				Topic: models.EMPTY,
				Title: "Test",
			},
			expectedErr: nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.purpose, func(t *testing.T) {
			var actual TestStruct

			err := json.Unmarshal(tt.json, &actual)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, actual)
			}
		})
	}
}

func Test_TopicGetAllNotificationTopics(t *testing.T) {
	topics := models.GetAllNotificationTopicsWithEmpty()

	require.NotEmpty(t, topics)
	require.Contains(t, topics, models.ALL)
	require.Contains(t, topics, models.PRIVATE)
	require.Contains(t, topics, models.INVOICE_ISSUED)
	require.Contains(t, topics, models.EMPTY)

	// Verificar se não há duplicatas
	seen := make(map[models.Topic]bool)
	for _, topic := range topics {
		require.False(t, seen[topic], "Duplicate topic found:", topic)
		seen[topic] = true
	}
}
