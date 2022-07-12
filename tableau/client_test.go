package tableau

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestDo(t *testing.T) {
	tests := []struct {
		desc          string
		response      string
		statusCode    int
		method        string
		expectedError error
		wantHeaders   map[string]string
		body          interface{}
		v             interface{}
		want          interface{}
	}{
		{
			desc:       "returns an HTTP response and no error for 2xx responses",
			statusCode: http.StatusOK,
			response:   `{}`,
			method:     http.MethodGet,
			wantHeaders: map[string]string{
				"User-Agent": "go-tableau/v0.0.1",
			},
		},
		{
			desc:       "returns an HTTP response 204 when deleting a request",
			statusCode: http.StatusNoContent,
			method:     http.MethodDelete,
			response:   "",
			body:       nil,
			v:          &Project{},
			want:       nil,
		},
		{
			desc:       "returns an non-204 HTTP response when deleting a request",
			statusCode: http.StatusAccepted,
			method:     http.MethodDelete,
			response: `{
			"id": "test"
			}`,
			body: nil,
			v:    &DeleteProjectRequest{},
			want: &DeleteProjectRequest{
				ID: "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ctx := context.Background()
			c := qt.New(t)
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)

				if tt.wantHeaders != nil {
					for key, value := range tt.wantHeaders {
						c.Assert(r.Header.Get(key), qt.Equals, value)
					}
				}

				res := []byte(tt.response)
				if tt.response == "" {
					res = nil
				}
				_, err := w.Write(res)
				if err != nil {
					t.Fatal(err)
				}
			}))
			t.Cleanup(ts.Close)

			client, err := NewClient(ts.URL, "", "", "")
			if err != nil {
				t.Fatal(err)
			}

			req, err := client.newRequest(tt.method, "/api-endpoint", tt.body)
			if err != nil {
				t.Fatal(err)
			}

			res, err := client.client.Do(req)
			c.Assert(err, qt.IsNil)
			defer res.Body.Close()

			err = client.handleResponse(ctx, res, &tt.v)
			if err != nil {
				if tt.expectedError != nil {
					c.Assert(tt.expectedError.Error(), qt.Equals, err.Error())
				}
			}

			c.Assert(res, qt.Not(qt.IsNil))
			c.Assert(res.StatusCode, qt.Equals, tt.statusCode)

			if tt.v != nil && tt.want != nil {
				c.Assert(tt.want, qt.DeepEquals, tt.v)
			}
		})
	}
}
