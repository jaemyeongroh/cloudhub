package server

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/bouk/httprouter"
	gocmp "github.com/google/go-cmp/cmp"
	cmp "github.com/snetsystems/cmp/backend"
	"github.com/snetsystems/cmp/backend/log"
	"github.com/snetsystems/cmp/backend/mocks"
)

func Test_ValidSourceRequest(t *testing.T) {
	type args struct {
		source       *cmp.Source
		defaultOrgID string
	}
	type wants struct {
		err    error
		source *cmp.Source
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "nil source",
			args: args{},
			wants: wants{
				err: fmt.Errorf("source must be non-nil"),
			},
		},
		{
			name: "missing url",
			args: args{
				source: &cmp.Source{
					ID:                 1,
					Name:               "I'm a really great source",
					Type:               cmp.InfluxDB,
					Username:           "fancy",
					Password:           "i'm so",
					SharedSecret:       "supersecret",
					MetaURL:            "http://www.so.meta.com",
					InsecureSkipVerify: true,
					Default:            true,
					Telegraf:           "telegraf",
					Organization:       "0",
				},
			},
			wants: wants{
				err: fmt.Errorf("url required"),
			},
		},
		{
			name: "invalid source type",
			args: args{
				source: &cmp.Source{
					ID:                 1,
					Name:               "I'm a really great source",
					Type:               "non-existent-type",
					Username:           "fancy",
					Password:           "i'm so",
					SharedSecret:       "supersecret",
					URL:                "http://www.any.url.com",
					MetaURL:            "http://www.so.meta.com",
					InsecureSkipVerify: true,
					Default:            true,
					Telegraf:           "telegraf",
					Organization:       "0",
				},
			},
			wants: wants{
				err: fmt.Errorf("invalid source type non-existent-type"),
			},
		},
		{
			name: "set organization to be default org if not specified",
			args: args{
				defaultOrgID: "2",
				source: &cmp.Source{
					ID:                 1,
					Name:               "I'm a really great source",
					Type:               cmp.InfluxDB,
					Username:           "fancy",
					Password:           "i'm so",
					SharedSecret:       "supersecret",
					URL:                "http://www.any.url.com",
					MetaURL:            "http://www.so.meta.com",
					InsecureSkipVerify: true,
					Default:            true,
					Telegraf:           "telegraf",
				},
			},
			wants: wants{
				source: &cmp.Source{
					ID:                 1,
					Name:               "I'm a really great source",
					Type:               cmp.InfluxDB,
					Username:           "fancy",
					Password:           "i'm so",
					SharedSecret:       "supersecret",
					URL:                "http://www.any.url.com",
					MetaURL:            "http://www.so.meta.com",
					InsecureSkipVerify: true,
					Default:            true,
					Organization:       "2",
					Telegraf:           "telegraf",
				},
			},
		},
		{
			name: "bad url",
			args: args{
				source: &cmp.Source{
					ID:                 1,
					Name:               "I'm a really great source",
					Type:               cmp.InfluxDB,
					Username:           "fancy",
					Password:           "i'm so",
					SharedSecret:       "supersecret",
					URL:                "im a bad url",
					MetaURL:            "http://www.so.meta.com",
					InsecureSkipVerify: true,
					Organization:       "0",
					Default:            true,
					Telegraf:           "telegraf",
				},
			},
			wants: wants{
				err: fmt.Errorf("invalid source URI: parse im a bad url: invalid URI for request"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidSourceRequest(tt.args.source, tt.args.defaultOrgID)
			if err == nil && tt.wants.err == nil {
				if diff := gocmp.Diff(tt.args.source, tt.wants.source); diff != "" {
					t.Errorf("%q. ValidSourceRequest():\n-got/+want\ndiff %s", tt.name, diff)
				}
				return
			}
			if err.Error() != tt.wants.err.Error() {
				t.Errorf("%q. ValidSourceRequest() = %q, want %q", tt.name, err, tt.wants.err)
			}
		})
	}
}

