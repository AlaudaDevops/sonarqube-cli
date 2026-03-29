package config

import (
	"testing"
)

func TestReplaceVariables(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		taskRunID  string
		pluginName string
		want       string
	}{
		{
			name:       "replace {{TASK_RUN_ID}}",
			input:      "test-{{TASK_RUN_ID}}",
			taskRunID:  "abc123",
			pluginName: "tektoncd",
			want:       "test-abc123",
		},
		{
			name:       "replace ${TASK_RUN_ID}",
			input:      "test-${TASK_RUN_ID}",
			taskRunID:  "abc123",
			pluginName: "tektoncd",
			want:       "test-abc123",
		},
		{
			name:       "replace {{PLUGIN_NAME}}",
			input:      "{{PLUGIN_NAME}}-project",
			taskRunID:  "abc123",
			pluginName: "tektoncd",
			want:       "tektoncd-project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReplaceVariables(tt.input, tt.taskRunID, tt.pluginName)
			if err != nil {
				t.Errorf("ReplaceVariables() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("ReplaceVariables() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("error if taskRunID empty", func(t *testing.T) {
		_, err := ReplaceVariables("test-{{TASK_RUN_ID}}", "", "tektoncd")
		if err == nil {
			t.Error("ReplaceVariables() expected error for empty taskRunID, got nil")
		}
	})
}

func TestApplyTemplate(t *testing.T) {
	t.Run("fail for empty taskRunID", func(t *testing.T) {
		res := &TempResource{
			Group: Group{Name: "test-{{TASK_RUN_ID}}"},
		}
		err := ApplyTemplate(res, "", "tekton")
		if err == nil {
			t.Error("ApplyTemplate() expected error for empty taskRunID, got nil")
		}
	})

	t.Run("success for valid variables", func(t *testing.T) {
		res := &TempResource{
			Group: Group{Name: "test-{{TASK_RUN_ID}}"},
		}
		err := ApplyTemplate(res, "abc", "tekton")
		if err != nil {
			t.Errorf("ApplyTemplate() unexpected error = %v", err)
		}
		if res.Group.Name != "test-abc" {
			t.Errorf("ApplyTemplate() name = %v, want test-abc", res.Group.Name)
		}
	})

	t.Run("ignore empty variables if no placeholder", func(t *testing.T) {
		res := &TempResource{
			Group: Group{Name: "fixed-name"},
		}
		err := ApplyTemplate(res, "", "")
		if err != nil {
			t.Errorf("ApplyTemplate() unexpected error = %v", err)
		}
		if res.Group.Name != "fixed-name" {
			t.Errorf("ApplyTemplate() name = %v, want fixed-name", res.Group.Name)
		}
	})
}
