package verification

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/fabric8-services/fabric8-common/log"
	"github.com/fabric8-services/fabric8-webhook/util"
	"github.com/goadesign/goa"
	goalogrus "github.com/goadesign/goa/logging/logrus"
)

var gs *goa.Service

func init() {
	gs = goa.New("fabric8-webhook-test")
	// record HTTP request metrics in prometh
	gs.WithLogger(goalogrus.New(log.Logger()))

}

func TestNew(t *testing.T) {
	type args struct {
		duration time.Duration
	}
	type fields struct {
		clientTransport util.RoundTripFunc
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        Service
		wantHookIPs []string
		wantErr     bool
	}{
		{
			name: "verification.New Test Positive",
			fields: fields{
				clientTransport: util.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
					// Test request parameters
					return &http.Response{
						StatusCode: 200,
						// Send response to be tested
						Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "hooks": [
    "192.30.252.0/22",
    "185.199.108.0/22",
    "140.82.112.0/20"
  ]
}`)),
					}, nil
				}),
			},
			args: args{15 * time.Minute},
			want: Service(&service{
				hookIPs: nil,
				Service: gs,
				ticker:  time.NewTicker(15 * time.Minute),
			}),
			wantHookIPs: []string{(&net.IPNet{IP: net.IPv4(192, 30, 252, 0), Mask: net.IPv4Mask(255, 255, 252, 0)}).String(), (&net.IPNet{IP: net.IPv4(185, 199, 108, 0), Mask: net.IPv4Mask(255, 255, 252, 0)}).String(), (&net.IPNet{IP: net.IPv4(140, 82, 112, 0), Mask: net.IPv4Mask(255, 255, 240, 0)}).String()},
			wantErr:     false,
		},
		{
			name: "verification.New Test Negative",
			fields: fields{
				clientTransport: util.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
					// Test request parameters
					return &http.Response{
						StatusCode: 400,
						// Send response to be tested
						Body: nil,
					}, errors.New("Mock Error Response")
				}),
			},
			args:        args{1 * time.Nanosecond},
			want:        nil,
			wantHookIPs: nil,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			util.SetMockNetClient(tt.fields.clientTransport)
			got, err := New(gs, tt.args.duration)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var hookIPs []string
			if err == nil {
				for _, ipnet := range got.(*service).hookIPs {
					hookIPs = append(hookIPs, ipnet.String())
				}
				// Changing s.HookIPs and ticker as reflect.Deepequal doesn't work for IPNet
				got.(*service).hookIPs = tt.want.(*service).hookIPs
				got.(*service).ticker = tt.want.(*service).ticker
				got.(*service).ticker.Stop()
			}

			if !reflect.DeepEqual(got, tt.want) ||
				!reflect.DeepEqual(hookIPs, tt.wantHookIPs) {
				t.Errorf("New() = %v, want %v, HookIPs %v, want %v",
					got, tt.want, hookIPs, tt.wantHookIPs)
			}
		})
	}
}

func Test_service_Verify(t *testing.T) {
	type fields struct {
		hooks []*net.IPNet
	}
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Verify Source Positive 1",
			fields: fields{
				hooks: []*net.IPNet{&net.IPNet{IP: net.IPv4(192, 30, 252, 0), Mask: net.IPv4Mask(255, 255, 252, 0)}, &net.IPNet{IP: net.IPv4(185, 199, 108, 0), Mask: net.IPv4Mask(255, 255, 252, 0)}, &net.IPNet{IP: net.IPv4(140, 82, 112, 0), Mask: net.IPv4Mask(255, 255, 240, 0)}},
			},
			args: args{
				&http.Request{
					Header: func() http.Header {
						h := http.Header{}
						h.Add("X-Forwarded-For", "192.30.252.1,92.30.252.3")
						return h
					}(),
				},
			},
			want: true,
		},
		{
			name: "Verify Source Positive 2",
			fields: fields{
				hooks: []*net.IPNet{&net.IPNet{IP: net.IPv4(192, 30, 252, 0), Mask: net.IPv4Mask(255, 255, 252, 0)}, &net.IPNet{IP: net.IPv4(185, 199, 108, 0), Mask: net.IPv4Mask(255, 255, 252, 0)}, &net.IPNet{IP: net.IPv4(140, 82, 112, 0), Mask: net.IPv4Mask(255, 255, 240, 0)}},
			},
			args: args{
				&http.Request{
					RemoteAddr: "192.30.252.1:8080",
				},
			},
			want: true,
		},
		{
			name: "Verify Source Negative 1",
			fields: fields{
				hooks: []*net.IPNet{&net.IPNet{IP: net.IPv4(192, 30, 252, 0), Mask: net.IPv4Mask(255, 255, 252, 0)}, &net.IPNet{IP: net.IPv4(185, 199, 108, 0), Mask: net.IPv4Mask(255, 255, 252, 0)}, &net.IPNet{IP: net.IPv4(140, 82, 112, 0), Mask: net.IPv4Mask(255, 255, 240, 0)}},
			},
			args: args{
				&http.Request{
					RemoteAddr: "92.30.252.0:8080",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				hookIPs: tt.fields.hooks,
				Service: gs,
			}
			if got := s.Verify(tt.args.req); got != tt.want {
				t.Errorf("service.Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_setHooks(t *testing.T) {
	type fields struct {
		clientTransport util.RoundTripFunc
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			name: "setHookIPs Verification Test Positive",
			fields: fields{
				clientTransport: util.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
					// Test request parameters
					return &http.Response{
						StatusCode: 200,
						// Send response to be tested
						Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "verifiable_password_authentication": true,
  "github_services_sha": "2f2313161ed4f940a57ae3f0936eb8e9695bb8a8",
  "hooks": [
    "192.30.252.0/22",
    "185.199.108.0/22",
    "140.82.112.0/20"
  ],
  "git": [
    "192.30.252.0/22",
    "185.199.108.0/22",
    "140.82.112.0/20",
    "13.229.188.59/32",
    "13.250.177.223/32",
    "18.194.104.89/32",
    "18.195.85.27/32",
    "35.159.8.160/32",
    "52.74.223.119/32"
  ],
  "pages": [
    "192.30.252.153/32",
    "192.30.252.154/32",
    "185.199.108.153/32",
    "185.199.109.153/32",
    "185.199.110.153/32",
    "185.199.111.153/32"
  ],
  "importer": [
    "54.87.5.173",
    "54.166.52.62",
    "23.20.92.3"
  ]
}`)),
					}, nil
				}),
			},
			want:    []string{(&net.IPNet{IP: net.IPv4(192, 30, 252, 0), Mask: net.IPv4Mask(255, 255, 252, 0)}).String(), (&net.IPNet{IP: net.IPv4(185, 199, 108, 0), Mask: net.IPv4Mask(255, 255, 252, 0)}).String(), (&net.IPNet{IP: net.IPv4(140, 82, 112, 0), Mask: net.IPv4Mask(255, 255, 240, 0)}).String()},
			wantErr: false,
		},
		{
			name: "setHookIPs Verification Test Negative - 1 Response Err",
			fields: fields{
				clientTransport: util.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
					// Test request parameters
					return &http.Response{
						StatusCode: 400,
						// Send response to be tested
						Body: nil,
					}, errors.New("Mock Error Response")
				}),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "setHookIPs Verification Test Negative - 2 Body Error",
			fields: fields{
				clientTransport: util.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
					// Test request parameters
					return &http.Response{
						StatusCode: 200,
						// Send response to be tested
						Body: util.ErrReader("Mock Error"),
					}, nil
				}),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "setHookIPs Verification Test Negative - 3 Unmarshal Error",
			fields: fields{
				clientTransport: util.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
					// Test request parameters
					return &http.Response{
						StatusCode: 200,
						// Send response to be tested
						Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "hooks": "192.30.252.0/22"
}`)),
					}, nil
				}),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "setHookIPs Verification Test Negative - 4 IPNet parsing",
			fields: fields{
				clientTransport: util.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
					// Test request parameters
					return &http.Response{
						StatusCode: 200,
						// Send response to be tested
						Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "hooks": [
    "192.30.252.0/22",
    "185.1979.108.777/722",
    "140.82.112.0/20"
  ]
}`)),
					}, nil
				}),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				Service: gs,
			}
			util.SetMockNetClient(tt.fields.clientTransport)
			err := s.setHookIPs()
			if (err != nil) != tt.wantErr {
				t.Errorf("service.setHooks() error = %v, wantErr %v", err, tt.wantErr)
			}
			var hookIPs []string
			if err == nil {
				for _, ipnet := range s.hookIPs {
					hookIPs = append(hookIPs, ipnet.String())
				}
			}
			if !reflect.DeepEqual(hookIPs, tt.want) {
				t.Errorf("New() = %v, want %v,",
					hookIPs, tt.want)
			}

		})
	}
}