func Test_newSourceResponse(t *testing.T) {
	tests := []struct {
		name string
		src  cmp.Source
		want sourceResponse
	}{
		{
			name: "Test empty telegraf",
			src: cmp.Source{
				ID:       1,
				Telegraf: "",
			},
			want: sourceResponse{
				Source: cmp.Source{
					ID:       1,
					Telegraf: "telegraf",
				},
				AuthenticationMethod: "unknown",
				Links: sourceLinks{
					Self:        "/cmp/v1/sources/1",
					Services:    "/cmp/v1/sources/1/services",
					Proxy:       "/cmp/v1/sources/1/proxy",
					Queries:     "/cmp/v1/sources/1/queries",
					Write:       "/cmp/v1/sources/1/write",
					Kapacitors:  "/cmp/v1/sources/1/kapacitors",
					Users:       "/cmp/v1/sources/1/users",
					Permissions: "/cmp/v1/sources/1/permissions",
					Databases:   "/cmp/v1/sources/1/dbs",
					Annotations: "/cmp/v1/sources/1/annotations",
					Health:      "/cmp/v1/sources/1/health",
				},
			},
		},
		{
			name: "Test non-default telegraf",
			src: cmp.Source{
				ID:       1,
				Telegraf: "howdy",
			},
			want: sourceResponse{
				Source: cmp.Source{
					ID:       1,
					Telegraf: "howdy",
				},
				AuthenticationMethod: "unknown",
				Links: sourceLinks{
					Self:        "/cmp/v1/sources/1",
					Proxy:       "/cmp/v1/sources/1/proxy",
					Services:    "/cmp/v1/sources/1/services",
					Queries:     "/cmp/v1/sources/1/queries",
					Write:       "/cmp/v1/sources/1/write",
					Kapacitors:  "/cmp/v1/sources/1/kapacitors",
					Users:       "/cmp/v1/sources/1/users",
					Permissions: "/cmp/v1/sources/1/permissions",
					Databases:   "/cmp/v1/sources/1/dbs",
					Annotations: "/cmp/v1/sources/1/annotations",
					Health:      "/cmp/v1/sources/1/health",
				},
			},
		},
	}
	for _, tt := range tests {
		if got := newSourceResponse(context.Background(), tt.src); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. newSourceResponse() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestService_newSourceKapacitor(t *testing.T) {
	type fields struct {
		SourcesStore cmp.SourcesStore
		ServersStore cmp.ServersStore
		Logger       cmp.Logger
	}
	type args struct {
		ctx  context.Context
		src  cmp.Source
		kapa cmp.Server
	}
	srcCount := 0
	srvCount := 0
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantSrc int
		wantSrv int
		wantErr bool
	}{
		{
			name: "Add when no existing sources",
			fields: fields{
				SourcesStore: &mocks.SourcesStore{
					AllF: func(ctx context.Context) ([]cmp.Source, error) {
						return []cmp.Source{}, nil
					},
					AddF: func(ctx context.Context, src cmp.Source) (cmp.Source, error) {
						srcCount++
						src.ID = srcCount
						return src, nil
					},
				},
				ServersStore: &mocks.ServersStore{
					AddF: func(ctx context.Context, srv cmp.Server) (cmp.Server, error) {
						srvCount++
						return srv, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				src: cmp.Source{
					Name: "Influx 1",
				},
				kapa: cmp.Server{
					Name: "Kapa 1",
				},
			},
			wantSrc: 1,
			wantSrv: 1,
		},
		{
			name: "Should not add if existing source",
			fields: fields{
				SourcesStore: &mocks.SourcesStore{
					AllF: func(ctx context.Context) ([]cmp.Source, error) {
						return []cmp.Source{
							{
								Name: "Influx 1",
							},
						}, nil
					},
					AddF: func(ctx context.Context, src cmp.Source) (cmp.Source, error) {
						srcCount++
						src.ID = srcCount
						return src, nil
					},
				},
				ServersStore: &mocks.ServersStore{
					AddF: func(ctx context.Context, srv cmp.Server) (cmp.Server, error) {
						srvCount++
						return srv, nil
					},
				},
				Logger: &mocks.TestLogger{},
			},
			args: args{
				ctx: context.Background(),
				src: cmp.Source{
					Name: "Influx 1",
				},
				kapa: cmp.Server{
					Name: "Kapa 1",
				},
			},
			wantSrc: 0,
			wantSrv: 0,
		},
		{
			name: "Error if All returns error",
			fields: fields{
				SourcesStore: &mocks.SourcesStore{
					AllF: func(ctx context.Context) ([]cmp.Source, error) {
						return nil, fmt.Errorf("error")
					},
				},
				Logger: &mocks.TestLogger{},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "Error if Add returns error",
			fields: fields{
				SourcesStore: &mocks.SourcesStore{
					AllF: func(ctx context.Context) ([]cmp.Source, error) {
						return []cmp.Source{}, nil
					},
					AddF: func(ctx context.Context, src cmp.Source) (cmp.Source, error) {
						return cmp.Source{}, fmt.Errorf("error")
					},
				},
				Logger: &mocks.TestLogger{},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "Error if kapa add is error",
			fields: fields{
				SourcesStore: &mocks.SourcesStore{
					AllF: func(ctx context.Context) ([]cmp.Source, error) {
						return []cmp.Source{}, nil
					},
					AddF: func(ctx context.Context, src cmp.Source) (cmp.Source, error) {
						srcCount++
						src.ID = srcCount
						return src, nil
					},
				},
				ServersStore: &mocks.ServersStore{
					AddF: func(ctx context.Context, srv cmp.Server) (cmp.Server, error) {
						srvCount++
						return cmp.Server{}, fmt.Errorf("error")
					},
				},
				Logger: &mocks.TestLogger{},
			},
			args: args{
				ctx: context.Background(),
				src: cmp.Source{
					Name: "Influx 1",
				},
				kapa: cmp.Server{
					Name: "Kapa 1",
				},
			},
			wantSrc: 1,
			wantSrv: 1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srcCount = 0
			srvCount = 0
			h := &Service{
				Store: &mocks.Store{
					SourcesStore: tt.fields.SourcesStore,
					ServersStore: tt.fields.ServersStore,
				},
				Logger: tt.fields.Logger,
			}
			if err := h.newSourceKapacitor(tt.args.ctx, tt.args.src, tt.args.kapa); (err != nil) != tt.wantErr {
				t.Errorf("Service.newSourceKapacitor() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantSrc != srcCount {
				t.Errorf("Service.newSourceKapacitor() count = %d, wantSrc %d", srcCount, tt.wantSrc)
			}
		})
	}
}

func TestService_SourcesID(t *testing.T) {
	type fields struct {
		SourcesStore cmp.SourcesStore
		Logger       cmp.Logger
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name            string
		args            args
		fields          fields
		ID              string
		wantStatusCode  int
		wantContentType string
		wantBody        string
	}{
		{
			name: "Get source without defaultRP includes empty defaultRP in response",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"GET",
					"http://any.url",
					nil,
				),
			},
			fields: fields{
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, ID int) (cmp.Source, error) {
						return cmp.Source{
							ID: 1,
						}, nil
					},
				},
				Logger: log.New(log.DebugLevel),
			},
			ID:              "1",
			wantStatusCode:  200,
			wantContentType: "application/json",
			wantBody: `{"id":"1","name":"","url":"","default":false,"telegraf":"telegraf","organization":"","defaultRP":"","version":"Unknown","authentication":"unknown","links":{"self":"/cmp/v1/sources/1","kapacitors":"/cmp/v1/sources/1/kapacitors","services":"/cmp/v1/sources/1/services","proxy":"/cmp/v1/sources/1/proxy","queries":"/cmp/v1/sources/1/queries","write":"/cmp/v1/sources/1/write","permissions":"/cmp/v1/sources/1/permissions","users":"/cmp/v1/sources/1/users","databases":"/cmp/v1/sources/1/dbs","annotations":"/cmp/v1/sources/1/annotations","health":"/cmp/v1/sources/1/health"}}
`,
		},
	}
	for _, tt := range tests {
		tt.args.r = tt.args.r.WithContext(httprouter.WithParams(
			context.Background(),
			httprouter.Params{
				{
					Key:   "id",
					Value: tt.ID,
				},
			}))
		h := &Service{
			Store: &mocks.Store{
				SourcesStore: tt.fields.SourcesStore,
			},
			Logger: tt.fields.Logger,
		}
		h.SourcesID(tt.args.w, tt.args.r)

		resp := tt.args.w.Result()
		contentType := resp.Header.Get("Content-Type")
		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != tt.wantStatusCode {
			t.Errorf("%q. SourcesID() = got %v, want %v", tt.name, resp.StatusCode, tt.wantStatusCode)
		}
		if tt.wantContentType != "" && contentType != tt.wantContentType {
			t.Errorf("%q. SourcesID() = got %v, want %v", tt.name, contentType, tt.wantContentType)
		}
		if tt.wantBody != "" && string(body) != tt.wantBody {
			t.Errorf("%q. SourcesID() =\ngot  ***%v***\nwant ***%v***\n", tt.name, string(body), tt.wantBody)
		}

	}
}
func TestService_UpdateSource(t *testing.T) {
	type fields struct {
		SourcesStore       cmp.SourcesStore
		OrganizationsStore cmp.OrganizationsStore
		Logger             cmp.Logger
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name            string
		args            args
		fields          fields
		ID              string
		wantStatusCode  int
		wantContentType string
		wantBody        func(string) string
	}{
		{
			name: "Update source updates fields",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"PATCH",
					"http://any.url",
					nil),
			},
			fields: fields{
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, ID int) (cmp.Source, error) {
						return cmp.Source{
							ID: 1,
						}, nil
					},
					UpdateF: func(ctx context.Context, upd cmp.Source) error {
						return nil
					},
				},
				OrganizationsStore: &mocks.OrganizationsStore{
					DefaultOrganizationF: func(context.Context) (*cmp.Organization, error) {
						return &cmp.Organization{
							ID:   "1337",
							Name: "pineapple_kingdom",
						}, nil
					},
				},
				Logger: log.New(log.DebugLevel),
			},
			ID:              "1",
			wantStatusCode:  200,
			wantContentType: "application/json",
			wantBody: func(url string) string {
				return fmt.Sprintf(`{"id":"1","name":"marty","type":"influx","username":"bob","url":"%s","metaUrl":"http://murl","default":false,"telegraf":"murlin","organization":"1337","defaultRP":"pineapple","authentication":"basic","links":{"self":"/cmp/v1/sources/1","kapacitors":"/cmp/v1/sources/1/kapacitors","services":"/cmp/v1/sources/1/services","proxy":"/cmp/v1/sources/1/proxy","queries":"/cmp/v1/sources/1/queries","write":"/cmp/v1/sources/1/write","permissions":"/cmp/v1/sources/1/permissions","users":"/cmp/v1/sources/1/users","databases":"/cmp/v1/sources/1/dbs","annotations":"/cmp/v1/sources/1/annotations","health":"/cmp/v1/sources/1/health"}}
`, url)
			},
		},
	}
	for _, tt := range tests {
		h := &Service{
			Store: &mocks.Store{
				SourcesStore:       tt.fields.SourcesStore,
				OrganizationsStore: tt.fields.OrganizationsStore,
			},
			Logger: tt.fields.Logger,
		}
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
			w.Header().Set("X-Influxdb-Build", "ENT")
		}))
		defer ts.Close()

		tt.args.r = tt.args.r.WithContext(httprouter.WithParams(
			context.Background(),
			httprouter.Params{
				{
					Key:   "id",
					Value: tt.ID,
				},
			}))
		tt.args.r.Body = ioutil.NopCloser(
			bytes.NewReader([]byte(
				fmt.Sprintf(`{"name":"marty","password":"the_lake","username":"bob","type":"influx","telegraf":"murlin","defaultRP":"pineapple","url":"%s","metaUrl":"http://murl"}`, ts.URL)),
			),
		)
		h.UpdateSource(tt.args.w, tt.args.r)

		resp := tt.args.w.Result()
		contentType := resp.Header.Get("Content-Type")
		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != tt.wantStatusCode {
			t.Errorf("%q. UpdateSource() = got %v, want %v", tt.name, resp.StatusCode, tt.wantStatusCode)
		}
		if contentType != tt.wantContentType {
			t.Errorf("%q. UpdateSource() = got %v, want %v", tt.name, contentType, tt.wantContentType)
		}
		wantBody := tt.wantBody(ts.URL)
		if string(body) != wantBody {
			t.Errorf("%q. UpdateSource() =\ngot  ***%v***\nwant ***%v***\n", tt.name, string(body), wantBody)
		}
	}
}

