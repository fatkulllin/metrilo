package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fatkulllin/metrilo/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusHandler(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}
	type request struct {
		method      string
		contentType string
		requestURI  string
	}
	tests := []struct {
		name    string
		want    want
		request request
	}{
		{
			name: "positive test counter #1",
			want: want{
				code:        200,
				contentType: "text/plain",
			},
			request: request{
				requestURI:  "/update/counter/a/10",
				contentType: "text/plain",
				method:      http.MethodPost,
			},
		},
		{
			name: "positive test gauge #2",
			want: want{
				code:        200,
				contentType: "text/plain",
			},
			request: request{
				requestURI:  "/update/gauge/a/10",
				contentType: "text/plain",
				method:      http.MethodPost,
			},
		},
		{
			name: "Method not allowed - wrong HTTP method #3",
			want: want{
				code:        405,
				contentType: "text/plain; charset=utf-8",
			},
			request: request{
				requestURI:  "/update/gauge/a/10",
				contentType: "text/plain",
				method:      http.MethodGet,
			},
		},
		{
			name: "Method not allowed - wrong Content-Type #4",
			want: want{
				code:        405,
				contentType: "text/plain; charset=utf-8",
			},
			request: request{
				requestURI:  "/update/gauge/a/10",
				contentType: "application/json",
				method:      http.MethodPost,
			},
		},
		{
			name: "Bad Request - wrong type metric #5",
			want: want{
				code:        400,
				contentType: "text/plain",
			},
			request: request{
				requestURI:  "/update/12a/a/10",
				contentType: "text/plain",
				method:      http.MethodPost,
			},
		},
		{
			name: "Not found typeMetric test #6",
			want: want{
				code:        404,
				contentType: "text/plain; charset=utf-8",
			},
			request: request{
				requestURI:  "/update/",
				contentType: "text/plain",
				method:      http.MethodPost,
			},
		},
		{
			name: "Not found nameMetric test #7",
			want: want{
				code:        404,
				contentType: "text/plain; charset=utf-8",
			},
			request: request{
				requestURI:  "/update/counter",
				contentType: "text/plain",
				method:      http.MethodPost,
			},
		},
		{
			name: "Not found valueMetric test #8",
			want: want{
				code:        404,
				contentType: "text/plain; charset=utf-8",
			},
			request: request{
				requestURI:  "/update/counter/a",
				contentType: "text/plain; charset=utf-8",
				method:      http.MethodPost,
			},
		},
		{
			name: "Bad request - wrong counter value metric #9",
			want: want{
				code:        400,
				contentType: "text/plain",
			},
			request: request{
				requestURI:  "/update/counter/a/aaa",
				contentType: "text/plain",
				method:      http.MethodPost,
			},
		},
		{
			name: "Bad request - wrong counter value metric #10",
			want: want{
				code:        400,
				contentType: "text/plain",
			},
			request: request{
				requestURI:  "/update/counter/a/1a",
				contentType: "text/plain",
				method:      http.MethodPost,
			},
		},
		{
			name: "Bad request - wrong gauge value metric #11",
			want: want{
				code:        400,
				contentType: "text/plain",
			},
			request: request{
				requestURI:  "/update/gauge/a/aaa",
				contentType: "text/plain",
				method:      http.MethodPost,
			},
		},
		{
			name: "Bad request - wrong gauge value metric #12",
			want: want{
				code:        400,
				contentType: "text/plain",
			},
			request: request{
				requestURI:  "/update/gauge/a/1a",
				contentType: "text/plain",
				method:      http.MethodPost,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			server := server.NewServer()
			mux := server.Start()

			httpMethod := test.request.method
			requestURI := test.request.requestURI
			contentType := test.request.contentType

			request := httptest.NewRequest(httpMethod, requestURI, nil)
			request.Header.Add("Content-Type", contentType)
			// создаём новый Recorder
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			_, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
