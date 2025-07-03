package btc

import (
	"testing"

	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestBTCMaxRangeCalculation(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		syncConfirmed  int64
		confirmed      int64
		maxRange       int
		expectedEnd    int64
		expectedBlocks int64
	}{
		{
			name:           "Normal range, not exceeding max limit",
			syncConfirmed:  1000,
			confirmed:      1100,
			maxRange:       100,
			expectedEnd:    1100,
			expectedBlocks: 100,
		},
		{
			name:           "Exceeds max range limit",
			syncConfirmed:  1000,
			confirmed:      1200,
			maxRange:       100,
			expectedEnd:    1100,
			expectedBlocks: 100,
		},
		{
			name:           "Small range sync",
			syncConfirmed:  1000,
			confirmed:      1010,
			maxRange:       100,
			expectedEnd:    1010,
			expectedBlocks: 10,
		},
		{
			name:           "Zero range",
			syncConfirmed:  1000,
			confirmed:      1000,
			maxRange:       100,
			expectedEnd:    1000,
			expectedBlocks: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate configuration
			originalMaxRange := config.AppConfig.BTCMaxRange
			config.AppConfig.BTCMaxRange = tc.maxRange
			defer func() { config.AppConfig.BTCMaxRange = originalMaxRange }()

			// Calculate end height
			maxRange := int64(config.AppConfig.BTCMaxRange)
			endHeight := tc.syncConfirmed + maxRange
			if endHeight > tc.confirmed {
				endHeight = tc.confirmed
			}

			// Verify results
			assert.Equal(t, tc.expectedEnd, endHeight)

			// Calculate actual number of blocks to sync
			actualBlocks := endHeight - tc.syncConfirmed
			assert.Equal(t, tc.expectedBlocks, actualBlocks)
		})
	}
}
