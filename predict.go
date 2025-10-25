package main

import (
	"fmt"
	"math"
	"time"
)
var availableForecasters = map[string]ForecastAlgorithm{
	"crude-linear-extrapolation": CrudeLinearExtrapolation,
}


type Forecast struct {
    AlgorithmName string       `json:"algorithmName"`
    Data          []DiskRecord `json:"data"` 
    Error         string       `json:"error,omitempty"` 
}

type Forecasts struct {
    UUID      string             `json:"uuid"`
    Timestamp int64              `json:"timestamp"`
    Forecasts []Forecast         `json:"forecasts"` 
}

//returns an array of results of different forecast algorithms of the given uuid disk.
func GetDiskForecasts(uuid string, limit int) (Forecasts, error) {
    history, err := GetDiskRecords(uuid, limit)
	forecasts := Forecasts{
        UUID:      uuid,
        Timestamp: time.Now().Unix(),
        Forecasts: []Forecast{},
    }
    if err != nil {
        return forecasts, err
    }
    if len(history) < 3 {
        return forecasts, nil
    }
	for _, forecaster := range availableForecasters {
		forecastData, err := GenerateForecast(forecaster, history, limit, int64(APP.cfg.CheckIntervalSeconds))
		if err != nil {
			fmt.Errorf("forecast generation failed: %w", forecaster, forecastData, err)
			continue
		}
		var forecast = Forecast{
			AlgorithmName: fmt.Sprintf("%T", forecaster),
			Data:       forecastData,
		}
		forecasts.Forecasts = append(forecasts.Forecasts, forecast)
	}
    return forecasts, nil
}

type ForecastAlgorithm func([]DiskRecord, int, int64) []DiskRecord
func GenerateForecast(forecaster ForecastAlgorithm, history []DiskRecord, futurePoints int, intervalSeconds int64) ([]DiskRecord, error) {
	if len(history) < 2 {
		return nil, fmt.Errorf("not enough data points (%d provided) for forecast algorithm %s; minimum required is 2", len(history), forecaster)
	}
	return forecaster(history, futurePoints, intervalSeconds), nil  
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
    var forecasts []DiskRecord

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

        forecast := DiskRecord{
            Timestamp:      nextTime,
            DiskPath:       diskPath,
            UUID:           uuid,
            TotalSize:      totalSize,
            UsedSize:       usedSize,
            AvailableSize:  availableSize,
            UsedPercentage: usedPercentage,
            Forecast:     1,
        }
        forecasts = append(forecasts, forecast)
    }
    return forecasts
}