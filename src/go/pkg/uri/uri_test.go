package uri

import (
	"reflect"
	"testing"
)

func TestURI_Addr(t *testing.T) {
	type fields struct {
		scheme   string
		host     string
		port     string
		resource string
		socket   string
		user     string
		password string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"Should return host:port",
			fields{host: "127.0.0.1", port: "8003"},
			"127.0.0.1:8003",
		},
		{
			"Should return socket",
			fields{host: "127.0.0.1", port: "8003", socket: "/var/lib/mysql/mysql.sock"},
			"/var/lib/mysql/mysql.sock",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &URI{
				scheme:   tt.fields.scheme,
				host:     tt.fields.host,
				port:     tt.fields.port,
				resource: tt.fields.resource,
				socket:   tt.fields.socket,
				user:     tt.fields.user,
				password: tt.fields.password,
			}
			if got := u.Addr(); got != tt.want {
				t.Errorf("Addr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURI_String(t *testing.T) {
	type fields struct {
		scheme   string
		host     string
		port     string
		resource string
		socket   string
		user     string
		password string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"Should return URI with creds",
			fields{scheme: "https", host: "127.0.0.1", port: "8003", user: "zabbix",
				password: "a35c2787-6ab4-4f6b-b538-0fcf91e678ed"},
			"https://zabbix:a35c2787-6ab4-4f6b-b538-0fcf91e678ed@127.0.0.1:8003",
		},
		{
			"Should return URI with username",
			fields{scheme: "https", host: "127.0.0.1", port: "8003", user: "zabbix"},
			"https://zabbix@127.0.0.1:8003",
		},
		{
			"Should return URI without creds",
			fields{scheme: "https", host: "127.0.0.1", port: "8003"},
			"https://127.0.0.1:8003",
		},
		{
			"Should return URI with path",
			fields{scheme: "oracle", host: "127.0.0.1", port: "1521", resource: "XE"},
			"oracle://127.0.0.1:1521/XE",
		},
		{
			"Should return URI without port",
			fields{scheme: "https", host: "127.0.0.1"},
			"https://127.0.0.1",
		},
		{
			"Should return URI socket",
			fields{scheme: "unix", socket: "/var/lib/mysql/mysql.sock"},
			"unix:///var/lib/mysql/mysql.sock",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &URI{
				scheme:   tt.fields.scheme,
				host:     tt.fields.host,
				port:     tt.fields.port,
				resource: tt.fields.resource,
				socket:   tt.fields.socket,
				user:     tt.fields.user,
				password: tt.fields.password,
			}
			if got := u.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	defaults              = &Defaults{Scheme: "https", Port: "443"}
	defaultsWithoutPort   = &Defaults{Scheme: "https"}
	defaultsWithoutScheme = &Defaults{Port: "443"}
	emptyDefaults         = &Defaults{}
	invalidDefaults       = &Defaults{Port: "99999"}
)

func TestNew(t *testing.T) {
	type args struct {
		rawuri   string
		defaults *Defaults
	}
	tests := []struct {
		name    string
		args    args
		wantRes *URI
		wantErr bool
	}{
		{
			"Parse URI with scheme and port, defaults are not set",
			args{"http://localhost:80", nil},
			&URI{scheme: "http", host: "localhost", port: "80"},
			false,
		},
		{
			"Parse URI without scheme and port, defaults are not set",
			args{"localhost", nil},
			&URI{scheme: "tcp", host: "localhost"},
			false,
		},
		{
			"Parse URI without scheme and port, defaults are empty",
			args{"localhost", emptyDefaults},
			&URI{scheme: "tcp", host: "localhost"},
			false,
		},
		{
			"Parse URI without scheme and port, defaults are fully set",
			args{"localhost", defaults},
			&URI{scheme: "https", host: "localhost", port: "443"},
			false,
		},
		{
			"Parse URI without scheme and port, defaults are partly set (only scheme)",
			args{"localhost", defaultsWithoutPort},
			&URI{scheme: "https", host: "localhost"},
			false,
		},
		{
			"Parse URI without scheme and port, defaults are partly set (only port)",
			args{"localhost", defaultsWithoutScheme},
			&URI{scheme: "tcp", host: "localhost", port: "443"},
			false,
		},
		{
			"Must fail if defaults are invalid",
			args{"localhost", invalidDefaults},
			nil,
			true,
		},
		{
			"Must fail if scheme is omitted",
			args{"://localhost", nil},
			nil,
			true,
		},
		{
			"Must fail if host is omitted",
			args{"tcp://:1521", nil},
			nil,
			true,
		},
		{
			"Must fail if port is greater than 65535",
			args{"tcp://localhost:65536", nil},
			nil,
			true,
		},
		{
			"Must fail if port is not integer",
			args{"tcp://:foo", nil},
			nil,
			true,
		},
		{
			"Should fail if URI is invalid",
			args{"!@#$%^&*()", nil},
			nil,
			true,
		},
		{
			"Parse URI with resource",
			args{"oracle://localhost:1521/XE", nil},
			&URI{scheme: "oracle", host: "localhost", port: "1521", resource: "XE"},
			false,
		},
		{
			"Parse URI with unix scheme",
			args{"unix:///var/run/memcached.sock", nil},
			&URI{scheme: "unix", socket: "/var/run/memcached.sock"},
			false,
		},
		{
			"Parse URI without unix scheme",
			args{"/var/run/memcached.sock", nil},
			&URI{scheme: "unix", socket: "/var/run/memcached.sock"},
			false,
		},
		{
			"Must fail if scheme is wrong",
			args{"tcp:///var/run/memcached.sock", nil},
			nil,
			true,
		},
		{
			"Must fail if socket is not specified",
			args{"unix://", nil},
			nil,
			true,
		},
		{
			"Parse URI with ipv6 address. Test 1",
			args{"tcp://[fe80::1ce7:d24a:97f0:3d83%25en0]:11211", nil},
			&URI{scheme: "tcp", host: "fe80::1ce7:d24a:97f0:3d83%en0", port: "11211"},
			false,
		},
		{
			"Parse URI with ipv6 address. Test 2",
			args{"tcp://[fe80::1ce7:d24a:97f0:3d83%en0]:11211", nil},
			&URI{scheme: "tcp", host: "fe80::1ce7:d24a:97f0:3d83%en0", port: "11211"},
			false,
		},
		{
			"Parse URI with ipv6 address. Test 3",
			args{"tcp://[fe80::1%25lo0]:11211", nil},
			&URI{scheme: "tcp", host: "fe80::1%lo0", port: "11211"},
			false,
		},
		{
			"Parse URI with ipv6 address. Test 4",
			args{"https://[::1]", defaults},
			&URI{scheme: "https", host: "::1", port: "443"},
			false,
		},
		{
			"Parse URI with ipv6 address. Test 5",
			args{"https://[::1]", nil},
			&URI{scheme: "https", host: "::1"},
			false,
		},
		{
			"Parse URI with ipv6 address. Test 6",
			args{"tcp://fe80::1:11211", nil},
			&URI{scheme: "tcp", host: "fe80::1", port: "11211"},
			false,
		},
		{
			"Parse URI with ipv6 address. Test 7",
			args{"tcp://::1:11289", nil},
			&URI{scheme: "tcp", host: "::1", port: "11289"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := New(tt.args.rawuri, tt.args.defaults)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("New() gotRes = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

var (
	uri              = "ssh://localhost:22"
	uriWithoutScheme = "localhost:22"
	uriOnlyHost      = "localhost"
)

func TestURIValidator_Validate(t *testing.T) {
	type fields struct {
		Defaults       *Defaults
		AllowedSchemes []string
	}
	type args struct {
		value *string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Validate uri with scheme in specified range",
			fields{nil, []string{"ssh"}},
			args{&uri},
			false,
		},
		{
			"Validate uri, scheme is not limited",
			fields{nil, nil},
			args{&uriWithoutScheme},
			false,
		},
		{
			"Must fail if scheme is out of range",
			fields{nil, []string{"ssh"}},
			args{&uriWithoutScheme},
			true,
		},
		{
			"Must fail if default scheme is out of range",
			fields{defaults, []string{"ssh"}},
			args{&uriOnlyHost},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := URIValidator{
				Defaults:       tt.fields.Defaults,
				AllowedSchemes: tt.fields.AllowedSchemes,
			}
			if err := v.Validate(tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
