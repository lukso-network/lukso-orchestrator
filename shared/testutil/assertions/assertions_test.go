package assertions_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assertions"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	eth "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func Test_Equal(t *testing.T) {
	type args struct {
		tb       *assertions.TBMock
		expected interface{}
		actual   interface{}
		msgs     []interface{}
	}
	tests := []struct {
		name        string
		args        args
		expectedErr string
	}{
		{
			name: "equal values",
			args: args{
				tb:       &assertions.TBMock{},
				expected: 42,
				actual:   42,
			},
		},
		{
			name: "equal values different types",
			args: args{
				tb:       &assertions.TBMock{},
				expected: uint64(42),
				actual:   42,
			},
			expectedErr: "Values are not equal, want: 42 (uint64), got: 42 (int)",
		},
		{
			name: "non-equal values",
			args: args{
				tb:       &assertions.TBMock{},
				expected: 42,
				actual:   41,
			},
			expectedErr: "Values are not equal, want: 42 (int), got: 41 (int)",
		},
		{
			name: "custom error message",
			args: args{
				tb:       &assertions.TBMock{},
				expected: 42,
				actual:   41,
				msgs:     []interface{}{"Custom values are not equal"},
			},
			expectedErr: "Custom values are not equal, want: 42 (int), got: 41 (int)",
		},
		{
			name: "custom error message with params",
			args: args{
				tb:       &assertions.TBMock{},
				expected: 42,
				actual:   41,
				msgs:     []interface{}{"Custom values are not equal (for slot %d)", 12},
			},
			expectedErr: "Custom values are not equal (for slot 12), want: 42 (int), got: 41 (int)",
		},
	}
	for _, tt := range tests {
		verify := func() {
			if tt.expectedErr == "" && tt.args.tb.ErrorfMsg != "" {
				t.Errorf("Unexpected error: %v", tt.args.tb.ErrorfMsg)
			} else if !strings.Contains(tt.args.tb.ErrorfMsg, tt.expectedErr) {
				t.Errorf("got: %q, want: %q", tt.args.tb.ErrorfMsg, tt.expectedErr)
			}
		}
		t.Run(fmt.Sprintf("Assert/%s", tt.name), func(t *testing.T) {
			assert.Equal(tt.args.tb, tt.args.expected, tt.args.actual, tt.args.msgs...)
			verify()
		})
		t.Run(fmt.Sprintf("Require/%s", tt.name), func(t *testing.T) {
			require.Equal(tt.args.tb, tt.args.expected, tt.args.actual, tt.args.msgs...)
			verify()
		})
	}
}

func Test_NotEqual(t *testing.T) {
	type args struct {
		tb       *assertions.TBMock
		expected interface{}
		actual   interface{}
		msgs     []interface{}
	}
	tests := []struct {
		name        string
		args        args
		expectedErr string
	}{
		{
			name: "equal values",
			args: args{
				tb:       &assertions.TBMock{},
				expected: 42,
				actual:   42,
			},
			expectedErr: "Values are equal, both values are equal: 42 (int)",
		},
		{
			name: "equal values different types",
			args: args{
				tb:       &assertions.TBMock{},
				expected: uint64(42),
				actual:   42,
			},
		},
		{
			name: "non-equal values",
			args: args{
				tb:       &assertions.TBMock{},
				expected: 42,
				actual:   41,
			},
		},
		{
			name: "custom error message",
			args: args{
				tb:       &assertions.TBMock{},
				expected: 42,
				actual:   42,
				msgs:     []interface{}{"Custom values are equal"},
			},
			expectedErr: "Custom values are equal, both values are equal",
		},
	}
	for _, tt := range tests {
		verify := func() {
			if tt.expectedErr == "" && tt.args.tb.ErrorfMsg != "" {
				t.Errorf("Unexpected error: %v", tt.args.tb.ErrorfMsg)
			} else if !strings.Contains(tt.args.tb.ErrorfMsg, tt.expectedErr) {
				t.Errorf("got: %q, want: %q", tt.args.tb.ErrorfMsg, tt.expectedErr)
			}
		}
		t.Run(fmt.Sprintf("Assert/%s", tt.name), func(t *testing.T) {
			assert.NotEqual(tt.args.tb, tt.args.expected, tt.args.actual, tt.args.msgs...)
			verify()
		})
		t.Run(fmt.Sprintf("Require/%s", tt.name), func(t *testing.T) {
			require.NotEqual(tt.args.tb, tt.args.expected, tt.args.actual, tt.args.msgs...)
			verify()
		})
	}
}