func TestService_NewSourceUser(t *testing.T) {
	type fields struct {
		SourcesStore cmp.SourcesStore
		TimeSeries   TimeSeriesClient
		Logger       cmp.Logger
		UseAuth      bool
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		ID              string
		wantStatus      int
		wantContentType string
		wantBody        string
	}{
		{
			name: "New user for data source",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"POST",
					"http://local/cmp/v1/sources/1",
					ioutil.NopCloser(
						bytes.NewReader([]byte(`{"name": "marty", "password": "the_lake"}`)))),
			},
			fields: fields{
				UseAuth: true,
				Logger:  log.New(log.DebugLevel),
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, ID int) (cmp.Source, error) {
						return cmp.Source{
							ID:       1,
							Name:     "muh source",
							Username: "name",
							Password: "hunter2",
							URL:      "http://localhost:8086",
						}, nil
					},
				},
				TimeSeries: &mocks.TimeSeries{
					ConnectF: func(ctx context.Context, src *cmp.Source) error {
						return nil
					},
					UsersF: func(ctx context.Context) cmp.UsersStore {
						return &mocks.UsersStore{
							AddF: func(ctx context.Context, u *cmp.User) (*cmp.User, error) {
								return u, nil
							},
						}
					},
					RolesF: func(ctx context.Context) (cmp.RolesStore, error) {
						return nil, fmt.Errorf("no roles")
					},
				},
			},
			ID:              "1",
			wantStatus:      http.StatusCreated,
			wantContentType: "application/json",
			wantBody: `{"links":{"self":"/cmp/v1/sources/1/users/marty"},"name":"marty","permissions":[]}
`,
		},
		{
			name: "New user for data source with roles",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"POST",
					"http://local/cmp/v1/sources/1",
					ioutil.NopCloser(
						bytes.NewReader([]byte(`{"name": "marty", "password": "the_lake"}`)))),
			},
			fields: fields{
				UseAuth: true,
				Logger:  log.New(log.DebugLevel),
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, ID int) (cmp.Source, error) {
						return cmp.Source{
							ID:       1,
							Name:     "muh source",
							Username: "name",
							Password: "hunter2",
							URL:      "http://localhost:8086",
						}, nil
					},
				},
				TimeSeries: &mocks.TimeSeries{
					ConnectF: func(ctx context.Context, src *cmp.Source) error {
						return nil
					},
					UsersF: func(ctx context.Context) cmp.UsersStore {
						return &mocks.UsersStore{
							AddF: func(ctx context.Context, u *cmp.User) (*cmp.User, error) {
								return u, nil
							},
						}
					},
					RolesF: func(ctx context.Context) (cmp.RolesStore, error) {
						return nil, nil
					},
				},
			},
			ID:              "1",
			wantStatus:      http.StatusCreated,
			wantContentType: "application/json",
			wantBody: `{"links":{"self":"/cmp/v1/sources/1/users/marty"},"name":"marty","permissions":[],"roles":[]}
`,
		},
		{
			name: "Error adding user",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"POST",
					"http://local/cmp/v1/sources/1",
					ioutil.NopCloser(
						bytes.NewReader([]byte(`{"name": "marty", "password": "the_lake"}`)))),
			},
			fields: fields{
				UseAuth: true,
				Logger:  log.New(log.DebugLevel),
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, ID int) (cmp.Source, error) {
						return cmp.Source{
							ID:       1,
							Name:     "muh source",
							Username: "name",
							Password: "hunter2",
							URL:      "http://localhost:8086",
						}, nil
					},
				},
				TimeSeries: &mocks.TimeSeries{
					ConnectF: func(ctx context.Context, src *cmp.Source) error {
						return nil
					},
					UsersF: func(ctx context.Context) cmp.UsersStore {
						return &mocks.UsersStore{
							AddF: func(ctx context.Context, u *cmp.User) (*cmp.User, error) {
								return nil, fmt.Errorf("Weight Has Nothing to Do With It")
							},
						}
					},
				},
			},
			ID:              "1",
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
			wantBody:        `{"code":400,"message":"Weight Has Nothing to Do With It"}`,
		},
		{
			name: "Failure connecting to user store",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"POST",
					"http://local/cmp/v1/sources/1",
					ioutil.NopCloser(
						bytes.NewReader([]byte(`{"name": "marty", "password": "the_lake"}`)))),
			},
			fields: fields{
				UseAuth: true,
				Logger:  log.New(log.DebugLevel),
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, ID int) (cmp.Source, error) {
						return cmp.Source{
							ID:       1,
							Name:     "muh source",
							Username: "name",
							Password: "hunter2",
							URL:      "http://localhost:8086",
						}, nil
					},
				},
				TimeSeries: &mocks.TimeSeries{
					ConnectF: func(ctx context.Context, src *cmp.Source) error {
						return fmt.Errorf("Biff just happens to be my supervisor")
					},
				},
			},
			ID:              "1",
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
			wantBody:        `{"code":400,"message":"Unable to connect to source 1: Biff just happens to be my supervisor"}`,
		},
		{
			name: "Failure getting source",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"POST",
					"http://local/cmp/v1/sources/1",
					ioutil.NopCloser(
						bytes.NewReader([]byte(`{"name": "marty", "password": "the_lake"}`)))),
			},
			fields: fields{
				UseAuth: true,
				Logger:  log.New(log.DebugLevel),
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, ID int) (cmp.Source, error) {
						return cmp.Source{}, fmt.Errorf("No McFly ever amounted to anything in the history of Hill Valley")
					},
				},
			},
			ID:              "1",
			wantStatus:      http.StatusNotFound,
			wantContentType: "application/json",
			wantBody:        `{"code":404,"message":"ID 1 not found"}`,
		},
		{
			name: "Bad ID",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"POST",
					"http://local/cmp/v1/sources/1",
					ioutil.NopCloser(
						bytes.NewReader([]byte(`{"name": "marty", "password": "the_lake"}`)))),
			},
			fields: fields{
				UseAuth: true,
				Logger:  log.New(log.DebugLevel),
			},
			ID:              "BAD",
			wantStatus:      http.StatusUnprocessableEntity,
			wantContentType: "application/json",
			wantBody:        `{"code":422,"message":"Error converting ID BAD"}`,
		},
		{
			name: "Bad name",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"POST",
					"http://local/cmp/v1/sources/1",
					ioutil.NopCloser(
						bytes.NewReader([]byte(`{"password": "the_lake"}`)))),
			},
			fields: fields{
				UseAuth: true,
				Logger:  log.New(log.DebugLevel),
			},
			ID:              "BAD",
			wantStatus:      http.StatusUnprocessableEntity,
			wantContentType: "application/json",
			wantBody:        `{"code":422,"message":"Username required"}`,
		},
		{
			name: "Bad JSON",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"POST",
					"http://local/cmp/v1/sources/1",
					ioutil.NopCloser(
						bytes.NewReader([]byte(`{password}`)))),
			},
			fields: fields{
				UseAuth: true,
				Logger:  log.New(log.DebugLevel),
			},
			ID:              "BAD",
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
			wantBody:        `{"code":400,"message":"Unparsable JSON"}`,
		},
	}
	for _, tt := range tests {
		tt.args.r = tt.args.r.WithContext(httprouter.WithParams(
			context.Background(),
			httprouter.Params{
				{
					Key:   "id",
					Value: tt.ID,
				},
			}))

		h := &Service{
			Store: &mocks.Store{
				SourcesStore: tt.fields.SourcesStore,
			},
			TimeSeriesClient: tt.fields.TimeSeries,
			Logger:           tt.fields.Logger,
			UseAuth:          tt.fields.UseAuth,
		}

		h.NewSourceUser(tt.args.w, tt.args.r)

		resp := tt.args.w.Result()
		content := resp.Header.Get("Content-Type")
		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != tt.wantStatus {
			t.Errorf("%q. NewSourceUser() = %v, want %v", tt.name, resp.StatusCode, tt.wantStatus)
		}
		if tt.wantContentType != "" && content != tt.wantContentType {
			t.Errorf("%q. NewSourceUser() = %v, want %v", tt.name, content, tt.wantContentType)
		}
		if tt.wantBody != "" && string(body) != tt.wantBody {
			t.Errorf("%q. NewSourceUser() = \n***%v***\n,\nwant\n***%v***", tt.name, string(body), tt.wantBody)
		}
	}
}

