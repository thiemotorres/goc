package gpx

import (
	"math"

	"github.com/tkrajina/gpxgo/gpx"
)

// Point represents a track point with distance
type Point struct {
	Lat       float64
	Lon       float64
	Elevation float64
	Distance  float64 // Cumulative distance from start in meters
}

// Route represents a loaded GPX route
type Route struct {
	Name          string
	Points        []Point
	TotalDistance float64
	TotalAscent   float64
	TotalDescent  float64
}

// Load parses a GPX file
func Load(path string) (*Route, error) {
	gpxFile, err := gpx.ParseFile(path)
	if err != nil {
		return nil, err
	}

	route := &Route{}

	// Get track name
	if len(gpxFile.Tracks) > 0 {
		route.Name = gpxFile.Tracks[0].Name
	}

	// Collect all points
	var cumDistance float64
	var prevPoint *gpx.GPXPoint

	for _, track := range gpxFile.Tracks {
		for _, segment := range track.Segments {
			for i, pt := range segment.Points {
				if i > 0 && prevPoint != nil {
					dist := haversineDistance(
						prevPoint.Latitude, prevPoint.Longitude,
						pt.Latitude, pt.Longitude,
					)
					cumDistance += dist

					eleDiff := pt.Elevation.Value() - prevPoint.Elevation.Value()
					if eleDiff > 0 {
						route.TotalAscent += eleDiff
					} else {
						route.TotalDescent += -eleDiff
					}
				}

				route.Points = append(route.Points, Point{
					Lat:       pt.Latitude,
					Lon:       pt.Longitude,
					Elevation: pt.Elevation.Value(),
					Distance:  cumDistance,
				})

				prevPoint = &pt
			}
		}
	}

	route.TotalDistance = cumDistance
	return route, nil
}

// GradientAt returns gradient (%) at given distance
func (r *Route) GradientAt(distance float64) float64 {
	if len(r.Points) < 2 {
		return 0
	}

	// Find segment containing this distance
	for i := 1; i < len(r.Points); i++ {
		if r.Points[i].Distance >= distance {
			prev := r.Points[i-1]
			curr := r.Points[i]

			segmentDist := curr.Distance - prev.Distance
			if segmentDist == 0 {
				return 0
			}

			elevationChange := curr.Elevation - prev.Elevation
			return (elevationChange / segmentDist) * 100
		}
	}

	// Past end, return last segment gradient
	if len(r.Points) >= 2 {
		prev := r.Points[len(r.Points)-2]
		curr := r.Points[len(r.Points)-1]
		segmentDist := curr.Distance - prev.Distance
		if segmentDist > 0 {
			return ((curr.Elevation - prev.Elevation) / segmentDist) * 100
		}
	}

	return 0
}

// ElevationAt returns elevation at given distance
func (r *Route) ElevationAt(distance float64) float64 {
	if len(r.Points) == 0 {
		return 0
	}

	if distance <= 0 {
		return r.Points[0].Elevation
	}

	for i := 1; i < len(r.Points); i++ {
		if r.Points[i].Distance >= distance {
			prev := r.Points[i-1]
			curr := r.Points[i]

			// Interpolate
			segmentDist := curr.Distance - prev.Distance
			if segmentDist == 0 {
				return curr.Elevation
			}

			ratio := (distance - prev.Distance) / segmentDist
			return prev.Elevation + ratio*(curr.Elevation-prev.Elevation)
		}
	}

	return r.Points[len(r.Points)-1].Elevation
}

// PositionAt returns lat/lon at given distance
func (r *Route) PositionAt(distance float64) (lat, lon float64) {
	if len(r.Points) == 0 {
		return 0, 0
	}

	if distance <= 0 {
		return r.Points[0].Lat, r.Points[0].Lon
	}

	for i := 1; i < len(r.Points); i++ {
		if r.Points[i].Distance >= distance {
			prev := r.Points[i-1]
			curr := r.Points[i]

			segmentDist := curr.Distance - prev.Distance
			if segmentDist == 0 {
				return curr.Lat, curr.Lon
			}

			ratio := (distance - prev.Distance) / segmentDist
			return prev.Lat + ratio*(curr.Lat-prev.Lat),
				prev.Lon + ratio*(curr.Lon-prev.Lon)
		}
	}

	last := r.Points[len(r.Points)-1]
	return last.Lat, last.Lon
}

// haversineDistance calculates distance between two points in meters
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // meters

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
