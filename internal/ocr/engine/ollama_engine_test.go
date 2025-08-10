package engine

// test extractJSON function
import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestExtractJSON(t *testing.T) {
	// arrange
	testCases := []struct {
		input    string
		expected json.RawMessage
	}{
		{
			input:    `{"key": "value"}`,
			expected: json.RawMessage(`{ "key" : "value" }`),
		},
		{
			input:    `Here is some text before the JSON {"key": "value"} and some after.`,
			expected: json.RawMessage(`{"key": "value"}`),
		},
		{
			input:    `No JSON here`,
			expected: nil,
		},
		{
			input:    `"Here is the extracted information in the requested format:\n\n{ \n  \"Name\": \"Sandra\", \n  \"Email\": \"de@@gmail.com\", \n  \"Phone\": \"+41799123123\", \n  \"Tags\": \"Age: 54, Birthday: May 14th, 1971\"\n}"`,
			expected: json.RawMessage(`{"Name":"Sandra","Email":"de@@gmail.com","Phone":"+41799123123","Tags":"Age: 54, Birthday: May 14th, 1971"}`),
		},
		{
			input:    "Blah blah blah. The user's tags are \"player_email_or_not\" and \"player_email_or_not\". \n\nHere is the answer in the requested format:\n\n{\n\t\"Name\": \"Sandra\",\n\t\"Email\": \"de@gmail.com\",\n\t\"Phone\": \"+41799123123\",\n\t\"Tags\": [\"player_email_or_not\", \"player_email_or_not\"]\n}",
			expected: json.RawMessage(`{"Name":"Sandra","Email":"de@gmail.com","Phone":"+41799123123","Tags":["player_email_or_not","player_email_or_not"]}`),
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("input=%q", tc.input), func(t *testing.T) {
			fmt.Printf("Testing input: %s\n", tc.input)

			// act
			actual, _ := extractJSON(tc.input)

			if tc.expected == nil {
				if actual != nil {
					t.Errorf("expected nil, got %q", actual)
				}
				return
			}

			// assert
			var expectedMap, resultMap map[string]any

			if err := json.Unmarshal(tc.expected, &expectedMap); err != nil {
				t.Fatalf("Failed to unmarshal expected JSON: %v", err)
			}

			if err := json.Unmarshal(actual, &resultMap); err != nil {
				t.Fatalf("Failed to unmarshal result JSON: %v", err)
			}

			if !reflect.DeepEqual(expectedMap, resultMap) {
				t.Errorf("JSON objects don't match.\nExpected: %v\nGot: %v", expectedMap, resultMap)
			}
		})
	}
}
