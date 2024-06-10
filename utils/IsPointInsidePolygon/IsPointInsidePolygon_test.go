package IsPointInsidePolygon

import "testing"

func TestIsPointInPolygon(t *testing.T) {

	type args struct {
		point   Point
		polygon []Point
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "the point should be inside the polygon",
			args: args{
				point: Point{
					X: -71.227575,
					Y: 27.750384,
				},
				polygon: []Point{
					{
						X: -80.31142183805764,
						Y: 25.759993440387746,
					},
					{
						X: -64.74136041917353,
						Y: 32.32679987282724,
					},
					{
						X: -66.09706497543537,
						Y: 18.380895026206034,
					},
					{
						X: -80.31142183805764,
						Y: 25.759993440387746,
					},
				},
			},
			want: true,
		},
		{
			name: "should be inside MARINE TELEMATICS",
			args: args{
				point: Point{
					X: -27.599327,
					Y: -48.605718,
				},
				polygon: []Point{
					{Y: -48.60565411678576, X: -27.600050127488487},
					{Y: -48.605719721111626, X: -27.597440868325883},
					{Y: -48.60728492673141, X: -27.596360196755654},
					{Y: -48.60793904044644, X: -27.597087580984365},
					{Y: -48.60862281774649, X: -27.5979069757488},
					{Y: -48.60705325170399, X: -27.598984360719403},
					{Y: -48.60665880670879, X: -27.599249422897003},
					{Y: -48.606489223861864, X: -27.60005403460859},
					{Y: -48.60606511528023, X: -27.600049176486444},
					{Y: -48.60565411678576, X: -27.600050127488487},
				},
			},
			want: true,
		},
		{
			name: "the point should be outside the polygon",
			args: args{
				point: Point{
					X: -72.836955,
					Y: 34.803454,
				},
				polygon: []Point{
					{
						X: -80.31142183805764,
						Y: 25.759993440387746,
					},
					{
						X: -64.74136041917353,
						Y: 32.32679987282724,
					},
					{
						X: -66.09706497543537,
						Y: 18.380895026206034,
					},
					{
						X: -80.31142183805764,
						Y: 25.759993440387746,
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPointInPolygon(tt.args.point, tt.args.polygon); got != tt.want {
				t.Errorf("IsPointInPolygon() = %v, want %v", got, tt.want)
			}
		})
	}
}
