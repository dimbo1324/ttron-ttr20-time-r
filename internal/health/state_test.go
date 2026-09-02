package health

import "testing"

func TestDefaultPolicyIsValid(t *testing.T) {
	policy := DefaultPolicy()
	if err := policy.Validate(); err != nil {
		t.Fatalf("DefaultPolicy() must be valid: %v", err)
	}
	if policy.DegradeAfter != DefaultDegradeAfter || policy.OfflineAfter != DefaultOfflineAfter {
		t.Fatalf("DefaultPolicy() = %+v", policy)
	}
}

func TestPolicyNormalize(t *testing.T) {
	tests := []struct {
		name string
		in   Policy
		want Policy
	}{
		{
			name: "zero falls back to defaults",
			in:   Policy{},
			want: DefaultPolicy(),
		},
		{
			name: "negative falls back to defaults",
			in:   Policy{DegradeAfter: -1, OfflineAfter: -2, RecoverAfter: -3, WindowSize: -4},
			want: DefaultPolicy(),
		},
		{
			name: "offline below degrade is raised",
			in:   Policy{DegradeAfter: 5, OfflineAfter: 2, RecoverAfter: 1, WindowSize: 8},
			want: Policy{DegradeAfter: 5, OfflineAfter: 5, RecoverAfter: 1, WindowSize: 8},
		},
		{
			name: "valid policy is preserved",
			in:   Policy{DegradeAfter: 2, OfflineAfter: 4, RecoverAfter: 1, WindowSize: 16},
			want: Policy{DegradeAfter: 2, OfflineAfter: 4, RecoverAfter: 1, WindowSize: 16},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.Normalize(); got != tt.want {
				t.Fatalf("Normalize() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestPolicyValidate(t *testing.T) {
	tests := []struct {
		name    string
		in      Policy
		wantErr bool
	}{
		{name: "valid", in: Policy{DegradeAfter: 2, OfflineAfter: 4, RecoverAfter: 1, WindowSize: 8}},
		{name: "zero degrade", in: Policy{OfflineAfter: 4, RecoverAfter: 1, WindowSize: 8}, wantErr: true},
		{name: "zero offline", in: Policy{DegradeAfter: 2, RecoverAfter: 1, WindowSize: 8}, wantErr: true},
		{name: "zero recover", in: Policy{DegradeAfter: 2, OfflineAfter: 4, WindowSize: 8}, wantErr: true},
		{name: "offline below degrade", in: Policy{DegradeAfter: 4, OfflineAfter: 2, RecoverAfter: 1, WindowSize: 8}, wantErr: true},
		{name: "zero window", in: Policy{DegradeAfter: 2, OfflineAfter: 4, RecoverAfter: 1}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.in.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %t", err, tt.wantErr)
			}
		})
	}
}

func TestStatePredicates(t *testing.T) {
	tests := []struct {
		state         State
		wantHealthy   bool
		wantReachable bool
	}{
		{state: StateUnknown},
		{state: StateOnline, wantHealthy: true, wantReachable: true},
		{state: StateDegraded, wantReachable: true},
		{state: StateOffline},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			if got := tt.state.Healthy(); got != tt.wantHealthy {
				t.Fatalf("Healthy() = %t, want %t", got, tt.wantHealthy)
			}
			if got := tt.state.Reachable(); got != tt.wantReachable {
				t.Fatalf("Reachable() = %t, want %t", got, tt.wantReachable)
			}
		})
	}
}