func TestService_SourceUsers(t *testing.T) {
	type fields struct {
		SourcesStore cmp.SourcesStore
		TimeSeries   TimeSeriesClient
		Logger       cmp.Logger
		UseAuth      bool
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		ID              string
		wantStatus      int
		wantContentType string
		wantBody        string
	}{
		{
			name: "All users for data source",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"GET",
					"http://local/cmp/v1/sources/1",
					nil),
			},
			fields: fields{
				UseAuth: true,
				Logger:  log.New(log.DebugLevel),
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, ID int) (cmp.Source, error) {
						return cmp.Source{
							ID:       1,
							Name:     "muh source",
							Username: "name",
							Password: "hunter2",
							URL:      "http://localhost:8086",
						}, nil
					},
				},
				TimeSeries: &mocks.TimeSeries{
					ConnectF: func(ctx context.Context, src *cmp.Source) error {
						return nil
					},
					RolesF: func(ctx context.Context) (cmp.RolesStore, error) {
						return nil, fmt.Errorf("no roles")
					},
					UsersF: func(ctx context.Context) cmp.UsersStore {
						return &mocks.UsersStore{
							AllF: func(ctx context.Context) ([]cmp.User, error) {
								return []cmp.User{
									{
										Name:   "strickland",
										Passwd: "discipline",
										Permissions: cmp.Permissions{
											{
												Scope:   cmp.AllScope,
												Allowed: cmp.Allowances{"READ"},
											},
										},
									},
								}, nil
							},
						}
					},
				},
			},
			ID:              "1",
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
			wantBody: `{"users":[{"links":{"self":"/cmp/v1/sources/1/users/strickland"},"name":"strickland","permissions":[{"scope":"all","allowed":["READ"]}]}]}
`,
		},
		{
			name: "All users for data source with roles",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"GET",
					"http://local/cmp/v1/sources/1",
					nil),
			},
			fields: fields{
				UseAuth: true,
				Logger:  log.New(log.DebugLevel),
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, ID int) (cmp.Source, error) {
						return cmp.Source{
							ID:       1,
							Name:     "muh source",
							Username: "name",
							Password: "hunter2",
							URL:      "http://localhost:8086",
						}, nil
					},
				},
				TimeSeries: &mocks.TimeSeries{
					ConnectF: func(ctx context.Context, src *cmp.Source) error {
						return nil
					},
					RolesF: func(ctx context.Context) (cmp.RolesStore, error) {
						return nil, nil
					},
					UsersF: func(ctx context.Context) cmp.UsersStore {
						return &mocks.UsersStore{
							AllF: func(ctx context.Context) ([]cmp.User, error) {
								return []cmp.User{
									{
										Name:   "strickland",
										Passwd: "discipline",
										Permissions: cmp.Permissions{
											{
												Scope:   cmp.AllScope,
												Allowed: cmp.Allowances{"READ"},
											},
										},
									},
								}, nil
							},
						}
					},
				},
			},
			ID:              "1",
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
			wantBody: `{"users":[{"links":{"self":"/cmp/v1/sources/1/users/strickland"},"name":"strickland","permissions":[{"scope":"all","allowed":["READ"]}],"roles":[]}]}
`,
		},
	}
	for _, tt := range tests {
		tt.args.r = tt.args.r.WithContext(httprouter.WithParams(
			context.Background(),
			httprouter.Params{
				{
					Key:   "id",
					Value: tt.ID,
				},
			}))
		h := &Service{
			Store: &mocks.Store{
				SourcesStore: tt.fields.SourcesStore,
			},
			TimeSeriesClient: tt.fields.TimeSeries,
			Logger:           tt.fields.Logger,
			UseAuth:          tt.fields.UseAuth,
		}

		h.SourceUsers(tt.args.w, tt.args.r)
		resp := tt.args.w.Result()
		content := resp.Header.Get("Content-Type")
		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != tt.wantStatus {
			t.Errorf("%q. SourceUsers() = %v, want %v", tt.name, resp.StatusCode, tt.wantStatus)
		}
		if tt.wantContentType != "" && content != tt.wantContentType {
			t.Errorf("%q. SourceUsers() = %v, want %v", tt.name, content, tt.wantContentType)
		}
		if tt.wantBody != "" && string(body) != tt.wantBody {
			t.Errorf("%q. SourceUsers() = \n***%v***\n,\nwant\n***%v***", tt.name, string(body), tt.wantBody)
		}
	}
}

