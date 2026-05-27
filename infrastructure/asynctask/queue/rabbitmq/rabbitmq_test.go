package rabbitmq

import "testing"

func TestNormalizeAMQPPriority(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want uint8
	}{
		{name: "negative", in: -1, want: 0},
		{name: "zero", in: 0, want: 0},
		{name: "middle", in: 6, want: 6},
		{name: "max", in: 9, want: 9},
		{name: "overflow", in: 100, want: 9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeAMQPPriority(tt.in); got != tt.want {
				t.Fatalf("normalizeAMQPPriority(%d) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
