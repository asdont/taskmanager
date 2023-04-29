package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateTaskCreateSetValues(t *testing.T) {
	cases := []struct {
		name        string
		requestBody updateTaskBody
		expected    []string
	}{
		{
			name: "test1",
			requestBody: updateTaskBody{
				Title:     "task1",
				Completed: false,
			},
			expected: []string{"status=false", "title='task1'", "updated=now()", "completed=null"},
		},
		{
			name: "test2",
			requestBody: updateTaskBody{
				Title:     "task2",
				Completed: true,
			},
			expected: []string{"status=true", "title='task2'", "completed=now()"},
		},
		{
			name: "test3",
			requestBody: updateTaskBody{
				Title: "task3",
			},
			expected: []string{"status=false", "title='task3'", "updated=now()", "completed=null"},
		},
		{
			name: "test4",
			requestBody: updateTaskBody{
				Completed: false,
			},
			expected: []string{"status=false", "updated=now()", "completed=null"},
		},
		{
			name: "test5",
			requestBody: updateTaskBody{
				Completed: true,
			},
			expected: []string{"status=true", "completed=now()"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			res := updateTaskCreateSetValues(tt.requestBody)
			assert.Equal(t, tt.expected, res)
		})
	}
}
