package main

import (
	"math"
)

func GeneratePredictions(history []DiskRecord, futurePoints int, intervalSeconds int64) []DiskRecord {
	if len(history) < 2 || futurePoints <= 0 {
		return nil
	}
	latest := history[0]
	previous := history[1]
	timeDiff := latest.Timestamp - previous.Timestamp
	if timeDiff <= 0 {
		return nil
	}
	usedDiff := latest.UsedSize - previous.UsedSize
	m := float64(usedDiff) / float64(timeDiff)
	latestTime := latest.Timestamp
	totalSize := latest.TotalSize
	diskPath := latest.DiskPath
	uuid := latest.UUID
	var predictions []DiskRecord
	for i := 1; i <= futurePoints; i++ {
		nextTime := latestTime + int64(i)*intervalSeconds
		timeSinceLatest := nextTime - latestTime
		predictedUsedSize := float64(latest.UsedSize) + (m * float64(timeSinceLatest))
		roundedUsedSize := math.Round(predictedUsedSize)
		if roundedUsedSize >= float64(totalSize) || roundedUsedSize < 0 {
			break
		}
		usedSize := int64(roundedUsedSize)
		availableSize := int64(totalSize) - usedSize
		usedPercentage := (float64(usedSize) / float64(totalSize)) * 100
		prediction := DiskRecord{
			Timestamp:      nextTime,
			DiskPath:       diskPath,
			UUID:           uuid,
			TotalSize:      totalSize,
			UsedSize:       float64(usedSize),
			AvailableSize:  float64(availableSize),
			UsedPercentage: usedPercentage,
			Prediction:     1,
		}
		predictions = append(predictions, prediction)
	}
	return predictions
}