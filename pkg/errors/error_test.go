package errors

import (
	"net/http"
	"reflect"
	"testing"
)

func TestSvrError_Error(t *testing.T) {
	type fields struct {
		code int
		info string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"unittestError",
			fields{
				http.StatusOK,
				"",
			},
			"",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			e := &SvrError{
				code: tt.fields.code,
				info: tt.fields.info,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("SvrError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSvrError_Code(t *testing.T) {
	type fields struct {
		code int
		info string
	}

	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			"unitTestCode",
			fields{
				http.StatusOK,
				"",
			},
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			e := &SvrError{
				code: tt.fields.code,
				info: tt.fields.info,
			}
			if got := e.Code(); got != tt.want {
				t.Errorf("SvrError.Code() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewWithCode(t *testing.T) {
	type args struct {
		code int
		err  error
	}

	tests := []struct {
		name string
		args args
		want CodeError
	}{
		{
			"unitTestNewWithCodeFirst",
			args{
				http.StatusOK,
				&SvrError{},
			},
			&SvrError{},
		},
		{
			"unitTestNewWithCodeSecond",
			args{
				http.StatusOK,
				New("test second return"),
			},
			&SvrError{
				http.StatusOK,
				"test second return",
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			if got := NewWithCode(tt.args.code, tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewWithCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewWithInfo(t *testing.T) {
	type args struct {
		code int
		info string
	}

	tests := []struct {
		name string
		args args
		want CodeError
	}{
		{
			"unitTestNewWithInfo",
			args{},
			&SvrError{},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			if got := NewWithInfo(tt.args.code, tt.args.info); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewWithInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
