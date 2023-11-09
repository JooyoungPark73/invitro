package generator

import (
	"github.com/vhive-serverless/loader/pkg/common"
	"math"
	"testing"
)

func TestWarmStartMatrix(t *testing.T) {
	tests := []struct {
		testName           string
		experimentDuration int
		rpsTarget          float64
		expectedIAT        common.IATArray
		expectedCount      []int
	}{
		{
			testName:           "2min_1rps",
			experimentDuration: 2,
			rpsTarget:          1,
			expectedIAT: []float64{
				// minute 1
				0, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000,
				1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000,
				1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000,
				1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000,
				1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000,
				1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000,
				// minute 2
				0, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000,
				1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000,
				1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000,
				1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000,
				1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000,
				1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 1_000_000, 0,
			},
			expectedCount: []int{61, 61},
		},
		{
			testName:           "2min_0.5rps",
			experimentDuration: 2,
			rpsTarget:          0.5,
			expectedIAT: []float64{
				// minute 1
				0, 2_000_000, 2_000_000, 2_000_000, 2_000_000, 2_000_000,
				2_000_000, 2_000_000, 2_000_000, 2_000_000, 2_000_000,
				2_000_000, 2_000_000, 2_000_000, 2_000_000, 2_000_000,
				2_000_000, 2_000_000, 2_000_000, 2_000_000, 2_000_000,
				2_000_000, 2_000_000, 2_000_000, 2_000_000, 2_000_000,
				2_000_000, 2_000_000, 2_000_000, 2_000_000, 2_000_000,
				// minute 2
				0, 2_000_000, 2_000_000, 2_000_000, 2_000_000, 2_000_000,
				2_000_000, 2_000_000, 2_000_000, 2_000_000, 2_000_000,
				2_000_000, 2_000_000, 2_000_000, 2_000_000, 2_000_000,
				2_000_000, 2_000_000, 2_000_000, 2_000_000, 2_000_000,
				2_000_000, 2_000_000, 2_000_000, 2_000_000, 2_000_000,
				2_000_000, 2_000_000, 2_000_000, 2_000_000, 2_000_000, 0,
			},
			expectedCount: []int{31, 31},
		},
		{
			testName:           "2min_0.125rps",
			experimentDuration: 2,
			rpsTarget:          0.125,
			expectedIAT: []float64{
				// minute 1
				0, 8_000_000, 8_000_000, 8_000_000, 8_000_000, 8_000_000, 8_000_000, 8_000_000, 8_000_000,
				// minute 2
				0, 8_000_000, 8_000_000, 8_000_000, 8_000_000, 8_000_000, 8_000_000, 8_000_000, 0,
			},
			expectedCount: []int{9, 8},
		},
	}

	epsilon := 0.01

	for _, test := range tests {
		t.Run("warm_start"+test.testName, func(t *testing.T) {
			matrix, minuteCount := GenerateWarmStartFunction(test.experimentDuration, test.rpsTarget)

			if len(matrix) != len(test.expectedIAT) {
				t.Errorf("Unexpected IAT array size - got: %d, expected: %d", len(matrix), len(test.expectedIAT))
			}
			if len(minuteCount) != len(test.expectedCount) {
				t.Errorf("Unexpected count array size - got: %d, expected: %d", len(minuteCount), len(test.expectedCount))
			}

			sum := 0.0
			count := 0
			currentMinute := 0

			for i := 0; i < len(matrix); i++ {
				if math.Abs(matrix[i]-test.expectedIAT[i]) > epsilon {
					t.Error("Unexpected IAT value.")
				}

				sum += matrix[i]
				count++

				if int(sum/60_000_000) != currentMinute {
					if count != test.expectedCount[currentMinute] {
						t.Error("Unexpected count array value.")
					}

					currentMinute = int(sum / 60_000_000)
					count = 0
				}
			}
		})
	}
}