func TestAssert_DeepEqual(t *testing.T) {
	type args struct {
		tb       *assertions.TBMock
		expected interface{}
		actual   interface{}
		msgs     []interface{}
	}
	tests := []struct {
		name        string
		args        args
		expectedErr string
	}{
		{
			name: "equal values",
			args: args{
				tb:       &assertions.TBMock{},
				expected: struct{ i int }{42},
				actual:   struct{ i int }{42},
			},
		},
		{
			name: "non-equal values",
			args: args{
				tb:       &assertions.TBMock{},
				expected: struct{ i int }{42},
				actual:   struct{ i int }{41},
			},
			expectedErr: "Values are not equal, want: struct { i int }{i:42}, got: struct { i int }{i:41}",
		},
		{
			name: "custom error message",
			args: args{
				tb:       &assertions.TBMock{},
				expected: struct{ i int }{42},
				actual:   struct{ i int }{41},
				msgs:     []interface{}{"Custom values are not equal"},
			},
			expectedErr: "Custom values are not equal, want: struct { i int }{i:42}, got: struct { i int }{i:41}",
		},
		{
			name: "custom error message with params",
			args: args{
				tb:       &assertions.TBMock{},
				expected: struct{ i int }{42},
				actual:   struct{ i int }{41},
				msgs:     []interface{}{"Custom values are not equal (for slot %d)", 12},
			},
			expectedErr: "Custom values are not equal (for slot 12), want: struct { i int }{i:42}, got: struct { i int }{i:41}",
		},
	}
	for _, tt := range tests {
		verify := func() {
			if tt.expectedErr == "" && tt.args.tb.ErrorfMsg != "" {
				t.Errorf("Unexpected error: %v", tt.args.tb.ErrorfMsg)
			} else if !strings.Contains(tt.args.tb.ErrorfMsg, tt.expectedErr) {
				t.Errorf("got: %q, want: %q", tt.args.tb.ErrorfMsg, tt.expectedErr)
			}
		}
		t.Run(fmt.Sprintf("Assert/%s", tt.name), func(t *testing.T) {
			assert.DeepEqual(tt.args.tb, tt.args.expected, tt.args.actual, tt.args.msgs...)
			verify()
		})
		t.Run(fmt.Sprintf("Require/%s", tt.name), func(t *testing.T) {
			require.DeepEqual(tt.args.tb, tt.args.expected, tt.args.actual, tt.args.msgs...)
			verify()
		})
	}
}

func TestAssert_DeepNotEqual(t *testing.T) {
	type args struct {
		tb       *assertions.TBMock
		expected interface{}
		actual   interface{}
		msgs     []interface{}
	}
	tests := []struct {
		name        string
		args        args
		expectedErr string
	}{
		{
			name: "equal values",
			args: args{
				tb:       &assertions.TBMock{},
				expected: struct{ i int }{42},
				actual:   struct{ i int }{42},
			},
			expectedErr: "Values are equal, want: struct { i int }{i:42}, got: struct { i int }{i:42}",
		},
		{
			name: "non-equal values",
			args: args{
				tb:       &assertions.TBMock{},
				expected: struct{ i int }{42},
				actual:   struct{ i int }{41},
			},
		},
		{
			name: "custom error message",
			args: args{
				tb:       &assertions.TBMock{},
				expected: struct{ i int }{42},
				actual:   struct{ i int }{42},
				msgs:     []interface{}{"Custom values are equal"},
			},
			expectedErr: "Custom values are equal, want: struct { i int }{i:42}, got: struct { i int }{i:42}",
		},
		{
			name: "custom error message with params",
			args: args{
				tb:       &assertions.TBMock{},
				expected: struct{ i int }{42},
				actual:   struct{ i int }{42},
				msgs:     []interface{}{"Custom values are equal (for slot %d)", 12},
			},
			expectedErr: "Custom values are equal (for slot 12), want: struct { i int }{i:42}, got: struct { i int }{i:42}",
		},
	}
	for _, tt := range tests {
		verify := func() {
			if tt.expectedErr == "" && tt.args.tb.ErrorfMsg != "" {
				t.Errorf("Unexpected error: %v", tt.args.tb.ErrorfMsg)
			} else if !strings.Contains(tt.args.tb.ErrorfMsg, tt.expectedErr) {
				t.Errorf("got: %q, want: %q", tt.args.tb.ErrorfMsg, tt.expectedErr)
			}
		}
		t.Run(fmt.Sprintf("Assert/%s", tt.name), func(t *testing.T) {
			assert.DeepNotEqual(tt.args.tb, tt.args.expected, tt.args.actual, tt.args.msgs...)
			verify()
		})
		t.Run(fmt.Sprintf("Require/%s", tt.name), func(t *testing.T) {
			require.DeepNotEqual(tt.args.tb, tt.args.expected, tt.args.actual, tt.args.msgs...)
			verify()
		})
	}
}