func TestService_SourceUserID(t *testing.T) {
	type fields struct {
		SourcesStore cmp.SourcesStore
		TimeSeries   TimeSeriesClient
		Logger       cmp.Logger
		UseAuth      bool
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		ID              string
		UID             string
		wantStatus      int
		wantContentType string
		wantBody        string
	}{
		{
			name: "Single user for data source",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"GET",
					"http://local/cmp/v1/sources/1",
					nil),
			},
			fields: fields{
				UseAuth: true,
				Logger:  log.New(log.DebugLevel),
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, ID int) (cmp.Source, error) {
						return cmp.Source{
							ID:       1,
							Name:     "muh source",
							Username: "name",
							Password: "hunter2",
							URL:      "http://localhost:8086",
						}, nil
					},
				},
				TimeSeries: &mocks.TimeSeries{
					ConnectF: func(ctx context.Context, src *cmp.Source) error {
						return nil
					},
					RolesF: func(ctx context.Context) (cmp.RolesStore, error) {
						return nil, fmt.Errorf("no roles")
					},
					UsersF: func(ctx context.Context) cmp.UsersStore {
						return &mocks.UsersStore{
							GetF: func(ctx context.Context, q cmp.UserQuery) (*cmp.User, error) {
								return &cmp.User{
									Name:   "strickland",
									Passwd: "discipline",
									Permissions: cmp.Permissions{
										{
											Scope:   cmp.AllScope,
											Allowed: cmp.Allowances{"READ"},
										},
									},
								}, nil
							},
						}
					},
				},
			},
			ID:              "1",
			UID:             "strickland",
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
			wantBody: `{"links":{"self":"/cmp/v1/sources/1/users/strickland"},"name":"strickland","permissions":[{"scope":"all","allowed":["READ"]}]}
`,
		},
		{
			name: "Single user for data source",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"GET",
					"http://local/cmp/v1/sources/1",
					nil),
			},
			fields: fields{
				UseAuth: true,
				Logger:  log.New(log.DebugLevel),
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, ID int) (cmp.Source, error) {
						return cmp.Source{
							ID:       1,
							Name:     "muh source",
							Username: "name",
							Password: "hunter2",
							URL:      "http://localhost:8086",
						}, nil
					},
				},
				TimeSeries: &mocks.TimeSeries{
					ConnectF: func(ctx context.Context, src *cmp.Source) error {
						return nil
					},
					RolesF: func(ctx context.Context) (cmp.RolesStore, error) {
						return nil, nil
					},
					UsersF: func(ctx context.Context) cmp.UsersStore {
						return &mocks.UsersStore{
							GetF: func(ctx context.Context, q cmp.UserQuery) (*cmp.User, error) {
								return &cmp.User{
									Name:   "strickland",
									Passwd: "discipline",
									Permissions: cmp.Permissions{
										{
											Scope:   cmp.AllScope,
											Allowed: cmp.Allowances{"READ"},
										},
									},
								}, nil
							},
						}
					},
				},
			},
			ID:              "1",
			UID:             "strickland",
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
			wantBody: `{"links":{"self":"/cmp/v1/sources/1/users/strickland"},"name":"strickland","permissions":[{"scope":"all","allowed":["READ"]}],"roles":[]}
`,
		},
	}
	for _, tt := range tests {
		tt.args.r = tt.args.r.WithContext(httprouter.WithParams(
			context.Background(),
			httprouter.Params{
				{
					Key:   "id",
					Value: tt.ID,
				},
			}))
		h := &Service{
			Store: &mocks.Store{
				SourcesStore: tt.fields.SourcesStore,
			},
			TimeSeriesClient: tt.fields.TimeSeries,
			Logger:           tt.fields.Logger,
			UseAuth:          tt.fields.UseAuth,
		}

		h.SourceUserID(tt.args.w, tt.args.r)
		resp := tt.args.w.Result()
		content := resp.Header.Get("Content-Type")
		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != tt.wantStatus {
			t.Errorf("%q. SourceUserID() = %v, want %v", tt.name, resp.StatusCode, tt.wantStatus)
		}
		if tt.wantContentType != "" && content != tt.wantContentType {
			t.Errorf("%q. SourceUserID() = %v, want %v", tt.name, content, tt.wantContentType)
		}
		if tt.wantBody != "" && string(body) != tt.wantBody {
			t.Errorf("%q. SourceUserID() = \n***%v***\n,\nwant\n***%v***", tt.name, string(body), tt.wantBody)
		}
	}
}

func TestService_RemoveSourceUser(t *testing.T) {
	type fields struct {
		SourcesStore cmp.SourcesStore
		TimeSeries   TimeSeriesClient
		Logger       cmp.Logger
		UseAuth      bool
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		ID              string
		UID             string
		wantStatus      int
		wantContentType string
		wantBody        string
	}{
		{
			name: "Delete user for data source",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"GET",
					"http://local/cmp/v1/sources/1",
					nil),
			},
			fields: fields{
				UseAuth: true,
				Logger:  log.New(log.DebugLevel),
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, ID int) (cmp.Source, error) {
						return cmp.Source{
							ID:       1,
							Name:     "muh source",
							Username: "name",
							Password: "hunter2",
							URL:      "http://localhost:8086",
						}, nil
					},
				},
				TimeSeries: &mocks.TimeSeries{
					ConnectF: func(ctx context.Context, src *cmp.Source) error {
						return nil
					},
					UsersF: func(ctx context.Context) cmp.UsersStore {
						return &mocks.UsersStore{
							DeleteF: func(ctx context.Context, u *cmp.User) error {
								return nil
							},
						}
					},
				},
			},
			ID:         "1",
			UID:        "strickland",
			wantStatus: http.StatusNoContent,
		},
	}
	for _, tt := range tests {
		tt.args.r = tt.args.r.WithContext(httprouter.WithParams(
			context.Background(),
			httprouter.Params{
				{
					Key:   "id",
					Value: tt.ID,
				},
			}))
		h := &Service{
			Store: &mocks.Store{
				SourcesStore: tt.fields.SourcesStore,
			},
			TimeSeriesClient: tt.fields.TimeSeries,
			Logger:           tt.fields.Logger,
			UseAuth:          tt.fields.UseAuth,
		}
		h.RemoveSourceUser(tt.args.w, tt.args.r)
		resp := tt.args.w.Result()
		content := resp.Header.Get("Content-Type")
		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != tt.wantStatus {
			t.Errorf("%q. RemoveSourceUser() = %v, want %v", tt.name, resp.StatusCode, tt.wantStatus)
		}
		if tt.wantContentType != "" && content != tt.wantContentType {
			t.Errorf("%q. RemoveSourceUser() = %v, want %v", tt.name, content, tt.wantContentType)
		}
		if tt.wantBody != "" && string(body) != tt.wantBody {
			t.Errorf("%q. RemoveSourceUser() = \n***%v***\n,\nwant\n***%v***", tt.name, string(body), tt.wantBody)
		}
	}
}