func TestColdStartMatrix(t *testing.T) {
	tests := []struct {
		testName           string
		experimentDuration int
		rpsTarget          float64
		cooldownSeconds    int
		expectedIAT        []common.IATArray
		expectedCount      [][]int
	}{
		{
			testName:           "2min_1rps",
			experimentDuration: 2,
			rpsTarget:          1,
			cooldownSeconds:    10,
			expectedIAT: []common.IATArray{
				{0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-1_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-2_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-3_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-4_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-5_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-6_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-7_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-8_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-9_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
			},
			expectedCount: [][]int{
				{7, 7},
				{8, 7},
				{8, 7},
				{8, 7},
				{8, 7},
				{8, 7},
				{8, 7},
				{8, 7},
				{8, 7},
				{8, 7},
			},
		},
		{
			testName:           "1min_0.25rps",
			experimentDuration: 1,
			rpsTarget:          0.25,
			cooldownSeconds:    10,
			expectedIAT: []common.IATArray{
				{0, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 0},
				{-4_000_000, 0, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 0},
				{-8_000_000, 0, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 0},
			},
			expectedCount: [][]int{
				{6},
				{7},
				{7},
			},
		},
		{
			testName:           "2min_0.25rps",
			experimentDuration: 2,
			rpsTarget:          0.25,
			cooldownSeconds:    10,
			expectedIAT: []common.IATArray{
				{0, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 0, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 0},
				{-4_000_000, 0, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 0, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 0},
				{-8_000_000, 0, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 0, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 0},
			},
			expectedCount: [][]int{
				{6, 6},
				{7, 6},
				{7, 6},
			},
		},
		{
			testName:           "1min_0.33rps",
			experimentDuration: 1,
			rpsTarget:          1.0 / 3,
			cooldownSeconds:    10,
			expectedIAT: []common.IATArray{
				{0, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 0},
				{-3_000_000, 0, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 0},
				{-6_000_000, 0, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 0},
				{-9_000_000, 0, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 12_000_000, 0},
			},
			expectedCount: [][]int{
				{6},
				{7},
				{7},
				{7},
			},
		},
		{
			testName:           "1min_5rps",
			experimentDuration: 1,
			rpsTarget:          5,
			cooldownSeconds:    10,
			expectedIAT: []common.IATArray{
				{0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-200_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-400_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-600_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-800_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},

				{-1_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-1_200_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-1_400_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-1_600_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-1_800_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},

				{-2_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-2_200_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-2_400_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-2_600_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-2_800_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},

				{-3_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-3_200_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-3_400_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-3_600_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-3_800_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},

				{-4_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-4_200_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-4_400_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-4_600_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-4_800_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},

				{-5_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-5_200_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-5_400_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-5_600_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-5_800_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},

				{-6_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-6_200_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-6_400_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-6_600_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-6_800_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},

				{-7_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-7_200_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-7_400_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-7_600_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-7_800_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},

				{-8_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-8_200_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-8_400_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-8_600_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-8_800_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},

				{-9_000_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-9_200_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-9_400_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-9_600_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
				{-9_800_000, 0, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 10_000_000, 0},
			},
			expectedCount: [][]int{
				{7},
				{8},
				{8},
				{8},
				{8},

				{8},
				{8},
				{8},
				{8},
				{8},

				{8},
				{8},
				{8},
				{8},
				{8},

				{8},
				{8},
				{8},
				{8},
				{8},

				{8},
				{8},
				{8},
				{8},
				{8},

				{8},
				{8},
				{8},
				{8},
				{8},

				{8},
				{8},
				{8},
				{8},
				{8},

				{8},
				{8},
				{8},
				{8},
				{8},

				{8},
				{8},
				{8},
				{8},
				{8},

				{8},
				{8},
				{8},
				{8},
				{8},
			},
		},
		{
			testName:           "1min_5rps_cooldown5s",
			experimentDuration: 1,
			rpsTarget:          5,
			cooldownSeconds:    5,
			expectedIAT: []common.IATArray{
				{0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},
				{-200_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},
				{-400_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},
				{-600_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},
				{-800_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},

				{-1_000_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},
				{-1_200_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},
				{-1_400_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},
				{-1_600_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},
				{-1_800_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},

				{-2_000_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},
				{-2_200_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},
				{-2_400_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},
				{-2_600_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},
				{-2_800_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},

				{-3_000_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},
				{-3_200_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},
				{-3_400_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},
				{-3_600_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},
				{-3_800_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},

				{-4_000_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},
				{-4_200_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},
				{-4_400_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},
				{-4_600_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},
				{-4_800_000, 0, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 5_000_000, 0},
			},
			expectedCount: [][]int{
				{13},
				{14},
				{14},
				{14},
				{14},

				{14},
				{14},
				{14},
				{14},
				{14},

				{14},
				{14},
				{14},
				{14},
				{14},

				{14},
				{14},
				{14},
				{14},
				{14},

				{14},
				{14},
				{14},
				{14},
				{14},
			},
		},
	}

	epsilon := 0.01

	for _, test := range tests {
		t.Run("cold_start_"+test.testName, func(t *testing.T) {
			matrix, minuteCounts := GenerateColdStartFunctions(test.experimentDuration, test.rpsTarget, test.cooldownSeconds)

			if len(matrix) != len(test.expectedIAT) {
				t.Errorf("Unexpected number of functions - got: %d, expected: %d", len(matrix), len(test.expectedIAT))
			}
			if len(minuteCounts) != len(test.expectedCount) {
				t.Errorf("Unexpected count array size - got: %d, expected: %d", len(minuteCounts), len(test.expectedCount))
			}

			for fIndex := 0; fIndex < len(matrix); fIndex++ {
				sum := 0.0
				count := 0
				currentMinute := 0

				if len(matrix[fIndex]) != len(test.expectedIAT[fIndex]) {
					t.Errorf("Unexpected length of function %d IAT array - got: %d, expected: %d", fIndex, len(matrix[fIndex]), len(test.expectedIAT[fIndex]))
				}

				for i := 0; i < len(matrix[fIndex]); i++ {
					if math.Abs(matrix[fIndex][i]-test.expectedIAT[fIndex][i]) > epsilon {
						t.Errorf("Unexpected value fx %d val %d - got: %f; expected: %f", fIndex, i, matrix[fIndex][i], test.expectedIAT[fIndex][i])
					}

					if currentMinute > len(test.expectedCount[fIndex]) {
						t.Errorf("Invalid expected count array size for function with index %d", fIndex)
					}

					if matrix[fIndex][i] >= 0 {
						sum += matrix[fIndex][i]
					}
					count++

					if int(sum/60_000_000) != currentMinute {
						if count != test.expectedCount[fIndex][currentMinute] {
							t.Errorf("Unexpected count array value fx %d; min %d - got: %d; expected: %d", fIndex, currentMinute, count, test.expectedCount[fIndex][currentMinute])
						}

						currentMinute = int(sum / 60_000_000)
						count = 0
					}
				}
			}
		})
	}
}
