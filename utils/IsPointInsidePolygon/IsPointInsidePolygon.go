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