func TestAssert_NoError(t *testing.T) {
	type args struct {
		tb   *assertions.TBMock
		err  error
		msgs []interface{}
	}
	tests := []struct {
		name        string
		args        args
		expectedErr string
	}{
		{
			name: "nil error",
			args: args{
				tb: &assertions.TBMock{},
			},
		},
		{
			name: "non-nil error",
			args: args{
				tb:  &assertions.TBMock{},
				err: errors.New("failed"),
			},
			expectedErr: "Unexpected error: failed",
		},
		{
			name: "custom non-nil error",
			args: args{
				tb:   &assertions.TBMock{},
				err:  errors.New("failed"),
				msgs: []interface{}{"Custom error message"},
			},
			expectedErr: "Custom error message: failed",
		},
		{
			name: "custom non-nil error with params",
			args: args{
				tb:   &assertions.TBMock{},
				err:  errors.New("failed"),
				msgs: []interface{}{"Custom error message (for slot %d)", 12},
			},
			expectedErr: "Custom error message (for slot 12): failed",
		},
	}
	for _, tt := range tests {
		verify := func() {
			if tt.expectedErr == "" && tt.args.tb.ErrorfMsg != "" {
				t.Errorf("Unexpected error: %v", tt.args.tb.ErrorfMsg)
			} else if !strings.Contains(tt.args.tb.ErrorfMsg, tt.expectedErr) {
				t.Errorf("got: %q, want: %q", tt.args.tb.ErrorfMsg, tt.expectedErr)
			}
		}
		t.Run(fmt.Sprintf("Assert/%s", tt.name), func(t *testing.T) {
			assert.NoError(tt.args.tb, tt.args.err, tt.args.msgs...)
			verify()
		})
		t.Run(fmt.Sprintf("Require/%s", tt.name), func(t *testing.T) {
			require.NoError(tt.args.tb, tt.args.err, tt.args.msgs...)
			verify()
		})
	}
}

