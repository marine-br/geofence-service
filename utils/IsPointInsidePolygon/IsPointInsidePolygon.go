package IsPointInsidePolygon

type Point struct {
	X, Y float64
}

// IsPointInPolygon determines if a point is inside a polygon
func IsPointInPolygon(point Point, polygon []Point) bool {
	n := len(polygon)
	inside := false

	j := n - 1
	for i := 0; i < n; i++ {
		if (polygon[i].Y > point.Y) != (polygon[j].Y > point.Y) &&
			point.X < (polygon[j].X-polygon[i].X)*(point.Y-polygon[i].Y)/(polygon[j].Y-polygon[i].Y)+polygon[i].X {
			inside = !inside
		}
		j = i
	}

	return inside
}

// IsPointInAnyPolygon reports whether the point is inside at least one of the
// polygons (a MultiPolygon or a FeatureCollection with several areas).
func IsPointInAnyPolygon(point Point, polygons [][]Point) bool {
	for _, polygon := range polygons {
		if IsPointInPolygon(point, polygon) {
			return true
		}
	}
	return false
}
