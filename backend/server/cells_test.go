package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/bouk/httprouter"
	gocmp "github.com/google/go-cmp/cmp"
	cmp "github.com/snetsystems/cmp/backend"
	"github.com/snetsystems/cmp/backend/mocks"
)

func Test_Cells_CorrectAxis(t *testing.T) {
	t.Parallel()

	axisTests := []struct {
		name       string
		cell       *cmp.DashboardCell
		shouldFail bool
	}{
		{
			name: "correct axes",
			cell: &cmp.DashboardCell{
				Axes: map[string]cmp.Axis{
					"x": cmp.Axis{
						Bounds: []string{"0", "100"},
					},
					"y": cmp.Axis{
						Bounds: []string{"0", "100"},
					},
					"y2": cmp.Axis{
						Bounds: []string{"0", "100"},
					},
				},
			},
		},
		{
			name: "invalid axes present",
			cell: &cmp.DashboardCell{
				Axes: map[string]cmp.Axis{
					"axis of evil": cmp.Axis{
						Bounds: []string{"666", "666"},
					},
					"axis of awesome": cmp.Axis{
						Bounds: []string{"1337", "31337"},
					},
				},
			},
			shouldFail: true,
		},
		{
			name: "linear scale value",
			cell: &cmp.DashboardCell{
				Axes: map[string]cmp.Axis{
					"x": cmp.Axis{
						Scale:  "linear",
						Bounds: []string{"0", "100"},
					},
				},
			},
		},
		{
			name: "log scale value",
			cell: &cmp.DashboardCell{
				Axes: map[string]cmp.Axis{
					"x": cmp.Axis{
						Scale:  "log",
						Bounds: []string{"0", "100"},
					},
				},
			},
		},
		{
			name: "invalid scale value",
			cell: &cmp.DashboardCell{
				Axes: map[string]cmp.Axis{
					"x": cmp.Axis{
						Scale:  "potatoes",
						Bounds: []string{"0", "100"},
					},
				},
			},
			shouldFail: true,
		},
		{
			name: "base 10 axis",
			cell: &cmp.DashboardCell{
				Axes: map[string]cmp.Axis{
					"x": cmp.Axis{
						Base:   "10",
						Bounds: []string{"0", "100"},
					},
				},
			},
		},
		{
			name: "base 2 axis",
			cell: &cmp.DashboardCell{
				Axes: map[string]cmp.Axis{
					"x": cmp.Axis{
						Base:   "2",
						Bounds: []string{"0", "100"},
					},
				},
			},
		},
		{
			name: "invalid base",
			cell: &cmp.DashboardCell{
				Axes: map[string]cmp.Axis{
					"x": cmp.Axis{
						Base:   "all your base are belong to us",
						Bounds: []string{"0", "100"},
					},
				},
			},
			shouldFail: true,
		},
	}

	for _, test := range axisTests {
		t.Run(test.name, func(tt *testing.T) {
			if err := HasCorrectAxes(test.cell); err != nil && !test.shouldFail {
				t.Errorf("%q: Unexpected error: err: %s", test.name, err)
			} else if err == nil && test.shouldFail {
				t.Errorf("%q: Expected error and received none", test.name)
			}
		})
	}
}