func TestService_UpdateSourceUser(t *testing.T) {
	type fields struct {
		SourcesStore cmp.SourcesStore
		TimeSeries   TimeSeriesClient
		Logger       cmp.Logger
		UseAuth      bool
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		ID              string
		UID             string
		wantStatus      int
		wantContentType string
		wantBody        string
	}{
		{
			name: "Update user password for data source",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"POST",
					"http://local/cmp/v1/sources/1",
					ioutil.NopCloser(
						bytes.NewReader([]byte(`{"name": "marty", "password": "the_lake"}`)))),
			},
			fields: fields{
				UseAuth: true,
				Logger:  log.New(log.DebugLevel),
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, ID int) (cmp.Source, error) {
						return cmp.Source{
							ID:       1,
							Name:     "muh source",
							Username: "name",
							Password: "hunter2",
							URL:      "http://localhost:8086",
						}, nil
					},
				},
				TimeSeries: &mocks.TimeSeries{
					ConnectF: func(ctx context.Context, src *cmp.Source) error {
						return nil
					},
					RolesF: func(ctx context.Context) (cmp.RolesStore, error) {
						return nil, fmt.Errorf("no roles")
					},
					UsersF: func(ctx context.Context) cmp.UsersStore {
						return &mocks.UsersStore{
							UpdateF: func(ctx context.Context, u *cmp.User) error {
								return nil
							},
							GetF: func(ctx context.Context, q cmp.UserQuery) (*cmp.User, error) {
								return &cmp.User{
									Name: "marty",
								}, nil
							},
						}
					},
				},
			},
			ID:              "1",
			UID:             "marty",
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
			wantBody: `{"links":{"self":"/cmp/v1/sources/1/users/marty"},"name":"marty","permissions":[]}
`,
		},
		{
			name: "Update user password for data source with roles",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"POST",
					"http://local/cmp/v1/sources/1",
					ioutil.NopCloser(
						bytes.NewReader([]byte(`{"name": "marty", "password": "the_lake"}`)))),
			},
			fields: fields{
				UseAuth: true,
				Logger:  log.New(log.DebugLevel),
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, ID int) (cmp.Source, error) {
						return cmp.Source{
							ID:       1,
							Name:     "muh source",
							Username: "name",
							Password: "hunter2",
							URL:      "http://localhost:8086",
						}, nil
					},
				},
				TimeSeries: &mocks.TimeSeries{
					ConnectF: func(ctx context.Context, src *cmp.Source) error {
						return nil
					},
					RolesF: func(ctx context.Context) (cmp.RolesStore, error) {
						return nil, nil
					},
					UsersF: func(ctx context.Context) cmp.UsersStore {
						return &mocks.UsersStore{
							UpdateF: func(ctx context.Context, u *cmp.User) error {
								return nil
							},
							GetF: func(ctx context.Context, q cmp.UserQuery) (*cmp.User, error) {
								return &cmp.User{
									Name: "marty",
								}, nil
							},
						}
					},
				},
			},
			ID:              "1",
			UID:             "marty",
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
			wantBody: `{"links":{"self":"/cmp/v1/sources/1/users/marty"},"name":"marty","permissions":[],"roles":[]}
`,
		},
		{
			name: "Invalid update JSON",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"POST",
					"http://local/cmp/v1/sources/1",
					ioutil.NopCloser(
						bytes.NewReader([]byte(`{"name": "marty"}`)))),
			},
			fields: fields{
				UseAuth: true,
				Logger:  log.New(log.DebugLevel),
			},
			ID:              "1",
			UID:             "marty",
			wantStatus:      http.StatusUnprocessableEntity,
			wantContentType: "application/json",
			wantBody:        `{"code":422,"message":"No fields to update"}`,
		},
	}
	for _, tt := range tests {
		tt.args.r = tt.args.r.WithContext(httprouter.WithParams(
			context.Background(),
			httprouter.Params{
				{
					Key:   "id",
					Value: tt.ID,
				},
				{
					Key:   "uid",
					Value: tt.UID,
				},
			}))
		h := &Service{
			Store: &mocks.Store{
				SourcesStore: tt.fields.SourcesStore,
			},
			TimeSeriesClient: tt.fields.TimeSeries,
			Logger:           tt.fields.Logger,
			UseAuth:          tt.fields.UseAuth,
		}
		h.UpdateSourceUser(tt.args.w, tt.args.r)
		resp := tt.args.w.Result()
		content := resp.Header.Get("Content-Type")
		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != tt.wantStatus {
			t.Errorf("%q. UpdateSourceUser() = %v, want %v", tt.name, resp.StatusCode, tt.wantStatus)
		}
		if tt.wantContentType != "" && content != tt.wantContentType {
			t.Errorf("%q. UpdateSourceUser() = %v, want %v", tt.name, content, tt.wantContentType)
		}
		if tt.wantBody != "" && string(body) != tt.wantBody {
			t.Errorf("%q. UpdateSourceUser() = \n***%v***\n,\nwant\n***%v***", tt.name, string(body), tt.wantBody)
		}
	}
}

func TestService_NewSourceRole(t *testing.T) {
	type fields struct {
		SourcesStore cmp.SourcesStore
		TimeSeries   TimeSeriesClient
		Logger       cmp.Logger
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		ID              string
		wantStatus      int
		wantContentType string
		wantBody        string
	}{
		{
			name: "Bad JSON",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"POST",
					"http://server.local/cmp/v1/sources/1/roles",
					ioutil.NopCloser(
						bytes.NewReader([]byte(`{BAD}`)))),
			},
			fields: fields{
				Logger: log.New(log.DebugLevel),
			},
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
			wantBody:        `{"code":400,"message":"Unparsable JSON"}`,
		},
		{
			name: "Invalid request",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"POST",
					"http://server.local/cmp/v1/sources/1/roles",
					ioutil.NopCloser(
						bytes.NewReader([]byte(`{"name": ""}`)))),
			},
			fields: fields{
				Logger: log.New(log.DebugLevel),
			},
			ID:              "1",
			wantStatus:      http.StatusUnprocessableEntity,
			wantContentType: "application/json",
			wantBody:        `{"code":422,"message":"Name is required for a role"}`,
		},
		{
			name: "Invalid source ID",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"POST",
					"http://server.local/cmp/v1/sources/1/roles",
					ioutil.NopCloser(
						bytes.NewReader([]byte(`{"name": "newrole"}`)))),
			},
			fields: fields{
				Logger: log.New(log.DebugLevel),
			},
			ID:              "BADROLE",
			wantStatus:      http.StatusUnprocessableEntity,
			wantContentType: "application/json",
			wantBody:        `{"code":422,"message":"Error converting ID BADROLE"}`,
		},
		{
			name: "Source doesn't support roles",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"POST",
					"http://server.local/cmp/v1/sources/1/roles",
					ioutil.NopCloser(
						bytes.NewReader([]byte(`{"name": "role"}`)))),
			},
			fields: fields{
				Logger: log.New(log.DebugLevel),
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, ID int) (cmp.Source, error) {
						return cmp.Source{
							ID:       1,
							Name:     "muh source",
							Username: "name",
							Password: "hunter2",
							URL:      "http://localhost:8086",
						}, nil
					},
				},
				TimeSeries: &mocks.TimeSeries{
					ConnectF: func(ctx context.Context, src *cmp.Source) error {
						return nil
					},
					RolesF: func(ctx context.Context) (cmp.RolesStore, error) {
						return nil, fmt.Errorf("roles not supported")
					},
				},
			},
			ID:              "1",
			wantStatus:      http.StatusNotFound,
			wantContentType: "application/json",
			wantBody:        `{"code":404,"message":"Source 1 does not have role capability"}`,
		},
		{
			name: "Unable to add role to server",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"POST",
					"http://server.local/cmp/v1/sources/1/roles",
					ioutil.NopCloser(
						bytes.NewReader([]byte(`{"name": "role"}`)))),
			},
			fields: fields{
				Logger: log.New(log.DebugLevel),
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, ID int) (cmp.Source, error) {
						return cmp.Source{
							ID:       1,
							Name:     "muh source",
							Username: "name",
							Password: "hunter2",
							URL:      "http://localhost:8086",
						}, nil
					},
				},
				TimeSeries: &mocks.TimeSeries{
					ConnectF: func(ctx context.Context, src *cmp.Source) error {
						return nil
					},
					RolesF: func(ctx context.Context) (cmp.RolesStore, error) {
						return &mocks.RolesStore{
							AddF: func(ctx context.Context, u *cmp.Role) (*cmp.Role, error) {
								return nil, fmt.Errorf("server had and issue")
							},
							GetF: func(ctx context.Context, name string) (*cmp.Role, error) {
								return nil, fmt.Errorf("No such role")
							},
						}, nil
					},
				},
			},
			ID:              "1",
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
			wantBody:        `{"code":400,"message":"server had and issue"}`,
		},
		{
			name: "New role for data source",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"POST",
					"http://server.local/cmp/v1/sources/1/roles",
					ioutil.NopCloser(
						bytes.NewReader([]byte(`{"name": "biffsgang","users": [{"name": "match"},{"name": "skinhead"},{"name": "3-d"}]}`)))),
			},
			fields: fields{
				Logger: log.New(log.DebugLevel),
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, ID int) (cmp.Source, error) {
						return cmp.Source{
							ID:       1,
							Name:     "muh source",
							Username: "name",
							Password: "hunter2",
							URL:      "http://localhost:8086",
						}, nil
					},
				},
				TimeSeries: &mocks.TimeSeries{
					ConnectF: func(ctx context.Context, src *cmp.Source) error {
						return nil
					},
					RolesF: func(ctx context.Context) (cmp.RolesStore, error) {
						return &mocks.RolesStore{
							AddF: func(ctx context.Context, u *cmp.Role) (*cmp.Role, error) {
								return u, nil
							},
							GetF: func(ctx context.Context, name string) (*cmp.Role, error) {
								return nil, fmt.Errorf("no such role")
							},
						}, nil
					},
				},
			},
			ID:              "1",
			wantStatus:      http.StatusCreated,
			wantContentType: "application/json",
			wantBody: `{"users":[{"links":{"self":"/cmp/v1/sources/1/users/match"},"name":"match"},{"links":{"self":"/cmp/v1/sources/1/users/skinhead"},"name":"skinhead"},{"links":{"self":"/cmp/v1/sources/1/users/3-d"},"name":"3-d"}],"name":"biffsgang","permissions":[],"links":{"self":"/cmp/v1/sources/1/roles/biffsgang"}}
`,
		},
	}
	for _, tt := range tests {
		h := &Service{
			Store: &mocks.Store{
				SourcesStore: tt.fields.SourcesStore,
			},
			TimeSeriesClient: tt.fields.TimeSeries,
			Logger:           tt.fields.Logger,
		}
		tt.args.r = tt.args.r.WithContext(httprouter.WithParams(
			context.Background(),
			httprouter.Params{
				{
					Key:   "id",
					Value: tt.ID,
				},
			}))

		h.NewSourceRole(tt.args.w, tt.args.r)

		resp := tt.args.w.Result()
		content := resp.Header.Get("Content-Type")
		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != tt.wantStatus {
			t.Errorf("%q. NewSourceRole() = %v, want %v", tt.name, resp.StatusCode, tt.wantStatus)
		}
		if tt.wantContentType != "" && content != tt.wantContentType {
			t.Errorf("%q. NewSourceRole() = %v, want %v", tt.name, content, tt.wantContentType)
		}
		if tt.wantBody != "" && string(body) != tt.wantBody {
			t.Errorf("%q. NewSourceRole() = \n***%v***\n,\nwant\n***%v***", tt.name, string(body), tt.wantBody)
		}
	}
}

