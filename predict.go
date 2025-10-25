package main

import (
	"math"
)


type PredictionAlgorithm func([]DiskRecord, int, int64) []DiskRecord
func GeneratePredictions(history []DiskRecord, futurePoints int, intervalSeconds int64) []DiskRecord {
    var predictor PredictionAlgorithm
    if len(history) < 2 {
        predictor = nil
    } else {
        predictor = CrudeLinearExtrapolation
    }
    var newPredictions []DiskRecord
    if predictor != nil {
        newPredictions = predictor(history, futurePoints, intervalSeconds)
    }
	return newPredictions
}

func CrudeLinearExtrapolation(history []DiskRecord, futurePoints int, intervalSeconds int64) []DiskRecord {
    if len(history) < 2 || futurePoints <= 0 {
        return nil
    }
    latest := history[0]
    previous := history[1]
    timeDiff := latest.Timestamp - previous.Timestamp
    usedDiff := latest.UsedSize - previous.UsedSize
    
    if timeDiff <= 0 {
        return nil
    }
    m := usedDiff / float64(timeDiff) 
    latestTime := latest.Timestamp
    totalSize := latest.TotalSize
    diskPath := latest.DiskPath
    uuid := latest.UUID
    var predictions []DiskRecord

    for i := 1; i <= futurePoints; i++ {
        nextTime := latestTime + int64(i)*intervalSeconds
        timeSinceLatest := nextTime - latestTime
        predictedUsedSize := latest.UsedSize + (m * float64(timeSinceLatest))
        roundedUsedSize := math.Round(predictedUsedSize)
        if roundedUsedSize >= totalSize || roundedUsedSize < 0 {
            break
        }

        usedSize := roundedUsedSize
        availableSize := totalSize - usedSize
        usedPercentage := (usedSize / totalSize) * 100

        prediction := DiskRecord{
            Timestamp:      nextTime,
            DiskPath:       diskPath,
            UUID:           uuid,
            TotalSize:      totalSize,
            UsedSize:       usedSize,
            AvailableSize:  availableSize,
            UsedPercentage: usedPercentage,
            Prediction:     1,
        }
        predictions = append(predictions, prediction)
    }
    return predictions
}