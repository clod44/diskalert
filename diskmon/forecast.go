package diskmon

import (
	"fmt"
	"time"
)
var availableForecasters = map[string]ForecastAlgorithm{
	"crude-linear-extrapolation": crudeLinearExtrapolation,
	"sma":						sma,
	"sma-extrapolation": 	   smaExtrapolation,
}


type Forecast struct {
    AlgorithmName string       `json:"algorithmName"`
    Data          []Point      `json:"data"` 
    Error         string       `json:"error,omitempty"` 
}

type Forecasts struct {
    UUID      string             `json:"uuid"`
    Timestamp int64              `json:"timestamp"`
    Forecasts []Forecast         `json:"forecasts"` 
}

type Point struct {
	X int64 `json:"x"` // timestamp
    Y float64 `json:"y"`
	Color string `json:"color"`
}

func clampSlicer(points []Point, minY float64, maxY float64) []Point {
	//from index 0 to n where the value of the n is above or lower than the specified values and the rest of the array is sliced away
	var clamped []Point
	for _, point := range points {
		if point.Y < minY || point.Y > maxY {
			break
		}
		clamped = append(clamped, point)
	}
	return clamped
}

//returns an array of results of different forecast algorithms of the given uuid disk.
//window is in number of records to consider from the past. usage of it may change from algorithm to algorithm
func GetDiskForecasts(uuid string, window int) (Forecasts, error) {
    history, err := GetDiskRecords(uuid, window)
	TotalSize := float64(history[0].TotalSize) //latest
	forecasts := Forecasts{
        UUID:      uuid,
        Timestamp: time.Now().Unix(),
        Forecasts: []Forecast{},
    }
    if err != nil {
        return forecasts, err
    }
	//convert history to points. forecasts work with simpler objects to minimize data transfer
	var points []Point
	for _, record := range history {
		point := Point{
			X: record.Timestamp,
			Y: record.UsedSize,
			Color: "rgba(150,150,150,1.0)",
		}
		points = append(points, point)
	}	
    if len(points) < 3 {
        return forecasts, nil
    }
	for name, forecaster := range availableForecasters {
		//window = 10 => futurePoints = 10
		forecastData, err := generateForecast(forecaster, points, window, int64(Cfg.CheckIntervalSeconds))
		if err != nil {
			fmt.Errorf("forecast generation failed: %w", name, forecastData, err)
			continue
		}
		var forecast = Forecast{
			AlgorithmName: name,
			Data:       clampSlicer(forecastData, 0, TotalSize),
		}
		forecasts.Forecasts = append(forecasts.Forecasts, forecast)
	}
    return forecasts, nil
}

type ForecastAlgorithm func([]Point, int, int64) []Point

func generateForecast(forecaster ForecastAlgorithm, points []Point, futurePoints int, intervalSeconds int64) ([]Point, error) {
	if len(points) < 2 {
		return nil, fmt.Errorf("not enough data points (%d provided) for forecast algorithm %s; minimum required is 2", len(points), forecaster)
	}
	if futurePoints <= 0 {
		return nil, fmt.Errorf("futurePoints must be greater than 0")
	}
	return forecaster(points, futurePoints, intervalSeconds), nil  
}

func sma(points []Point, _futurePoints int, _intervalSeconds int64)[]Point{
color := "rgba(255,0,150,1.0)"
	const windowSize = 3
	pointsLen := len(points)
	sma := make([]Point, pointsLen)
	copy(sma, points)

	if pointsLen < windowSize {
		return points
	}
	
	for i := range pointsLen {
		var sumUsedSize float64
		var divide = 1;
		for j := i; j < i+windowSize; j++ {
			if(j >= pointsLen){
				continue
			}
			divide++;
			sumUsedSize += points[j].Y
		}
		sma[i].Y = sumUsedSize / float64(divide);
		sma[i].Color = color
	}
	return sma
}
func smaExtrapolation(points []Point, futurePoints int, intervalSeconds int64)[]Point{
	smaPoints := sma(points, 0, 0)
	linear := crudeLinearExtrapolation(smaPoints, futurePoints, intervalSeconds)
	//linear=[6,5,4], sma=[3,2,1] => [6,5,4,3,2,1] (from the future to the past)
	return append(linear, smaPoints...)
}
func crudeLinearExtrapolation(points []Point, futurePoints int, intervalSeconds int64) []Point {
	color := "rgba(100,255,0,1.0)"
	latest := points[0]
    previous := points[1]
    timeDiff := latest.X - previous.X
    usedDiff := latest.Y - previous.Y
    
    if timeDiff <= 0 {
        return nil
    }
    m := usedDiff / float64(timeDiff) 
	f := func(x int64) float64 {
		return m*float64(x-latest.X) + latest.Y
    }
    newPoints := make([]Point, futurePoints)
	for i := 0; i < futurePoints; i++ {
		p := Point{
			X: latest.X + int64(i+1)*intervalSeconds,
			Y: f(latest.X + int64(i+1)*intervalSeconds),
			Color: color,
		}
		newPoints[i] = p;
	}
    return newPoints
}