func TestService_UpdateSourceRole(t *testing.T) {
	type fields struct {
		SourcesStore cmp.SourcesStore
		TimeSeries   TimeSeriesClient
		Logger       cmp.Logger
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		ID              string
		RoleID          string
		wantStatus      int
		wantContentType string
		wantBody        string
	}{
		{
			name: "Update role for data source",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"POST",
					"http://server.local/cmp/v1/sources/1/roles",
					ioutil.NopCloser(
						bytes.NewReader([]byte(`{"name": "biffsgang","users": [{"name": "match"},{"name": "skinhead"},{"name": "3-d"}]}`)))),
			},
			fields: fields{
				Logger: log.New(log.DebugLevel),
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, ID int) (cmp.Source, error) {
						return cmp.Source{
							ID:       1,
							Name:     "muh source",
							Username: "name",
							Password: "hunter2",
							URL:      "http://localhost:8086",
						}, nil
					},
				},
				TimeSeries: &mocks.TimeSeries{
					ConnectF: func(ctx context.Context, src *cmp.Source) error {
						return nil
					},
					RolesF: func(ctx context.Context) (cmp.RolesStore, error) {
						return &mocks.RolesStore{
							UpdateF: func(ctx context.Context, u *cmp.Role) error {
								return nil
							},
							GetF: func(ctx context.Context, name string) (*cmp.Role, error) {
								return &cmp.Role{
									Name: "biffsgang",
									Users: []cmp.User{
										{
											Name: "match",
										},
										{
											Name: "skinhead",
										},
										{
											Name: "3-d",
										},
									},
								}, nil
							},
						}, nil
					},
				},
			},
			ID:              "1",
			RoleID:          "biffsgang",
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
			wantBody: `{"users":[{"links":{"self":"/cmp/v1/sources/1/users/match"},"name":"match"},{"links":{"self":"/cmp/v1/sources/1/users/skinhead"},"name":"skinhead"},{"links":{"self":"/cmp/v1/sources/1/users/3-d"},"name":"3-d"}],"name":"biffsgang","permissions":[],"links":{"self":"/cmp/v1/sources/1/roles/biffsgang"}}
`,
		},
	}
	for _, tt := range tests {
		h := &Service{
			Store: &mocks.Store{
				SourcesStore: tt.fields.SourcesStore,
			},
			TimeSeriesClient: tt.fields.TimeSeries,
			Logger:           tt.fields.Logger,
		}

		tt.args.r = tt.args.r.WithContext(httprouter.WithParams(
			context.Background(),
			httprouter.Params{
				{
					Key:   "id",
					Value: tt.ID,
				},
				{
					Key:   "rid",
					Value: tt.RoleID,
				},
			}))

		h.UpdateSourceRole(tt.args.w, tt.args.r)

		resp := tt.args.w.Result()
		content := resp.Header.Get("Content-Type")
		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != tt.wantStatus {
			t.Errorf("%q. UpdateSourceRole() = %v, want %v", tt.name, resp.StatusCode, tt.wantStatus)
		}
		if tt.wantContentType != "" && content != tt.wantContentType {
			t.Errorf("%q. UpdateSourceRole() = %v, want %v", tt.name, content, tt.wantContentType)
		}
		if tt.wantBody != "" && string(body) != tt.wantBody {
			t.Errorf("%q. UpdateSourceRole() = \n***%v***\n,\nwant\n***%v***", tt.name, string(body), tt.wantBody)
		}
	}
}