func Test_service_isGithubIP(t *testing.T) {
	type fields struct {
		hooks   []*net.IPNet
		Service *goa.Service
	}
	type args struct {
		i string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "isGithubIP test positive",
			fields: fields{
				hooks: []*net.IPNet{&net.IPNet{IP: net.IPv4(135, 104, 0, 0), Mask: net.IPv4Mask(255, 255, 255, 255)}, &net.IPNet{IP: net.IPv4(185, 199, 108, 0), Mask: net.IPv4Mask(255, 255, 252, 0)}, &net.IPNet{IP: net.IPv4(140, 82, 112, 0), Mask: net.IPv4Mask(255, 255, 240, 0)}},
			},
			args: args{i: "185.199.108.17"},
			want: true,
		},
		{
			name: "isGithubIP test negative",
			fields: fields{
				hooks: []*net.IPNet{&net.IPNet{IP: net.IPv4(135, 104, 0, 0), Mask: net.IPv4Mask(255, 255, 255, 255)}, &net.IPNet{IP: net.IPv4(185, 199, 108, 0), Mask: net.IPv4Mask(255, 255, 252, 0)}, &net.IPNet{IP: net.IPv4(140, 82, 112, 0), Mask: net.IPv4Mask(255, 255, 240, 0)}},
			},
			args: args{i: "25.199.108.17"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				hookIPs: tt.fields.hooks,
				Service: tt.fields.Service,
			}
			if got := s.isGithubIP(tt.args.i); got != tt.want {
				t.Errorf("service.isGithubIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