func TestAssert_ErrorContains(t *testing.T) {
	type args struct {
		tb   *assertions.TBMock
		want string
		err  error
		msgs []interface{}
	}
	tests := []struct {
		name        string
		args        args
		expectedErr string
	}{
		{
			name: "nil error",
			args: args{
				tb:   &assertions.TBMock{},
				want: "some error",
			},
			expectedErr: "Expected error not returned, got: <nil>, want: some error",
		},
		{
			name: "unexpected error",
			args: args{
				tb:   &assertions.TBMock{},
				want: "another error",
				err:  errors.New("failed"),
			},
			expectedErr: "Expected error not returned, got: failed, want: another error",
		},
		{
			name: "expected error",
			args: args{
				tb:   &assertions.TBMock{},
				want: "failed",
				err:  errors.New("failed"),
			},
			expectedErr: "",
		},
		{
			name: "custom unexpected error",
			args: args{
				tb:   &assertions.TBMock{},
				want: "another error",
				err:  errors.New("failed"),
				msgs: []interface{}{"Something wrong"},
			},
			expectedErr: "Something wrong, got: failed, want: another error",
		},
		{
			name: "expected error",
			args: args{
				tb:   &assertions.TBMock{},
				want: "failed",
				err:  errors.New("failed"),
				msgs: []interface{}{"Something wrong"},
			},
			expectedErr: "",
		},
		{
			name: "custom unexpected error with params",
			args: args{
				tb:   &assertions.TBMock{},
				want: "another error",
				err:  errors.New("failed"),
				msgs: []interface{}{"Something wrong (for slot %d)", 12},
			},
			expectedErr: "Something wrong (for slot 12), got: failed, want: another error",
		},
		{
			name: "expected error with params",
			args: args{
				tb:   &assertions.TBMock{},
				want: "failed",
				err:  errors.New("failed"),
				msgs: []interface{}{"Something wrong (for slot %d)", 12},
			},
			expectedErr: "",
		},
	}
	for _, tt := range tests {
		verify := func() {
			if tt.expectedErr == "" && tt.args.tb.ErrorfMsg != "" {
				t.Errorf("Unexpected error: %v", tt.args.tb.ErrorfMsg)
			} else if !strings.Contains(tt.args.tb.ErrorfMsg, tt.expectedErr) {
				t.Errorf("got: %q, want: %q", tt.args.tb.ErrorfMsg, tt.expectedErr)
			}
		}
		t.Run(fmt.Sprintf("Assert/%s", tt.name), func(t *testing.T) {
			assert.ErrorContains(tt.args.tb, tt.args.want, tt.args.err, tt.args.msgs...)
			verify()
		})
		t.Run(fmt.Sprintf("Require/%s", tt.name), func(t *testing.T) {
			require.ErrorContains(tt.args.tb, tt.args.want, tt.args.err, tt.args.msgs...)
			verify()
		})
	}
}

func Test_NotNil(t *testing.T) {
	type args struct {
		tb   *assertions.TBMock
		obj  interface{}
		msgs []interface{}
	}
	var nilBlock *eth.SignedBeaconBlock = nil
	tests := []struct {
		name        string
		args        args
		expectedErr string
	}{
		{
			name: "nil",
			args: args{
				tb: &assertions.TBMock{},
			},
			expectedErr: "Unexpected nil value",
		},
		{
			name: "nil custom message",
			args: args{
				tb:   &assertions.TBMock{},
				msgs: []interface{}{"This should not be nil"},
			},
			expectedErr: "This should not be nil",
		},
		{
			name: "nil custom message with params",
			args: args{
				tb:   &assertions.TBMock{},
				msgs: []interface{}{"This should not be nil (for slot %d)", 12},
			},
			expectedErr: "This should not be nil (for slot 12)",
		},
		{
			name: "not nil",
			args: args{
				tb:  &assertions.TBMock{},
				obj: "some value",
			},
			expectedErr: "",
		},
		{
			name: "nil value of dynamic type",
			args: args{
				tb:  &assertions.TBMock{},
				obj: nilBlock,
			},
			expectedErr: "Unexpected nil value",
		},
		{
			name: "make sure that assertion works for basic type",
			args: args{
				tb:  &assertions.TBMock{},
				obj: 15,
			},
		},
	}
	for _, tt := range tests {
		verify := func() {
			if tt.expectedErr == "" && tt.args.tb.ErrorfMsg != "" {
				t.Errorf("Unexpected error: %v", tt.args.tb.ErrorfMsg)
			} else if !strings.Contains(tt.args.tb.ErrorfMsg, tt.expectedErr) {
				t.Errorf("got: %q, want: %q", tt.args.tb.ErrorfMsg, tt.expectedErr)
			}
		}
		t.Run(fmt.Sprintf("Assert/%s", tt.name), func(t *testing.T) {
			assert.NotNil(tt.args.tb, tt.args.obj, tt.args.msgs...)
			verify()
		})
		t.Run(fmt.Sprintf("Require/%s", tt.name), func(t *testing.T) {
			require.NotNil(tt.args.tb, tt.args.obj, tt.args.msgs...)
			verify()
		})
	}
}