func TestService_SourceRoleID(t *testing.T) {
	type fields struct {
		SourcesStore cmp.SourcesStore
		TimeSeries   TimeSeriesClient
		Logger       cmp.Logger
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		ID              string
		RoleID          string
		wantStatus      int
		wantContentType string
		wantBody        string
	}{
		{
			name: "Get role for data source",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"GET",
					"http://server.local/cmp/v1/sources/1/roles/biffsgang",
					nil),
			},
			fields: fields{
				Logger: log.New(log.DebugLevel),
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, ID int) (cmp.Source, error) {
						return cmp.Source{
							ID:       1,
							Name:     "muh source",
							Username: "name",
							Password: "hunter2",
							URL:      "http://localhost:8086",
						}, nil
					},
				},
				TimeSeries: &mocks.TimeSeries{
					ConnectF: func(ctx context.Context, src *cmp.Source) error {
						return nil
					},
					RolesF: func(ctx context.Context) (cmp.RolesStore, error) {
						return &mocks.RolesStore{
							GetF: func(ctx context.Context, name string) (*cmp.Role, error) {
								return &cmp.Role{
									Name: "biffsgang",
									Permissions: cmp.Permissions{
										{
											Name:  "grays_sports_almanac",
											Scope: "DBScope",
											Allowed: cmp.Allowances{
												"ReadData",
											},
										},
									},
									Users: []cmp.User{
										{
											Name: "match",
										},
										{
											Name: "skinhead",
										},
										{
											Name: "3-d",
										},
									},
								}, nil
							},
						}, nil
					},
				},
			},
			ID:              "1",
			RoleID:          "biffsgang",
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
			wantBody: `{"users":[{"links":{"self":"/cmp/v1/sources/1/users/match"},"name":"match"},{"links":{"self":"/cmp/v1/sources/1/users/skinhead"},"name":"skinhead"},{"links":{"self":"/cmp/v1/sources/1/users/3-d"},"name":"3-d"}],"name":"biffsgang","permissions":[{"scope":"DBScope","name":"grays_sports_almanac","allowed":["ReadData"]}],"links":{"self":"/cmp/v1/sources/1/roles/biffsgang"}}
`,
		},
	}
	for _, tt := range tests {
		h := &Service{
			Store: &mocks.Store{
				SourcesStore: tt.fields.SourcesStore,
			},
			TimeSeriesClient: tt.fields.TimeSeries,
			Logger:           tt.fields.Logger,
		}

		tt.args.r = tt.args.r.WithContext(httprouter.WithParams(
			context.Background(),
			httprouter.Params{
				{
					Key:   "id",
					Value: tt.ID,
				},
				{
					Key:   "rid",
					Value: tt.RoleID,
				},
			}))

		h.SourceRoleID(tt.args.w, tt.args.r)

		resp := tt.args.w.Result()
		content := resp.Header.Get("Content-Type")
		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != tt.wantStatus {
			t.Errorf("%q. SourceRoleID() = %v, want %v", tt.name, resp.StatusCode, tt.wantStatus)
		}
		if tt.wantContentType != "" && content != tt.wantContentType {
			t.Errorf("%q. SourceRoleID() = %v, want %v", tt.name, content, tt.wantContentType)
		}
		if tt.wantBody != "" && string(body) != tt.wantBody {
			t.Errorf("%q. SourceRoleID() = \n***%v***\n,\nwant\n***%v***", tt.name, string(body), tt.wantBody)
		}
	}
}

func TestService_RemoveSourceRole(t *testing.T) {
	type fields struct {
		SourcesStore cmp.SourcesStore
		TimeSeries   TimeSeriesClient
		Logger       cmp.Logger
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		ID         string
		RoleID     string
		wantStatus int
	}{
		{
			name: "remove role for data source",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"GET",
					"http://server.local/cmp/v1/sources/1/roles/biffsgang",
					nil),
			},
			fields: fields{
				Logger: log.New(log.DebugLevel),
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, ID int) (cmp.Source, error) {
						return cmp.Source{
							ID:       1,
							Name:     "muh source",
							Username: "name",
							Password: "hunter2",
							URL:      "http://localhost:8086",
						}, nil
					},
				},
				TimeSeries: &mocks.TimeSeries{
					ConnectF: func(ctx context.Context, src *cmp.Source) error {
						return nil
					},
					RolesF: func(ctx context.Context) (cmp.RolesStore, error) {
						return &mocks.RolesStore{
							DeleteF: func(context.Context, *cmp.Role) error {
								return nil
							},
						}, nil
					},
				},
			},
			ID:         "1",
			RoleID:     "biffsgang",
			wantStatus: http.StatusNoContent,
		},
	}
	for _, tt := range tests {
		h := &Service{
			Store: &mocks.Store{
				SourcesStore: tt.fields.SourcesStore,
			},
			TimeSeriesClient: tt.fields.TimeSeries,
			Logger:           tt.fields.Logger,
		}

		tt.args.r = tt.args.r.WithContext(httprouter.WithParams(
			context.Background(),
			httprouter.Params{
				{
					Key:   "id",
					Value: tt.ID,
				},
				{
					Key:   "rid",
					Value: tt.RoleID,
				},
			}))

		h.RemoveSourceRole(tt.args.w, tt.args.r)

		resp := tt.args.w.Result()
		if resp.StatusCode != tt.wantStatus {
			t.Errorf("%q. RemoveSourceRole() = %v, want %v", tt.name, resp.StatusCode, tt.wantStatus)
		}
	}
}

func TestService_SourceRoles(t *testing.T) {
	type fields struct {
		SourcesStore cmp.SourcesStore
		TimeSeries   TimeSeriesClient
		Logger       cmp.Logger
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		ID              string
		RoleID          string
		wantStatus      int
		wantContentType string
		wantBody        string
	}{
		{
			name: "Get roles for data source",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(
					"GET",
					"http://server.local/cmp/v1/sources/1/roles",
					nil),
			},
			fields: fields{
				Logger: log.New(log.DebugLevel),
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, ID int) (cmp.Source, error) {
						return cmp.Source{
							ID: 1,
						}, nil
					},
				},
				TimeSeries: &mocks.TimeSeries{
					ConnectF: func(ctx context.Context, src *cmp.Source) error {
						return nil
					},
					RolesF: func(ctx context.Context) (cmp.RolesStore, error) {
						return &mocks.RolesStore{
							AllF: func(ctx context.Context) ([]cmp.Role, error) {
								return []cmp.Role{
									cmp.Role{
										Name: "biffsgang",
										Permissions: cmp.Permissions{
											{
												Name:  "grays_sports_almanac",
												Scope: "DBScope",
												Allowed: cmp.Allowances{
													"ReadData",
												},
											},
										},
										Users: []cmp.User{
											{
												Name: "match",
											},
											{
												Name: "skinhead",
											},
											{
												Name: "3-d",
											},
										},
									},
								}, nil
							},
						}, nil
					},
				},
			},
			ID:              "1",
			RoleID:          "biffsgang",
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
			wantBody: `{"roles":[{"users":[{"links":{"self":"/cmp/v1/sources/1/users/match"},"name":"match"},{"links":{"self":"/cmp/v1/sources/1/users/skinhead"},"name":"skinhead"},{"links":{"self":"/cmp/v1/sources/1/users/3-d"},"name":"3-d"}],"name":"biffsgang","permissions":[{"scope":"DBScope","name":"grays_sports_almanac","allowed":["ReadData"]}],"links":{"self":"/cmp/v1/sources/1/roles/biffsgang"}}]}
`,
		},
	}
	for _, tt := range tests {
		h := &Service{
			Store: &mocks.Store{
				SourcesStore: tt.fields.SourcesStore,
			},
			TimeSeriesClient: tt.fields.TimeSeries,
			Logger:           tt.fields.Logger,
		}

		tt.args.r = tt.args.r.WithContext(httprouter.WithParams(
			context.Background(),
			httprouter.Params{
				{
					Key:   "id",
					Value: tt.ID,
				},
				{
					Key:   "rid",
					Value: tt.RoleID,
				},
			}))

		h.SourceRoles(tt.args.w, tt.args.r)

		resp := tt.args.w.Result()
		content := resp.Header.Get("Content-Type")
		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != tt.wantStatus {
			t.Errorf("%q. SourceRoles() = %v, want %v", tt.name, resp.StatusCode, tt.wantStatus)
		}
		if tt.wantContentType != "" && content != tt.wantContentType {
			t.Errorf("%q. SourceRoles() = %v, want %v", tt.name, content, tt.wantContentType)
		}
		if tt.wantBody != "" && string(body) != tt.wantBody {
			t.Errorf("%q. SourceRoles() = \n***%v***\n,\nwant\n***%v***", tt.name, string(body), tt.wantBody)
		}
	}
}