func Test_Service_DashboardCells(t *testing.T) {
	cellsTests := []struct {
		name         string
		reqURL       *url.URL
		ctxParams    map[string]string
		mockResponse []cmp.DashboardCell
		expected     []cmp.DashboardCell
		expectedCode int
	}{
		{
			name: "happy path",
			reqURL: &url.URL{
				Path: "/cmp/v1/dashboards/1/cells",
			},
			ctxParams: map[string]string{
				"id": "1",
			},
			mockResponse: []cmp.DashboardCell{},
			expected:     []cmp.DashboardCell{},
			expectedCode: http.StatusOK,
		},
		{
			name: "cell axes should always be \"x\", \"y\", and \"y2\"",
			reqURL: &url.URL{
				Path: "/cmp/v1/dashboards/1/cells",
			},
			ctxParams: map[string]string{
				"id": "1",
			},
			mockResponse: []cmp.DashboardCell{
				{
					ID:             "3899be5a-f6eb-4347-b949-de2f4fbea859",
					X:              0,
					Y:              0,
					W:              4,
					H:              4,
					Name:           "CPU",
					Type:           "bar",
					Queries:        []cmp.DashboardQuery{},
					Axes:           map[string]cmp.Axis{},
					NoteVisibility: "",
				},
			},
			expected: []cmp.DashboardCell{
				{
					ID:         "3899be5a-f6eb-4347-b949-de2f4fbea859",
					X:          0,
					Y:          0,
					W:          4,
					H:          4,
					Name:       "CPU",
					Type:       "bar",
					Queries:    []cmp.DashboardQuery{},
					CellColors: []cmp.CellColor{},
					Axes: map[string]cmp.Axis{
						"x": cmp.Axis{
							Bounds: []string{"", ""},
						},
						"y": cmp.Axis{
							Bounds: []string{"", ""},
						},
						"y2": cmp.Axis{
							Bounds: []string{"", ""},
						},
					},
					NoteVisibility: "default",
				},
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, test := range cellsTests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// setup context with params
			ctx := context.Background()
			params := httprouter.Params{}
			for k, v := range test.ctxParams {
				params = append(params, httprouter.Param{
					Key:   k,
					Value: v,
				})
			}
			ctx = httprouter.WithParams(ctx, params)

			// setup response recorder and request
			rr := httptest.NewRecorder()
			req := httptest.NewRequest("GET", test.reqURL.RequestURI(), strings.NewReader("")).WithContext(ctx)

			// setup mock DashboardCells store and logger
			tlog := &mocks.TestLogger{}
			svc := &Service{
				Store: &mocks.Store{
					DashboardsStore: &mocks.DashboardsStore{
						GetF: func(ctx context.Context, id cmp.DashboardID) (cmp.Dashboard, error) {
							return cmp.Dashboard{
								ID:        cmp.DashboardID(1),
								Cells:     test.mockResponse,
								Templates: []cmp.Template{},
								Name:      "empty dashboard",
							}, nil
						},
					},
				},
				Logger: tlog,
			}

			// invoke DashboardCell handler
			svc.DashboardCells(rr, req)

			// setup frame to decode response into
			respFrame := []struct {
				cmp.DashboardCell
				Links json.RawMessage `json:"links"` // ignore links
			}{}

			// decode response
			resp := rr.Result()

			if resp.StatusCode != test.expectedCode {
				tlog.Dump(t)
				t.Fatalf("%q - Status codes do not match. Want %d (%s), Got %d (%s)", test.name, test.expectedCode, http.StatusText(test.expectedCode), resp.StatusCode, http.StatusText(resp.StatusCode))
			}

			if err := json.NewDecoder(resp.Body).Decode(&respFrame); err != nil {
				t.Fatalf("%q - Error unmarshaling response body: err: %s", test.name, err)
			}

			// extract actual
			actual := []cmp.DashboardCell{}
			for _, rsp := range respFrame {
				actual = append(actual, rsp.DashboardCell)
			}

			// compare actual and expected
			if !gocmp.Equal(actual, test.expected) {
				t.Fatalf("%q - Dashboard Cells do not match: diff: %s", test.name, gocmp.Diff(actual, test.expected))
			}
		})
	}
}