func Test_LogsContainDoNotContain(t *testing.T) {
	type args struct {
		tb   *assertions.TBMock
		want string
		flag bool
		msgs []interface{}
	}
	tests := []struct {
		name        string
		args        args
		updateLogs  func(log *logrus.Logger)
		expectedErr string
	}{
		{
			name: "should contain not found",
			args: args{
				tb:   &assertions.TBMock{},
				want: "here goes some expected log string",
				flag: true,
			},
			expectedErr: "Expected log not found: here goes some expected log string",
		},
		{
			name: "should contain found",
			args: args{
				tb:   &assertions.TBMock{},
				want: "here goes some expected log string",
				flag: true,
			},
			updateLogs: func(log *logrus.Logger) {
				log.Info("here goes some expected log string")
			},
			expectedErr: "",
		},
		{
			name: "should contain not found custom message",
			args: args{
				tb:   &assertions.TBMock{},
				msgs: []interface{}{"Waited for logs"},
				want: "here goes some expected log string",
				flag: true,
			},
			expectedErr: "Waited for logs: here goes some expected log string",
		},
		{
			name: "should contain not found custom message with params",
			args: args{
				tb:   &assertions.TBMock{},
				msgs: []interface{}{"Waited for %d logs", 10},
				want: "here goes some expected log string",
				flag: true,
			},
			expectedErr: "Waited for 10 logs: here goes some expected log string",
		},
		{
			name: "should not contain and not found",
			args: args{
				tb:   &assertions.TBMock{},
				want: "here goes some unexpected log string",
			},
			expectedErr: "",
		},
		{
			name: "should not contain but found",
			args: args{
				tb:   &assertions.TBMock{},
				want: "here goes some unexpected log string",
			},
			updateLogs: func(log *logrus.Logger) {
				log.Info("here goes some unexpected log string")
			},
			expectedErr: "Unexpected log found: here goes some unexpected log string",
		},
		{
			name: "should not contain but found custom message",
			args: args{
				tb:   &assertions.TBMock{},
				msgs: []interface{}{"Dit not expect logs"},
				want: "here goes some unexpected log string",
			},
			updateLogs: func(log *logrus.Logger) {
				log.Info("here goes some unexpected log string")
			},
			expectedErr: "Dit not expect logs: here goes some unexpected log string",
		},
		{
			name: "should not contain but found custom message with params",
			args: args{
				tb:   &assertions.TBMock{},
				msgs: []interface{}{"Dit not expect %d logs", 10},
				want: "here goes some unexpected log string",
			},
			updateLogs: func(log *logrus.Logger) {
				log.Info("here goes some unexpected log string")
			},
			expectedErr: "Dit not expect 10 logs: here goes some unexpected log string",
		},
	}
	for _, tt := range tests {
		verify := func() {
			if tt.expectedErr == "" && tt.args.tb.ErrorfMsg != "" {
				t.Errorf("Unexpected error: %v", tt.args.tb.ErrorfMsg)
			} else if !strings.Contains(tt.args.tb.ErrorfMsg, tt.expectedErr) {
				t.Errorf("got: %q, want: %q", tt.args.tb.ErrorfMsg, tt.expectedErr)
			}
		}
		t.Run(fmt.Sprintf("Assert/%s", tt.name), func(t *testing.T) {
			log, hook := test.NewNullLogger()
			if tt.updateLogs != nil {
				tt.updateLogs(log)
			}
			if tt.args.flag {
				assert.LogsContain(tt.args.tb, hook, tt.args.want, tt.args.msgs...)
			} else {
				assert.LogsDoNotContain(tt.args.tb, hook, tt.args.want, tt.args.msgs...)
			}
			verify()
		})
		t.Run(fmt.Sprintf("Require/%s", tt.name), func(t *testing.T) {
			log, hook := test.NewNullLogger()
			if tt.updateLogs != nil {
				tt.updateLogs(log)
			}
			if tt.args.flag {
				require.LogsContain(tt.args.tb, hook, tt.args.want, tt.args.msgs...)
			} else {
				require.LogsDoNotContain(tt.args.tb, hook, tt.args.want, tt.args.msgs...)
			}
			verify()
		})
	}
}