func TestHasCorrectColors(t *testing.T) {
	tests := []struct {
		name    string
		c       *cmp.DashboardCell
		wantErr bool
	}{
		{
			name: "min type is valid",
			c: &cmp.DashboardCell{
				CellColors: []cmp.CellColor{
					{
						Type: "min",
						Hex:  "#FFFFFF",
					},
				},
			},
		},
		{
			name: "max type is valid",
			c: &cmp.DashboardCell{
				CellColors: []cmp.CellColor{
					{
						Type: "max",
						Hex:  "#FFFFFF",
					},
				},
			},
		},
		{
			name: "threshold type is valid",
			c: &cmp.DashboardCell{
				CellColors: []cmp.CellColor{
					{
						Type: "threshold",
						Hex:  "#FFFFFF",
					},
				},
			},
		},
		{
			name: "invalid color type",
			c: &cmp.DashboardCell{
				CellColors: []cmp.CellColor{
					{
						Type: "unknown",
						Hex:  "#FFFFFF",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid color hex",
			c: &cmp.DashboardCell{
				CellColors: []cmp.CellColor{
					{
						Type: "min",
						Hex:  "bad",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := HasCorrectColors(tt.c); (err != nil) != tt.wantErr {
				t.Errorf("HasCorrectColors() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_ReplaceDashboardCell(t *testing.T) {
	tests := []struct {
		name            string
		DashboardsStore cmp.DashboardsStore
		ID              string
		CID             string
		w               *httptest.ResponseRecorder
		r               *http.Request
		want            string
	}{
		{
			name: "update cell retains query config",
			ID:   "1",
			CID:  "3c5c4102-fa40-4585-a8f9-917c77e37192",
			DashboardsStore: &mocks.DashboardsStore{
				UpdateF: func(ctx context.Context, target cmp.Dashboard) error {
					return nil
				},
				GetF: func(ctx context.Context, ID cmp.DashboardID) (cmp.Dashboard, error) {
					return cmp.Dashboard{
						ID: ID,
						Cells: []cmp.DashboardCell{
							{
								ID:   "3c5c4102-fa40-4585-a8f9-917c77e37192",
								W:    4,
								H:    4,
								Name: "Untitled Cell",
								Queries: []cmp.DashboardQuery{
									{
										Command: "SELECT mean(\"usage_user\") AS \"mean_usage_user\" FROM \"telegraf\".\"autogen\".\"cpu\" WHERE time > :dashboardTime: AND \"cpu\"=:cpu: GROUP BY :interval: FILL(null)",
										QueryConfig: cmp.QueryConfig{
											ID:              "3cd3eaa4-a4b8-44b3-b69e-0c7bf6b91d9e",
											Database:        "telegraf",
											Measurement:     "cpu",
											RetentionPolicy: "autogen",
											Fields: []cmp.Field{
												{
													Value: "mean",
													Type:  "func",
													Alias: "mean_usage_user",
													Args: []cmp.Field{
														{
															Value: "usage_user",
															Type:  "field",
														},
													},
												},
											},
											Tags: map[string][]string{
												"cpu": {
													"ChristohersMBP2.lan",
												},
											},
											GroupBy: cmp.GroupBy{
												Time: "2s",
												Tags: []string{},
											},
											AreTagsAccepted: true,
											Fill:            "null",
											RawText:         strPtr("SELECT mean(\"usage_user\") AS \"mean_usage_user\" FROM \"telegraf\".\"autogen\".\"cpu\" WHERE time > :dashboardTime: AND \"cpu\"=:cpu: GROUP BY :interval: FILL(null)"),
											Range: &cmp.DurationRange{
												Lower: "now() - 15m"},
											Shifts: []cmp.TimeShift{},
										},
										Type: "influxql",
									},
								},
								Axes: map[string]cmp.Axis{
									"x": {
										Bounds: []string{"", ""},
									},
									"y": {
										Bounds: []string{"", ""},
									},
									"y2": {
										Bounds: []string{"", ""},
									},
								},
								Type: "line",
								CellColors: []cmp.CellColor{
									{
										ID:    "0",
										Type:  "min",
										Hex:   "#00C9FF",
										Name:  "laser",
										Value: "0",
									},
									{
										ID:    "1",
										Type:  "max",
										Hex:   "#9394FF",
										Name:  "comet",
										Value: "100",
									},
								},
								NoteVisibility: "default",
							},
						},
					}, nil
				},
			},
			w: httptest.NewRecorder(),
			r: httptest.NewRequest("POST", "/queries", bytes.NewReader([]byte(`
				{
					"i": "3c5c4102-fa40-4585-a8f9-917c77e37192",
					"x": 0,
					"y": 0,
					"w": 4,
					"h": 4,
					"name": "Untitled Cell",
					"queries": [
					  {
						"queryConfig": {
						  "id": "3cd3eaa4-a4b8-44b3-b69e-0c7bf6b91d9e",
						  "database": "telegraf",
						  "measurement": "cpu",
						  "retentionPolicy": "autogen",
						  "fields": [
							{
							  "value": "mean",
							  "type": "func",
							  "alias": "mean_usage_user",
							  "args": [{"value": "usage_user", "type": "field", "alias": ""}]
							}
						  ],
						  "tags": {"cpu": ["ChristohersMBP2.lan"]},
						  "groupBy": {"time": "2s", "tags": []},
						  "areTagsAccepted": true,
						  "fill": "null",
						  "rawText":
							"SELECT mean(\"usage_user\") AS \"mean_usage_user\" FROM \"telegraf\".\"autogen\".\"cpu\" WHERE time > :dashboardTime: AND \"cpu\"=:cpu: GROUP BY :interval: FILL(null)",
						  "range": {"upper": "", "lower": "now() - 15m"},
						  "shifts": []
						},
						"query":
						  "SELECT mean(\"usage_user\") AS \"mean_usage_user\" FROM \"telegraf\".\"autogen\".\"cpu\" WHERE time > :dashboardTime: AND \"cpu\"=:cpu: GROUP BY :interval: FILL(null)",
						"source": null,
						"type": "influxql"
					  }
					],
					"axes": {
					  "x": {
						"bounds": ["",""],
						"label": "",
						"prefix": "",
						"suffix": "",
						"base": "",
						"scale": ""
					  },
					  "y": {
						"bounds": ["",""],
						"label": "",
						"prefix": "",
						"suffix": "",
						"base": "",
						"scale": ""
					  },
					  "y2": {
						"bounds": ["",""],
						"label": "",
						"prefix": "",
						"suffix": "",
						"base": "",
						"scale": ""
					  }
					},
					"type": "line",
					"colors": [
					  {"type": "min", "hex": "#00C9FF", "id": "0", "name": "laser", "value": "0"},
					  {
						"type": "max",
						"hex": "#9394FF",
						"id": "1",
						"name": "comet",
						"value": "100"
					  }
					],
					"note": "",
					"noteVisibility": "default",
					"links": {
					  "self":
						"/cmp/v1/dashboards/6/cells/3c5c4102-fa40-4585-a8f9-917c77e37192"
					}
				  }
				  `))),
			want: `{"i":"3c5c4102-fa40-4585-a8f9-917c77e37192","x":0,"y":0,"w":4,"h":4,"name":"Untitled Cell","queries":[{"query":"SELECT mean(\"usage_user\") AS \"mean_usage_user\" FROM \"telegraf\".\"autogen\".\"cpu\" WHERE time \u003e :dashboardTime: AND \"cpu\"=:cpu: GROUP BY :interval: FILL(null)","queryConfig":{"id":"3cd3eaa4-a4b8-44b3-b69e-0c7bf6b91d9e","database":"telegraf","measurement":"cpu","retentionPolicy":"autogen","fields":[{"value":"mean","type":"func","alias":"mean_usage_user","args":[{"value":"usage_user","type":"field","alias":""}]}],"tags":{"cpu":["ChristohersMBP2.lan"]},"groupBy":{"time":"2s","tags":[]},"areTagsAccepted":true,"fill":"null","rawText":"SELECT mean(\"usage_user\") AS \"mean_usage_user\" FROM \"telegraf\".\"autogen\".\"cpu\" WHERE time \u003e :dashboardTime: AND \"cpu\"=:cpu: GROUP BY :interval: FILL(null)","range":{"upper":"","lower":"now() - 15m"},"shifts":[]},"source":"","type":"influxql"}],"axes":{"x":{"bounds":["",""],"label":"","prefix":"","suffix":"","base":"","scale":""},"y":{"bounds":["",""],"label":"","prefix":"","suffix":"","base":"","scale":""},"y2":{"bounds":["",""],"label":"","prefix":"","suffix":"","base":"","scale":""}},"type":"line","colors":[{"id":"0","type":"min","hex":"#00C9FF","name":"laser","value":"0"},{"id":"1","type":"max","hex":"#9394FF","name":"comet","value":"100"}],"legend":{},"tableOptions":{"verticalTimeAxis":false,"sortBy":{"internalName":"","displayName":"","visible":false},"wrapping":"","fixFirstColumn":false},"fieldOptions":null,"timeFormat":"","decimalPlaces":{"isEnforced":false,"digits":0},"note":"","noteVisibility":"default","links":{"self":"/cmp/v1/dashboards/1/cells/3c5c4102-fa40-4585-a8f9-917c77e37192"}}
`,
		},
		{
			name: "dashboard doesn't exist",
			ID:   "1",
			DashboardsStore: &mocks.DashboardsStore{
				GetF: func(ctx context.Context, ID cmp.DashboardID) (cmp.Dashboard, error) {
					return cmp.Dashboard{}, fmt.Errorf("doesn't exist")
				},
			},
			w:    httptest.NewRecorder(),
			r:    httptest.NewRequest("PUT", "/cmp/v1/dashboards/1/cells/3c5c4102-fa40-4585-a8f9-917c77e37192", nil),
			want: `{"code":404,"message":"ID 1 not found"}`,
		},
		{
			name: "cell doesn't exist",
			ID:   "1",
			CID:  "3c5c4102-fa40-4585-a8f9-917c77e37192",
			DashboardsStore: &mocks.DashboardsStore{
				GetF: func(ctx context.Context, ID cmp.DashboardID) (cmp.Dashboard, error) {
					return cmp.Dashboard{}, nil
				},
			},
			w:    httptest.NewRecorder(),
			r:    httptest.NewRequest("PUT", "/cmp/v1/dashboards/1/cells/3c5c4102-fa40-4585-a8f9-917c77e37192", nil),
			want: `{"code":404,"message":"ID 3c5c4102-fa40-4585-a8f9-917c77e37192 not found"}`,
		},
		{
			name: "invalid query config",
			ID:   "1",
			CID:  "3c5c4102-fa40-4585-a8f9-917c77e37192",
			DashboardsStore: &mocks.DashboardsStore{
				GetF: func(ctx context.Context, ID cmp.DashboardID) (cmp.Dashboard, error) {
					return cmp.Dashboard{
						ID: ID,
						Cells: []cmp.DashboardCell{
							{
								ID: "3c5c4102-fa40-4585-a8f9-917c77e37192",
							},
						},
					}, nil
				},
			},
			w: httptest.NewRecorder(),
			r: httptest.NewRequest("PUT", "/cmp/v1/dashboards/1/cells/3c5c4102-fa40-4585-a8f9-917c77e37192", bytes.NewReader([]byte(`{
					"i": "3c5c4102-fa40-4585-a8f9-917c77e37192",
					"x": 0,
					"y": 0,
					"w": 4,
					"h": 4,
					"name": "Untitled Cell",
					"queries": [
					  {
						"queryConfig": {
						  "fields": [
							{
							  "value": "invalid",
							  "type": "invalidType"
							}
						  ]
						}
					  }
					]
				  }`))),
			want: `{"code":422,"message":"invalid field type \"invalidType\" ; expect func, field, integer, number, regex, wildcard"}`,
		},
		{
			name: "JSON is not parsable",
			ID:   "1",
			CID:  "3c5c4102-fa40-4585-a8f9-917c77e37192",
			DashboardsStore: &mocks.DashboardsStore{
				GetF: func(ctx context.Context, ID cmp.DashboardID) (cmp.Dashboard, error) {
					return cmp.Dashboard{
						ID: ID,
						Cells: []cmp.DashboardCell{
							{
								ID: "3c5c4102-fa40-4585-a8f9-917c77e37192",
							},
						},
					}, nil
				},
			},
			w:    httptest.NewRecorder(),
			r:    httptest.NewRequest("PUT", "/cmp/v1/dashboards/1/cells/3c5c4102-fa40-4585-a8f9-917c77e37192", nil),
			want: `{"code":400,"message":"Unparsable JSON"}`,
		},
		{
			name: "not able to update store returns error message",
			ID:   "1",
			CID:  "3c5c4102-fa40-4585-a8f9-917c77e37192",
			DashboardsStore: &mocks.DashboardsStore{
				UpdateF: func(ctx context.Context, target cmp.Dashboard) error {
					return fmt.Errorf("error")
				},
				GetF: func(ctx context.Context, ID cmp.DashboardID) (cmp.Dashboard, error) {
					return cmp.Dashboard{
						ID: ID,
						Cells: []cmp.DashboardCell{
							{
								ID: "3c5c4102-fa40-4585-a8f9-917c77e37192",
							},
						},
					}, nil
				},
			},
			w: httptest.NewRecorder(),
			r: httptest.NewRequest("PUT", "/cmp/v1/dashboards/1/cells/3c5c4102-fa40-4585-a8f9-917c77e37192", bytes.NewReader([]byte(`{
					"i": "3c5c4102-fa40-4585-a8f9-917c77e37192",
					"x": 0,
					"y": 0,
					"w": 4,
					"h": 4,
					"name": "Untitled Cell",
					"queries": [
					  {
						"queryConfig": {
						  "fields": [
							{
							  "value": "usage_user",
							  "type": "field"
							}
						  ]
						}
					  }
					]
				  }`))),
			want: `{"code":500,"message":"Error updating cell 3c5c4102-fa40-4585-a8f9-917c77e37192 in dashboard 1: error"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				Store: &mocks.Store{
					DashboardsStore: tt.DashboardsStore,
				},
				Logger: &mocks.TestLogger{},
			}
			tt.r = WithContext(tt.r.Context(), tt.r, map[string]string{
				"id":  tt.ID,
				"cid": tt.CID,
			})
			s.ReplaceDashboardCell(tt.w, tt.r)
			got := tt.w.Body.String()
			if got != tt.want {
				t.Errorf("ReplaceDashboardCell() = got/want\n%s\n%s\n", got, tt.want)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func Test_newCellResponses(t *testing.T) {
	tests := []struct {
		name   string
		dID    cmp.DashboardID
		dcells []cmp.DashboardCell
		want   []dashboardCellResponse
	}{
		{
			name: "all fields set",
			dID:  cmp.DashboardID(1),
			dcells: []cmp.DashboardCell{
				cmp.DashboardCell{
					ID:   "445f8dc0-4d73-4168-8477-f628690d18a3",
					X:    0,
					Y:    0,
					W:    4,
					H:    4,
					Name: "Untitled Cell",
					Queries: []cmp.DashboardQuery{
						{
							Command: "SELECT mean(\"usage_user\") AS \"mean_usage_user\" FROM \"telegraf\".\"autogen\".\"cpu\" WHERE time > :dashboardTime: AND \"cpu\"=:cpu: GROUP BY :interval: FILL(null)",
							Label:   "",
							QueryConfig: cmp.QueryConfig{
								ID:              "8d5ec6da-13a5-423e-9026-7bc45649766c",
								Database:        "telegraf",
								Measurement:     "cpu",
								RetentionPolicy: "autogen",
								Fields: []cmp.Field{
									{
										Value: "mean",
										Type:  "func",
										Alias: "mean_usage_user",
										Args: []cmp.Field{
											{
												Value: "usage_user",
												Type:  "field",
												Alias: "",
											},
										},
									},
								},
								Tags: map[string][]string{"cpu": []string{"ChristohersMBP2.lan"}},
								GroupBy: cmp.GroupBy{
									Time: "2s",
								},
								AreTagsAccepted: true,
								Fill:            "null",
								RawText:         strPtr("SELECT mean(\"usage_user\") AS \"mean_usage_user\" FROM \"telegraf\".\"autogen\".\"cpu\" WHERE time > :dashboardTime: AND \"cpu\"=:cpu: GROUP BY :interval: FILL(null)"),
								Range: &cmp.DurationRange{
									Lower: "now() - 15m",
								},
							},
							Source: "",
							Type:   "influxql",
						},
					},
					Axes: map[string]cmp.Axis{
						"x": cmp.Axis{
							Bounds: []string{"", ""},
						},
						"y": cmp.Axis{
							Bounds: []string{"", ""},
						},
						"y2": cmp.Axis{
							Bounds: []string{"", ""},
						},
					},
					Type: "line",
					CellColors: []cmp.CellColor{
						cmp.CellColor{ID: "0", Type: "min", Hex: "#00C9FF", Name: "laser", Value: "0"},
						cmp.CellColor{ID: "1", Type: "max", Hex: "#9394FF", Name: "comet", Value: "100"},
					},
					Legend: cmp.Legend{
						Type:        "static",
						Orientation: "bottom",
					},
					Note:           "",
					NoteVisibility: "showWhenNoData",
				},
			},
			want: []dashboardCellResponse{
				{
					DashboardCell: cmp.DashboardCell{
						ID:   "445f8dc0-4d73-4168-8477-f628690d18a3",
						W:    4,
						H:    4,
						Name: "Untitled Cell",
						Queries: []cmp.DashboardQuery{
							{
								Command: "SELECT mean(\"usage_user\") AS \"mean_usage_user\" FROM \"telegraf\".\"autogen\".\"cpu\" WHERE time > :dashboardTime: AND \"cpu\"=:cpu: GROUP BY :interval: FILL(null)",
								QueryConfig: cmp.QueryConfig{
									ID:              "8d5ec6da-13a5-423e-9026-7bc45649766c",
									Database:        "telegraf",
									Measurement:     "cpu",
									RetentionPolicy: "autogen",
									Fields: []cmp.Field{
										{
											Value: "mean",
											Type:  "func",
											Alias: "mean_usage_user",
											Args: []cmp.Field{
												{
													Value: "usage_user",
													Type:  "field",
												},
											},
										},
									},
									Tags: map[string][]string{"cpu": {"ChristohersMBP2.lan"}},
									GroupBy: cmp.GroupBy{
										Time: "2s",
									},
									AreTagsAccepted: true,
									Fill:            "null",
									RawText:         strPtr("SELECT mean(\"usage_user\") AS \"mean_usage_user\" FROM \"telegraf\".\"autogen\".\"cpu\" WHERE time > :dashboardTime: AND \"cpu\"=:cpu: GROUP BY :interval: FILL(null)"),
									Range: &cmp.DurationRange{
										Lower: "now() - 15m",
									},
								},
								Type: "influxql",
							},
						},
						Axes: map[string]cmp.Axis{
							"x": cmp.Axis{
								Bounds: []string{"", ""},
							},
							"y": cmp.Axis{
								Bounds: []string{"", ""},
							},
							"y2": cmp.Axis{
								Bounds: []string{"", ""},
							},
						},
						Type: "line",
						CellColors: []cmp.CellColor{
							{
								ID:    "0",
								Type:  "min",
								Hex:   "#00C9FF",
								Name:  "laser",
								Value: "0",
							},
							{
								ID:    "1",
								Type:  "max",
								Hex:   "#9394FF",
								Name:  "comet",
								Value: "100",
							},
						},
						Legend: cmp.Legend{
							Type:        "static",
							Orientation: "bottom",
						},
						Note:           "",
						NoteVisibility: "showWhenNoData",
					},
					Links: dashboardCellLinks{
						Self: "/cmp/v1/dashboards/1/cells/445f8dc0-4d73-4168-8477-f628690d18a3"},
				},
			},
		},
		{
			name: "nothing set",
			dID:  cmp.DashboardID(1),
			dcells: []cmp.DashboardCell{
				cmp.DashboardCell{
					ID:   "445f8dc0-4d73-4168-8477-f628690d18a3",
					X:    0,
					Y:    0,
					W:    4,
					H:    4,
					Name: "Untitled Cell",
				},
			},
			want: []dashboardCellResponse{
				{
					DashboardCell: cmp.DashboardCell{
						ID:      "445f8dc0-4d73-4168-8477-f628690d18a3",
						W:       4,
						H:       4,
						Name:    "Untitled Cell",
						Queries: []cmp.DashboardQuery{},
						Axes: map[string]cmp.Axis{
							"x": cmp.Axis{
								Bounds: []string{"", ""},
							},
							"y": cmp.Axis{
								Bounds: []string{"", ""},
							},
							"y2": cmp.Axis{
								Bounds: []string{"", ""},
							},
						},
						CellColors:     []cmp.CellColor{},
						Legend:         cmp.Legend{},
						Note:           "",
						NoteVisibility: "default",
					},
					Links: dashboardCellLinks{
						Self: "/cmp/v1/dashboards/1/cells/445f8dc0-4d73-4168-8477-f628690d18a3"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newCellResponses(tt.dID, tt.dcells); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newCellResponses() = got-/want+ %s", gocmp.Diff(got, tt.want))
			}
		})
	}
}

func TestHasCorrectLegend(t *testing.T) {
	tests := []struct {
		name    string
		c       *cmp.DashboardCell
		wantErr bool
	}{
		{
			name: "empty legend is ok",
			c:    &cmp.DashboardCell{},
		},
		{
			name: "must have both an orientation and type",
			c: &cmp.DashboardCell{
				Legend: cmp.Legend{
					Type: "static",
				},
			},
			wantErr: true,
		},
		{
			name: "must have both a type and orientation",
			c: &cmp.DashboardCell{
				Legend: cmp.Legend{
					Orientation: "bottom",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid types",
			c: &cmp.DashboardCell{
				Legend: cmp.Legend{
					Type:        "no such type",
					Orientation: "bottom",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid orientation",
			c: &cmp.DashboardCell{
				Legend: cmp.Legend{
					Type:        "static",
					Orientation: "no such orientation",
				},
			},
			wantErr: true,
		},
		{
			name: "orientation bottom valid",
			c: &cmp.DashboardCell{
				Legend: cmp.Legend{
					Type:        "static",
					Orientation: "bottom",
				},
			},
		},
		{
			name: "orientation top valid",
			c: &cmp.DashboardCell{
				Legend: cmp.Legend{
					Type:        "static",
					Orientation: "top",
				},
			},
		},
		{
			name: "orientation right valid",
			c: &cmp.DashboardCell{
				Legend: cmp.Legend{
					Type:        "static",
					Orientation: "right",
				},
			},
		},
		{
			name: "orientation left valid",
			c: &cmp.DashboardCell{
				Legend: cmp.Legend{
					Type:        "static",
					Orientation: "left",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := HasCorrectLegend(tt.c); (err != nil) != tt.wantErr {
				t.Errorf("HasCorrectLegend() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateNote(t *testing.T) {
	type want struct {
		Note           string
		NoteVisibility string
	}
	tests := []struct {
		name    string
		c       *cmp.DashboardCell
		want    want
		wantErr bool
	}{
		{
			name: "note text & visibility defaults",
			c:    &cmp.DashboardCell{},
			want: want{
				Note:           "",
				NoteVisibility: "default",
			},
			wantErr: false,
		},
		{
			name: "note text - allows non-html",
			c: &cmp.DashboardCell{
				Note: "pineapples are tasty",
			},
			want: want{
				Note:           "pineapples are tasty",
				NoteVisibility: "default",
			},
			wantErr: false,
		},
		{
			name: "note text - eliminates xss-vulnerable html",
			c: &cmp.DashboardCell{
				Note: `
<script>alert('bob');</script>
<p>benevolent pineapples paragraph</p>
<style>evil style</style>
<iframe>evil iframe</iframe>
<object>evil object</object>
<embed>evil embed</embed>
<base>evil base</base>
`,
			},
			want: want{
				Note: `

<p>benevolent pineapples paragraph</p>



evil embed
evil base
`,
				NoteVisibility: "default",
			},
			wantErr: false,
		},
		{
			name: "note visibility - valid default value",
			c: &cmp.DashboardCell{
				Note:           "",
				NoteVisibility: "default",
			},
			want: want{
				Note:           "",
				NoteVisibility: "default",
			},
			wantErr: false,
		},
		{
			name: "note visibility - valid non-default value",
			c: &cmp.DashboardCell{
				Note:           "",
				NoteVisibility: "showWhenNoData",
			},
			want: want{
				Note:           "",
				NoteVisibility: "showWhenNoData",
			},
			wantErr: false,
		},
		{
			name: "note visibility - invalid value",
			c: &cmp.DashboardCell{
				Note:           "",
				NoteVisibility: "pineapple",
			},
			want:    want{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNote(tt.c)
			if (err == nil && tt.wantErr) || (err != nil && !tt.wantErr) {
				t.Errorf("ValidateNote() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if tt.c.Note != tt.want.Note || tt.c.NoteVisibility != tt.want.NoteVisibility {
					t.Errorf("ValidateNote()\ngot = **%v**, %v\nwant = **%v**, %v", tt.c.Note, tt.c.NoteVisibility, tt.want.Note, tt.want.NoteVisibility)
				}
			}
		})
	}
}

func TestHasCorrectQueryType(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		c       *cmp.DashboardCell
	}{
		{
			name:    "Should error if type is not flux or influxql",
			wantErr: true,
			c: &cmp.DashboardCell{
				Queries: []cmp.DashboardQuery{
					{
						Type: "howdy",
					},
				},
			},
		},
		{
			name: "A flux query type",
			c: &cmp.DashboardCell{
				Queries: []cmp.DashboardQuery{
					{
						Type: "flux",
					},
				},
			},
		},
		{
			name: "An influxql query type",
			c: &cmp.DashboardCell{
				Queries: []cmp.DashboardQuery{
					{
						Type: "influxql",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := HasCorrectQueryType(tt.c); (err != nil) != tt.wantErr {
				t.Errorf("HasCorrectQueryType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
